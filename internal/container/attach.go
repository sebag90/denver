package container

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"golang.org/x/term"
)

// AttachToContainerHijack attaches to the container's main process using HTTP
// connection hijacking.
func AttachToContainerHijack(socketPath, containerId string) error {
	// 1. Connect directly to the Podman/Docker unix socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("dial unix socket: %w", err)
	}
	defer conn.Close()

	// 2. Build attach request with streaming query parameters and Upgrade headers
	endpoint := fmt.Sprintf("/v%s/containers/%s/attach?stream=1&stdin=1&stdout=1&stderr=1", apiVersion, containerId)
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Host = "container"
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	// 3. Write raw HTTP request to socket
	if err := req.Write(conn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	// 4. Read HTTP response from bufio.Reader
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols { // HTTP 101
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// The container's PTY is created with a default size (80x24). Without
	// telling the daemon our real window size, full-screen programs and line
	// wrapping use that default and the output only fills part of the screen.
	// Resize now (before we start pumping output) so the shell gets SIGWINCH
	// and redraws its prompt with the correct dimensions.
	resizeClient := NewClient(socketPath)
	stopResize := watchWindowSize(func(width, height int) error {
		return resizeContainer(resizeClient, containerId, width, height)
	})
	defer stopResize()

	// Put the terminal into raw mode so keystrokes are passed directly
	// without local echo or line buffering.
	stdinFd := int(os.Stdin.Fd())
	if term.IsTerminal(stdinFd) {
		oldState, err := term.MakeRaw(stdinFd)
		if err != nil {
			return fmt.Errorf("set raw terminal: %w", err)
		}
		defer term.Restore(stdinFd, oldState)
	}

	// 5. Pipe stdin -> container
	go func() {
		io.Copy(conn, os.Stdin)
	}()

	// 6. Pipe container -> stdout (read from reader to consume any buffered bytes)
	_, err = io.Copy(os.Stdout, reader)
	return err
}
