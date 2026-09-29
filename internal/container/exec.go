package container

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"golang.org/x/term"
)

type ExecConfig struct {
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	Tty          bool     `json:"Tty"`
	Cmd          []string `json:"Cmd"`
	Env          []string `json:"Env,omitempty"`
}

type ExecCreateResponse struct {
	ID string `json:"Id"`
}

type ExecStartConfig struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}

// createExecInstance sends a POST request to create an exec instance and returns its ID.
func createExecInstance(client *http.Client, containerId string, cmd []string) (string, error) {
	payload := ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          cmd,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal exec config: %w", err)
	}

	endpoint := fmt.Sprintf("%s/containers/%s/exec", baseApiUrl, containerId)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create exec request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute exec request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("daemon error %s: %s", resp.Status, body)
	}

	var res ExecCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode exec response: %w", err)
	}

	return res.ID, nil
}

// startExecHijack connects to the unix socket, starts the exec instance, and hijacks the stream.
func startExecHijack(socketPath, execId string) error {
	// 1. Connect directly to the Podman/Docker unix socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("dial unix socket: %w", err)
	}
	defer conn.Close()

	// 2. Prepare the start request with body {"Detach": false, "Tty": true}
	startPayload, err := json.Marshal(ExecStartConfig{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return fmt.Errorf("marshal start config: %w", err)
	}

	endpoint := fmt.Sprintf("/v%s/exec/%s/start", apiVersion, execId)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(startPayload))
	if err != nil {
		return fmt.Errorf("create start request: %w", err)
	}

	req.Host = "container"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	// 3. Send request
	if err := req.Write(conn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	// 4. Read response
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols { // HTTP 101
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// 5. Raw mode on host terminal
	stdinFd := int(os.Stdin.Fd())
	if term.IsTerminal(stdinFd) {
		oldState, err := term.MakeRaw(stdinFd)
		if err != nil {
			return fmt.Errorf("set raw terminal: %w", err)
		}
		defer term.Restore(stdinFd, oldState)
	}

	// 6. Pipe stdin -> container
	go func() {
		io.Copy(conn, os.Stdin)
	}()

	// 7. Pipe container -> stdout
	_, err = io.Copy(os.Stdout, reader)
	return err
}

// execInContainer is the high-level function that creates and starts an exec session.
func execInContainer(client *http.Client, socketPath, containerId string, cmd []string) error {
	execId, err := createExecInstance(client, containerId, cmd)
	if err != nil {
		return fmt.Errorf("create exec instance: %w", err)
	}

	return startExecHijack(socketPath, execId)
}
