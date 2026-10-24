package container

import (
	"fmt"
	"io"
	"net/http"
)

// StopContainer stops a running container. Stopping an already stopped
// container is not an error.
func StopContainer(client *http.Client, containerID string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/containers/"+containerID+"/stop?t=10",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create stop request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("stop container: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusNotModified: // 304: already stopped
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return nil
	default:
		return fmt.Errorf("stop container: %s", resp.Status)
	}
}

// RemoveContainer stops the named container if needed and removes it. The
// boolean reports whether a container was actually removed.
func RemoveContainer(client *http.Client, name string) (bool, error) {
	c, err := FindContainer(client, name)
	if err != nil {
		return false, err
	}
	if c == nil {
		return false, nil
	}

	if c.Running() {
		if err := StopContainer(client, c.ID); err != nil {
			return false, err
		}
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		baseApiUrl+"/containers/"+c.ID+"?force=1",
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("create remove request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("remove container: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return true, nil
	case resp.StatusCode == http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("remove container: %s", resp.Status)
	}
}
