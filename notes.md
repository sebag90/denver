#### 1. Engine Detection

When connecting to the socket, inspect the engine via GET /v{apiVersion}/version or GET /v{apiVersion}/info:

- Podman returns "Components" with names like "Podman Engine" (or headers like Libpod-API-Version).
- You can also check if you are connected to a rootless socket ($XDG_RUNTIME_DIR/...).

#### 2. Conditional createContainer Payload

```go
func createContainer(client *http.Client, isPodman bool, isRootless bool) {
   hostConfig := map[string]any{
       "PortBindings": map[string]any{...},
   }

   payload := map[string]any{
       "Image":      imageName,
       "HostConfig": hostConfig,
       // ...
   }

   if isPodman && isRootless {
       // Podman rootless requires keep-id to map host UID -> container UID
       hostConfig["UsernsMode"] = "keep-id"
   } else {
       // Standard Docker (and rootful Podman):
       // Run container directly as the host user's UID:GID
       payload["User"] = fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
   }
   // ...
}
```

#### 3. Exposing the Engine Socket

The container gets the daemon socket dv itself talks to (`SocketPath()`), mounted
at the path the matching CLI expects:

- Podman: `<socket>:/run/podman/podman.sock`
- Docker: `<socket>:/var/run/docker.sock`

with `CONTAINER_HOST` and `DOCKER_HOST` both pointing at it (the Podman CLI reads
only the former, the Docker CLI only the latter). Nothing is mounted when the
socket does not exist, so creating a container never fails because the daemon is
unreachable. Containers started from inside the project container run on the host
daemon, next to the project container itself (docker-outside-of-docker).

The socket is mounted without `:Z` — relabeling it would relabel the host's
runtime socket and break the host session — and it keeps the host's SELinux label
(`user_tmp_t` under `$XDG_RUNTIME_DIR`). A container running as `container_t`
with its own MCS category pair is not allowed to `connect()` to it, so any
container that gets a socket bind (daemon socket or Wayland) is created with
SELinux labeling disabled:

- libpod: `"selinux_opts": ["disable"]`
- Docker-compatible API: `HostConfig.SecurityOpt: ["label=disable"]`

This is the same trade-off the original devcontainer shell made with
`--security-opt label=disable`. Plain containers without a socket bind keep their
confinement, and their file mounts are still relabeled with `:Z`.

Mounting the socket is not enough on its own: the container user must be the
socket's owner on the host. Rootless Podman gets that from `keep-id`; a rootful
daemon whose socket is group-owned would need the user in that group too.
