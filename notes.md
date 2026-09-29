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
