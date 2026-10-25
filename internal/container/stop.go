package container

import (
	"errors"
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
	return removeContainer(client, c.ID, c.Running())
}

// RemoveAllContainers stops and removes every container dv manages, whatever
// directory it belongs to. The names of the removed containers are returned.
// A failing container does not stop the remaining ones from being removed; the
// returned error joins the failures.
func RemoveAllContainers(client *http.Client) ([]string, error) {
	containers, err := ListManagedContainers(client)
	if err != nil {
		return nil, err
	}

	var (
		removed []string
		errs    []error
	)
	for _, c := range containers {
		ok, err := removeContainer(client, c.ID, c.Running())
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.Name(), err))
			continue
		}
		if ok {
			removed = append(removed, c.Name())
		}
	}
	return removed, errors.Join(errs...)
}

// removeContainer stops a running container and removes it. The boolean
// reports whether a container was actually removed.
func removeContainer(client *http.Client, containerID string, running bool) (bool, error) {
	// Stopping first lets the container shut down cleanly.
	if running {
		if err := StopContainer(client, containerID); err != nil {
			return false, err
		}
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		baseApiUrl+"/containers/"+containerID+"?force=1",
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
