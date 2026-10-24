package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Engine describes the container daemon behind the socket.
type Engine struct {
	Podman   bool
	Rootless bool
}

// DetectEngine reports which daemon is listening on the socket and whether it
// is running rootless.
//
// Podman advertises itself as a "Podman Engine" component in /version, while
// /info reports "name=rootless" in its security options for both Podman and
// Docker. Engine detection is needed because rootless Podman requires a user
// namespace mapping for containers to run as the host user.
func DetectEngine(client *http.Client) (Engine, error) {
	var e Engine

	var version struct {
		Components []struct {
			Name string `json:"Name"`
		} `json:"Components"`
	}
	if err := getJSON(client, "/version", &version); err != nil {
		return e, fmt.Errorf("detect engine version: %w", err)
	}
	for _, c := range version.Components {
		if c.Name == "Podman Engine" {
			e.Podman = true
			break
		}
	}

	var info struct {
		SecurityOptions []string `json:"SecurityOptions"`
	}
	if err := getJSON(client, "/info", &info); err != nil {
		return e, fmt.Errorf("detect engine info: %w", err)
	}
	for _, opt := range info.SecurityOptions {
		if opt == "name=rootless" {
			e.Rootless = true
			break
		}
	}

	return e, nil
}

func getJSON(client *http.Client, path string, dst any) error {
	req, err := http.NewRequest(http.MethodGet, baseApiUrl+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, body)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
