package container

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/term"
)

// ExecConfig is the payload for POST /containers/{id}/exec.
type ExecConfig struct {
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	Tty          bool     `json:"Tty"`
	Cmd          []string `json:"Cmd"`
	Env          []string `json:"Env,omitempty"`
}

type execCreateResponse struct {
	ID string `json:"Id"`
}

type execStartConfig struct {
	Detach bool `json:"Detach"`
	Tty    bool `json:"Tty"`
}

type execInspectResponse struct {
	Running  bool `json:"Running"`
	ExitCode int  `json:"ExitCode"`
}

// Exec runs cmd inside the container and returns its exit code.
// Exec never touches the container's main process: ending a session only ends
// the exec'd command, so the container keeps running.
func Exec(client *http.Client, socketPath, containerID string, cmd []string, tty bool) (int, error) {
	execID, err := createExecInstance(client, containerID, cmd, tty)
	if err != nil {
		return -1, err
	}

	if err := startExec(socketPath, execID, tty); err != nil {
		return -1, err
	}

	return waitExecExit(client, execID)
}

// createExecInstance sends a POST request to create an exec instance and
// returns its ID.
func createExecInstance(client *http.Client, containerID string, cmd []string, tty bool) (string, error) {
	payload := ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          tty,
		Cmd:          cmd,
		// The exec'd process gets the container's environment plus this, so
		// TERM has to be passed explicitly - it is not inherited from the host.
		Env: []string{"TERM=" + terminalName()},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal exec config: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/containers/"+containerID+"/exec",
		bytes.NewReader(jsonData),
	)
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

	var res execCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode exec response: %w", err)
	}
	if res.ID == "" {
		return "", errors.New("daemon returned an empty exec ID")
	}
	return res.ID, nil
}

// startExec hijacks the connection and streams the exec session until it ends.
func startExec(socketPath, execID string, tty bool) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("dial unix socket: %w", err)
	}
	defer conn.Close()

	startPayload, err := json.Marshal(execStartConfig{Detach: false, Tty: tty})
	if err != nil {
		return fmt.Errorf("marshal start config: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/v%s/exec/%s/start", apiVersion, execID),
		bytes.NewReader(startPayload),
	)
	if err != nil {
		return fmt.Errorf("create start request: %w", err)
	}
	req.Host = "container"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	// Write the raw request to the socket so net/http does not try to manage
	// the connection for us.
	if err := req.Write(conn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols { // HTTP 101
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("unexpected status: %s: %s", resp.Status, bytes.TrimSpace(body))
	}

	if tty {
		// The exec PTY starts at the daemon's default size (80x24). Resize it
		// before pumping output so the shell redraws with the real dimensions,
		// and keep it in sync on SIGWINCH.
		resizeClient := NewClient(socketPath)
		stopResize := watchWindowSize(func(width, height int) error {
			return resizeExec(resizeClient, execID, width, height)
		})
		defer stopResize()

		stdinFd := int(os.Stdin.Fd())
		if term.IsTerminal(stdinFd) {
			oldState, err := term.MakeRaw(stdinFd)
			if err != nil {
				return fmt.Errorf("set raw terminal: %w", err)
			}
			defer term.Restore(stdinFd, oldState)
		}
	}

	// stdin -> container
	go io.Copy(conn, os.Stdin)

	// container -> terminal
	if tty {
		_, err = io.Copy(os.Stdout, reader)
		return err
	}
	return demux(reader, os.Stdout, os.Stderr)
}

// demux splits the Docker/Podman multiplexed stream (8 byte header followed by
// the payload) into stdout and stderr. It is used for non-tty sessions; with a
// tty the daemon sends a raw stream instead.
func demux(r io.Reader, stdout, stderr io.Writer) error {
	header := make([]byte, 8)

	for {
		if _, err := io.ReadFull(r, header); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil
			}
			return err
		}

		size := binary.BigEndian.Uint32(header[4:8])
		if size == 0 {
			continue
		}

		var dst io.Writer = stdout
		if header[0] == 2 { // 1 = stdout, 2 = stderr
			dst = stderr
		}
		if _, err := io.CopyN(dst, r, int64(size)); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return nil
			}
			return err
		}
	}
}

// waitExecExit asks the daemon for the exit code of the finished exec session.
func waitExecExit(client *http.Client, execID string) (int, error) {
	deadline := time.Now().Add(5 * time.Second)

	for {
		res, err := inspectExec(client, execID)
		if err != nil {
			return -1, err
		}
		if !res.Running || time.Now().After(deadline) {
			return res.ExitCode, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func inspectExec(client *http.Client, execID string) (execInspectResponse, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		baseApiUrl+"/exec/"+execID+"/json",
		nil,
	)
	if err != nil {
		return execInspectResponse{}, fmt.Errorf("create exec inspect request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return execInspectResponse{}, fmt.Errorf("inspect exec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return execInspectResponse{}, fmt.Errorf("inspect exec: %s: %s", resp.Status, body)
	}

	var res execInspectResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return execInspectResponse{}, fmt.Errorf("decode exec inspect response: %w", err)
	}
	return res, nil
}

// terminalName reports the terminal type to advertise to the container.
func terminalName() string {
	if t := os.Getenv("TERM"); t != "" {
		return t
	}
	return "xterm-256color"
}
