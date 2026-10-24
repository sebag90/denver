package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string, req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}
}

// createClient answers engine detection and captures the create payload.
func createClient(t *testing.T, podman bool, createPath string, payload *map[string]any) *http.Client {
	t.Helper()
	components := `[]`
	if podman {
		components = `[{"Name":"Podman Engine"}]`
	}
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/version"):
			return jsonResponse(http.StatusOK, `{"Components":`+components+`}`, r), nil
		case strings.HasSuffix(r.URL.Path, "/info"):
			return jsonResponse(http.StatusOK, `{"SecurityOptions":[]}`, r), nil
		case strings.HasSuffix(r.URL.Path, createPath) && r.Method == http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
				t.Errorf("decode create payload: %v", err)
			}
			return jsonResponse(http.StatusCreated, `{"Id":"deadbeef"}`, r), nil
		default:
			return nil, fmt.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})}
}

// noDaemon points SocketPath at a socket that does not exist, so that the
// create payloads stay the same on machines that do have a daemon running.
func noDaemon(t *testing.T) {
	t.Helper()
	t.Setenv("CONTAINER_HOST", filepath.Join(t.TempDir(), "missing.sock"))
	t.Setenv("DOCKER_HOST", "")
}

// daemonSocket creates a listening socket and points SocketPath at it.
func daemonSocket(t *testing.T) string {
	t.Helper()
	socket := listenSocket(t, "podman.sock")
	t.Setenv("CONTAINER_HOST", "unix://"+socket)
	t.Setenv("DOCKER_HOST", "")
	return socket
}

func TestCreateDockerContainerBindsWorkspaceByDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	noDaemon(t)

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if _, err := CreateContainer(createClient(t, false, "/containers/create", &payload), "image", false); err != nil {
		t.Fatal(err)
	}

	if got := payload["WorkingDir"]; got != dir {
		t.Errorf("WorkingDir = %v, want %q", got, dir)
	}
	labels := payload["Labels"].(map[string]any)
	if labels[sandboxLabel] != "false" {
		t.Errorf("sandbox label = %v, want false", labels[sandboxLabel])
	}
	hostConfig := payload["HostConfig"].(map[string]any)
	binds := hostConfig["Binds"].([]any)
	want := workspaceBind(dir).dockerBind()
	if len(binds) != 1 || binds[0] != want {
		t.Errorf("Binds = %v, want [%q]", binds, want)
	}
	if _, ok := hostConfig["SecurityOpt"]; ok {
		t.Errorf("SecurityOpt = %v, want it omitted without a socket bind", hostConfig["SecurityOpt"])
	}
}

func TestCreateDockerContainerSandboxSkipsWorkspaceBind(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	noDaemon(t)

	var payload map[string]any
	if _, err := CreateContainer(createClient(t, false, "/containers/create", &payload), "image", true); err != nil {
		t.Fatal(err)
	}

	if got, ok := payload["WorkingDir"]; !ok || got != "" {
		t.Errorf("WorkingDir = %v, want empty", got)
	}
	labels := payload["Labels"].(map[string]any)
	if labels[sandboxLabel] != "true" {
		t.Errorf("sandbox label = %v, want true", labels[sandboxLabel])
	}
	if binds := payload["HostConfig"].(map[string]any)["Binds"].([]any); len(binds) != 0 {
		t.Errorf("Binds = %v, want none", binds)
	}
}

func TestCreatePodmanContainerSandboxOmitsWorkDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	noDaemon(t)

	var payload map[string]any
	if _, err := CreateContainer(createClient(t, true, "/libpod/containers/create", &payload), "image", true); err != nil {
		t.Fatal(err)
	}

	if got, ok := payload["work_dir"]; ok {
		t.Errorf("work_dir = %v, want it omitted", got)
	}
	if got, ok := payload["selinux_opts"]; ok {
		t.Errorf("selinux_opts = %v, want it omitted without a socket bind", got)
	}
	if mounts := payload["mounts"].([]any); len(mounts) != 0 {
		t.Errorf("mounts = %v, want none", mounts)
	}
}

func TestCreateDockerContainerMountsDaemonSocket(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	socket := daemonSocket(t)

	var payload map[string]any
	if _, err := CreateContainer(createClient(t, false, "/containers/create", &payload), "image", false); err != nil {
		t.Fatal(err)
	}

	hostConfig := payload["HostConfig"].(map[string]any)
	want := Bind{Source: socket, Destination: dockerSocketPath}.dockerBind()
	binds := hostConfig["Binds"].([]any)
	if !slices.ContainsFunc(binds, func(b any) bool { return b == want }) {
		t.Errorf("Binds = %v, want it to contain %q", binds, want)
	}

	secOpt := hostConfig["SecurityOpt"].([]any)
	if len(secOpt) != 1 || secOpt[0] != "label=disable" {
		t.Errorf("SecurityOpt = %v, want [label=disable] for a socket bind", secOpt)
	}

	env := payload["Env"].([]any)
	for _, want := range []string{"CONTAINER_HOST", "DOCKER_HOST"} {
		if !slices.ContainsFunc(env, func(e any) bool { return e == want+"=unix://"+dockerSocketPath }) {
			t.Errorf("Env = %v, want it to point %s at the socket", env, want)
		}
	}
}

func TestCreatePodmanContainerMountsDaemonSocket(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	socket := daemonSocket(t)

	var payload map[string]any
	if _, err := CreateContainer(createClient(t, true, "/libpod/containers/create", &payload), "image", false); err != nil {
		t.Fatal(err)
	}

	mounts := payload["mounts"].([]any)
	if !slices.ContainsFunc(mounts, func(m any) bool {
		mount := m.(map[string]any)
		return mount["source"] == socket && mount["destination"] == podmanSocketPath
	}) {
		t.Errorf("mounts = %v, want the daemon socket at %q", mounts, podmanSocketPath)
	}

	env := payload["env"].(map[string]any)
	if got := env["CONTAINER_HOST"]; got != "unix://"+podmanSocketPath {
		t.Errorf("CONTAINER_HOST = %v, want %q", got, "unix://"+podmanSocketPath)
	}

	selinuxOpts := payload["selinux_opts"].([]any)
	if len(selinuxOpts) != 1 || selinuxOpts[0] != "disable" {
		t.Errorf("selinux_opts = %v, want [disable] for a socket bind", selinuxOpts)
	}
}

func TestCheckSandbox(t *testing.T) {
	tests := []struct {
		name    string
		labels  map[string]string
		sandbox bool
		wantErr bool
	}{
		{"legacy container is not sandboxed", nil, false, false},
		{"legacy container cannot satisfy sandbox", nil, true, true},
		{"matching sandbox", map[string]string{sandboxLabel: "true"}, true, false},
		{"matching normal", map[string]string{sandboxLabel: "false"}, false, false},
		{"normal container cannot satisfy sandbox", map[string]string{sandboxLabel: "false"}, true, true},
		{"sandbox container cannot satisfy normal", map[string]string{sandboxLabel: "true"}, false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkSandbox(&Container{Names: []string{"/denver.1"}, Labels: tc.labels}, tc.sandbox)
			if (err != nil) != tc.wantErr {
				t.Errorf("checkSandbox() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
