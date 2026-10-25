package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

// ErrContainerExists is returned by CreateContainer when a container with the
// derived name already exists.
var ErrContainerExists = errors.New("container already exists")

const (
	projectLabel   = "denver.project"
	sandboxLabel   = "denver.sandbox"
	directoryLabel = "denver.directory"
)

// CreateContainer creates the per-directory container and returns its ID.
//
// Podman is created through its libpod API so the container can get a passwd
// entry for the "dev" user. Docker uses the compatible API instead, where no
// such option exists.
func CreateContainer(client *http.Client, imageName string, sandbox bool) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determine working directory: %w", err)
	}

	engine, err := DetectEngine(client)
	if err != nil {
		return "", fmt.Errorf("detect container engine: %w", err)
	}

	mountBinds, mountEnv := hostMounts()
	engineBinds, engineEnv := engineMounts(SocketPath(), engine)
	binds := slices.Concat(mountBinds, engineBinds)
	workDir := ""
	if !sandbox {
		binds = append([]Bind{workspaceBind(dir)}, binds...)
		workDir = dir
	}
	env := slices.Concat(commonEnv, mountEnv, engineEnv)
	name := ContainerName()
	labels := map[string]string{
		projectLabel:   name,
		sandboxLabel:   fmt.Sprintf("%t", sandbox),
		directoryLabel: dir,
	}

	if engine.Podman {
		return createPodmanContainer(client, name, imageName, workDir, env, labels, binds, engine)
	}
	return createDockerContainer(client, name, imageName, workDir, env, labels, binds)
}

func createPodmanContainer(
	client *http.Client,
	name, imageName, workDir string,
	env []string,
	labels map[string]string,
	binds []Bind,
	engine Engine,
) (string, error) {
	mounts := make([]map[string]any, len(binds))
	for i, bind := range binds {
		mounts[i] = bind.libpodMount()
	}

	uid, gid := os.Getuid(), os.Getgid()
	spec := map[string]any{
		"name":         name,
		"image":        imageName,
		"command":      []string{"tail", "-f", "/dev/null"},
		"entrypoint":   []string{},
		"env":          envMap(env),
		"labels":       labels,
		"user":         fmt.Sprintf("%d:%d", uid, gid),
		"passwd_entry": fmt.Sprintf("dev:*:%d:%d::/home/dev:/bin/sh", uid, gid),
		"terminal":     false,
		"stdin":        false,
		"init":         true,
		"netns":        map[string]string{"nsmode": "host"},
		"mounts":       mounts,
	}
	if workDir != "" {
		spec["work_dir"] = workDir
	}
	if unconfined(binds) {
		// SELinux would deny the container access to the host's sockets (the
		// daemon socket, Wayland) because those keep the host's label. Disabling
		// labeling is what makes them usable from inside, see unconfined.
		spec["selinux_opts"] = []string{"disable"}
	}
	if engine.Rootless {
		// Map the host user to the same UID inside the container so files
		// created in bind mounts stay host-owned.
		spec["userns"] = map[string]string{"nsmode": "keep-id"}
	}

	return doCreate(client, "/libpod/containers/create", nil, spec)
}

func createDockerContainer(
	client *http.Client,
	name, imageName, workDir string,
	env []string,
	labels map[string]string,
	binds []Bind,
) (string, error) {
	dockerBinds := make([]string, len(binds))
	for i, bind := range binds {
		dockerBinds[i] = bind.dockerBind()
	}

	hostConfig := newHostConfig(dockerBinds)
	if unconfined(binds) {
		// Same reason as in createPodmanContainer, in the form the Docker API
		// spells it.
		hostConfig["SecurityOpt"] = []string{"label=disable"}
	}

	payload := map[string]any{
		"Image":      imageName,
		"Env":        env,
		"WorkingDir": workDir,
		"User":       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		"Entrypoint": []string{},
		"Cmd":        []string{"tail", "-f", "/dev/null"},
		"Tty":        false,
		"OpenStdin":  false,
		"Labels":     labels,
		"HostConfig": hostConfig,
	}

	return doCreate(client, "/containers/create", map[string]string{"name": name}, payload)
}

// doCreate posts a create payload and returns the new container's ID.
func doCreate(client *http.Client, path string, query map[string]string, payload any) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal create payload: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+path,
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("create create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	for key, value := range query {
		q.Set(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", ErrContainerExists
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create container: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var res struct {
		ID string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode create response: %w", err)
	}
	if res.ID == "" {
		return "", errors.New("create container: daemon returned an empty ID")
	}
	return res.ID, nil
}

func envMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, kv := range env {
		key, value, _ := strings.Cut(kv, "=")
		m[key] = value
	}
	return m
}
