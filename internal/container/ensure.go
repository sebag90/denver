package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EnsureContainer makes sure a running container for the current directory
// exists. It creates, starts and waits for the container when necessary, and is
// safe to call repeatedly.
func EnsureContainer(client *http.Client, image string, sandbox bool) (string, error) {
	image = NormalizeImage(image)
	name := ContainerName()

	for attempt := 0; attempt < 2; attempt++ {
		c, err := FindContainer(client, name)
		if err != nil {
			return "", err
		}

		if c == nil {
			if err := EnsureImage(client, image); err != nil {
				return "", err
			}

			id, err := CreateContainer(client, image, sandbox)
			if errors.Is(err, ErrContainerExists) {
				// Someone else created it between list and create; reuse it.
				continue
			}
			if err != nil {
				return "", err
			}

			if err := startAndWait(client, id); err != nil {
				return "", err
			}
			return id, nil
		}

		if err := checkSandbox(c, sandbox); err != nil {
			return "", err
		}
		if c.Running() {
			return c.ID, nil
		}
		if err := startAndWait(client, c.ID); err != nil {
			return "", err
		}
		return c.ID, nil
	}

	return "", errors.New("could not create container")
}

// checkSandbox reports whether an existing container matches the requested
// sandbox setting. Containers created before the sandbox label existed have no
// label and are treated as non-sandbox containers.
func checkSandbox(c *Container, sandbox bool) error {
	if (c.Labels[sandboxLabel] == "true") == sandbox {
		return nil
	}
	return fmt.Errorf("container %s already exists with a different sandbox setting; run \"dv down\" and try again", c.Name())
}

func startAndWait(client *http.Client, containerID string) error {
	if err := StartContainer(client, containerID); err != nil {
		return err
	}
	return WaitRunning(client, containerID, 10*time.Second)
}

// WaitRunning polls the container until it reports Running or the timeout
// expires.
func WaitRunning(client *http.Client, containerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		running, err := InspectRunning(client, containerID)
		if err != nil {
			return err
		}
		if running {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("container %s did not reach running state", ShortID(containerID))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// InspectRunning reports whether the container's main process is running.
func InspectRunning(client *http.Client, containerID string) (bool, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		baseApiUrl+"/containers/"+containerID+"/json",
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("create inspect request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("inspect container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("inspect container: %s: %s", resp.Status, body)
	}

	var res struct {
		State struct {
			Running bool `json:"Running"`
		} `json:"State"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, fmt.Errorf("decode inspect response: %w", err)
	}
	return res.State.Running, nil
}
