package container

import (
	"fmt"
	"os"
	"path/filepath"
)

func getCurrentPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

var commonEnv = []string{
	"COLORTERM=truecolor",
	"CONTAINER_HOST=unix:///run/podman/podman.sock",
}

var hostConfig = map[string]any{
	"NetworkMode": "host",
	"Binds": []string{
		fmt.Sprintf("%s:%s:Z", getCurrentPath(), getCurrentPath()),
	},
}
