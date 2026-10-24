package container

import (
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestHostMountsIncludesExistingPathsAndWayland(t *testing.T) {
	home := t.TempDir()
	runtimeDir := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")

	for _, rel := range []string{".pi/agent", ".ssh", ".gitconfig", ".netrc"} {
		path := filepath.Join(home, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	socket := filepath.Join(runtimeDir, "wayland-0")
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	bindList, env := hostMounts()
	binds := make([]string, len(bindList))
	for i, bind := range bindList {
		binds[i] = bind.dockerBind()
	}

	for _, rel := range []string{".pi/agent", ".ssh", ".gitconfig", ".netrc"} {
		want := filepath.Join(home, rel) + ":" + containerHome + "/" + rel + ":Z"
		if !slices.Contains(binds, want) {
			t.Errorf("missing bind %q in %v", want, binds)
		}
	}
	if slices.ContainsFunc(binds, func(bind string) bool {
		return strings.Contains(bind, "/.aws:")
	}) {
		t.Errorf("mounted .aws even though it does not exist: %v", binds)
	}

	if want := socket + ":/tmp/wayland-0"; !slices.Contains(binds, want) {
		t.Errorf("missing Wayland bind %q in %v", want, binds)
	}
	if want := "WAYLAND_DISPLAY=/tmp/wayland-0"; !slices.Contains(env, want) {
		t.Errorf("missing Wayland env %q in %v", want, env)
	}
	if !unconfined(bindList) {
		t.Error("unconfined() = false, want true with a Wayland socket bind")
	}
}

// listenSocket creates a unix socket at a temporary path.
func listenSocket(t *testing.T, name string) string {
	t.Helper()
	socket := filepath.Join(t.TempDir(), name)
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	return socket
}

// engineEnv asserts that env points both CLIs at the given socket path.
func engineEnv(t *testing.T, env []string, socketPath string) {
	t.Helper()
	for _, want := range []string{
		"CONTAINER_HOST=unix://" + socketPath,
		"DOCKER_HOST=unix://" + socketPath,
	} {
		if !slices.Contains(env, want) {
			t.Errorf("missing env %q in %v", want, env)
		}
	}
}

func TestEngineMountsExposesPodmanSocket(t *testing.T) {
	socket := listenSocket(t, "podman.sock")

	binds, env := engineMounts(socket, Engine{Podman: true, Rootless: true})
	if len(binds) != 1 {
		t.Fatalf("binds = %v, want one", binds)
	}
	if binds[0].Source != socket || binds[0].Destination != podmanSocketPath {
		t.Errorf("bind = %+v, want %q at %q", binds[0], socket, podmanSocketPath)
	}
	if len(binds[0].Options) != 0 {
		t.Errorf("bind options = %v, want none (relabeling the socket would break the host)", binds[0].Options)
	}
	engineEnv(t, env, podmanSocketPath)
	if !unconfined(binds) {
		t.Error("unconfined() = false, want true for a socket bind")
	}
}

func TestEngineMountsExposesDockerSocket(t *testing.T) {
	socket := listenSocket(t, "docker.sock")

	binds, env := engineMounts(socket, Engine{})
	if len(binds) != 1 {
		t.Fatalf("binds = %v, want one", binds)
	}
	if binds[0].Source != socket || binds[0].Destination != dockerSocketPath {
		t.Errorf("bind = %+v, want %q at %q", binds[0], socket, dockerSocketPath)
	}
	engineEnv(t, env, dockerSocketPath)
}

func TestUnconfined(t *testing.T) {
	socketBind := func(source string) []Bind {
		return []Bind{{Source: source, Destination: "/tmp/x", Socket: true}}
	}
	tests := []struct {
		name  string
		binds []Bind
		want  bool
	}{
		{"no binds", nil, false},
		{"file binds only", []Bind{workspaceBind("/src")}, false},
		{"socket bind", socketBind("/run/user/1000/podman/podman.sock"), true},
		{"socket among files", append([]Bind{workspaceBind("/src")}, socketBind("/tmp/wayland-0")...), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := unconfined(tc.binds); got != tc.want {
				t.Errorf("unconfined(%v) = %t, want %t", tc.binds, got, tc.want)
			}
		})
	}
}

func TestEngineMountsSkipsWithoutDaemonSocket(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.sock")
	if binds, env := engineMounts(missing, Engine{Podman: true}); binds != nil || env != nil {
		t.Errorf("engineMounts(missing) = %v, %v, want nil", binds, env)
	}

	regular := filepath.Join(t.TempDir(), "regular-file")
	if err := os.WriteFile(regular, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if binds, env := engineMounts(regular, Engine{Podman: true}); binds != nil || env != nil {
		t.Errorf("engineMounts(non-socket) = %v, %v, want nil", binds, env)
	}
}

func TestHostMountsSkipsWaylandWithoutSocket(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")

	bindList, env := hostMounts()
	binds := make([]string, len(bindList))
	for i, bind := range bindList {
		binds[i] = bind.dockerBind()
	}
	if len(env) != 0 {
		t.Errorf("unexpected env: %v", env)
	}
	if slices.ContainsFunc(binds, func(bind string) bool {
		return strings.Contains(bind, "wayland-0")
	}) {
		t.Errorf("unexpected Wayland bind: %v", binds)
	}
}
