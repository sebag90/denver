package container

import (
	"fmt"
	"io"
	"net/http"
)

// StartContainer starts a created or stopped container. Starting a container
// that is already running is not an error.
func StartContainer(client *http.Client, containerID string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/containers/"+containerID+"/start",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create start request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusNotModified: // 304: already running
		return nil
	default:
		return fmt.Errorf("start container: %s", resp.Status)
	}
}
