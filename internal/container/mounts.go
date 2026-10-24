package container

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// containerHome is the home directory of the unprivileged user in the dev
// image. Host dotfiles are mounted below it so tools inside the container find
// them where they expect them.
const containerHome = "/home/dev"

// The daemon socket is mounted at the path the CLI expects for the engine in
// use, so that "podman" and "docker" work inside the container without extra
// configuration.
const (
	podmanSocketPath = "/run/podman/podman.sock"
	dockerSocketPath = "/var/run/docker.sock"
)

// hostHomePaths are paths relative to the host's $HOME that are made available
// in the container when they exist.
var hostHomePaths = []string{
	".pi/agent",
	".aws",
	".gitconfig",
	".ssh",
	".netrc",
}

// Bind is a bind mount. Options holds mount options such as "Z" that are
// specific to the mount; each API renderer adds its own defaults.
type Bind struct {
	Source      string
	Destination string
	Options     []string
	// Socket marks a bind of a host socket (the daemon socket, Wayland). Such
	// a mount keeps the host's SELinux label, so the container has to run with
	// SELinux labeling disabled to be allowed to use it.
	Socket bool
}

// unconfined reports whether the binds include a host socket, and therefore
// whether the container needs SELinux labeling disabled. Relabeling a socket
// belonging to the host session is not an option, and the container's own
// label (container_t with a per-container category pair) is not allowed to
// connect to it.
func unconfined(binds []Bind) bool {
	return slices.ContainsFunc(binds, func(b Bind) bool { return b.Socket })
}

// dockerBind renders the bind for the Docker-compatible API.
func (b Bind) dockerBind() string {
	bind := b.Source + ":" + b.Destination
	if len(b.Options) > 0 {
		bind += ":" + strings.Join(b.Options, ",")
	}
	return bind
}

// libpodMount renders the bind for Podman's libpod API.
func (b Bind) libpodMount() map[string]any {
	return map[string]any{
		"destination": b.Destination,
		"type":        "bind",
		"source":      b.Source,
		"options":     append([]string{"rbind", "rw"}, b.Options...),
	}
}

// hostMounts returns the bind mounts and environment variables that expose the
// host's dotfiles and Wayland socket to the container.
//
// Bind mounts for host files are relabeled with :Z so that SELinux allows the
// container to read them. The Wayland socket is mounted without a label: it is
// a socket, and relabeling the user's runtime directory would break the host
// session. Mounting a socket makes the container run with SELinux labeling
// disabled instead (see unconfined).
func hostMounts() (binds []Bind, env []string) {
	if home, err := os.UserHomeDir(); err == nil {
		for _, rel := range hostHomePaths {
			source := filepath.Join(home, rel)
			if _, err := os.Stat(source); err != nil {
				continue
			}
			target := filepath.Join(containerHome, rel)
			binds = append(binds, Bind{Source: source, Destination: target, Options: []string{"Z"}})
		}
	}

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	display := os.Getenv("WAYLAND_DISPLAY")
	if runtimeDir != "" && display != "" {
		socket := filepath.Join(runtimeDir, display)
		if info, err := os.Stat(socket); err == nil && info.Mode()&os.ModeSocket != 0 {
			binds = append(binds, Bind{Source: socket, Destination: "/tmp/wayland-0", Socket: true})
			env = append(env, "WAYLAND_DISPLAY=/tmp/wayland-0")
		}
	}

	return binds, env
}

// engineMounts exposes the daemon socket dv is talking to inside the container,
// together with the environment variables that point the CLIs at it. This makes
// the container a "docker-outside-of-docker" style sandbox: containers started
// from inside it run on the host daemon, next to the project container itself.
//
// Both CONTAINER_HOST and DOCKER_HOST are set, because the Podman CLI only reads
// the former and the Docker CLI only reads the latter.
//
// The socket cannot be relabeled for the container, so mounting it also makes
// the container run with SELinux labeling disabled (see unconfined). The
// container user still has to be the socket's owner on the host: rootless
// Podman gets that from keep-id (see createPodmanContainer), while a rootful
// daemon whose socket is group-owned would need the container user in that
// group as well.
//
// When no socket is listening (for example when dv has not been pointed at a
// daemon yet), nothing is mounted and the container simply has no engine access.
func engineMounts(socket string, engine Engine) (binds []Bind, env []string) {
	info, err := os.Stat(socket)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return nil, nil
	}

	destination := dockerSocketPath
	if engine.Podman {
		destination = podmanSocketPath
	}

	// The socket is mounted without a label: relabeling it with :Z would
	// relabel the host's runtime socket and break the host session.
	binds = append(binds, Bind{Source: socket, Destination: destination, Socket: true})
	env = append(env,
		"CONTAINER_HOST=unix://"+destination,
		"DOCKER_HOST=unix://"+destination,
	)
	return binds, env
}
