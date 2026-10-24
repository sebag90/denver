package container

// commonEnv is set for every container, whatever the engine.
var commonEnv = []string{
	"COLORTERM=truecolor",
}

// workspaceBind mounts the current directory at the same path inside the
// container, so the container's working directory matches the host's.
func workspaceBind(path string) Bind {
	return Bind{Source: path, Destination: path, Options: []string{"Z"}}
}

func newHostConfig(binds []string) map[string]any {
	return map[string]any{
		"NetworkMode": "host",
		"Init":        true,
		"Binds":       binds,
	}
}
