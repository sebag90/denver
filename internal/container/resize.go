package container

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// newSocketClient returns an http.Client that talks to the Podman/Docker unix socket.
func newSocketClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

// resizeContainer resizes the container's TTY (TIOCSWINSZ under the hood).
func resizeContainer(client *http.Client, containerId string, width, height int) error {
	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/containers/"+containerId+"/resize",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create resize request: %w", err)
	}

	q := req.URL.Query()
	q.Set("h", fmt.Sprint(height))
	q.Set("w", fmt.Sprint(width))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resize request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resize failed: %s", resp.Status)
	}
	return nil
}

// resizeExec resizes an exec session's TTY.
func resizeExec(client *http.Client, execId string, width, height int) error {
	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/exec/"+execId+"/resize",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create resize request: %w", err)
	}

	q := req.URL.Query()
	q.Set("h", fmt.Sprint(height))
	q.Set("w", fmt.Sprint(width))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resize request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resize failed: %s", resp.Status)
	}
	return nil
}

// watchWindowSize applies the current terminal size immediately and then again
// on every SIGWINCH (window resize). It returns a func that stops the watcher.
func watchWindowSize(fn func(width, height int) error) func() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return func() {}
	}

	apply := func() {
		w, h, err := term.GetSize(fd)
		if err != nil {
			return
		}
		fn(w, h)
	}

	apply()

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-winch:
				apply()
			case <-done:
				return
			}
		}
	}()

	return func() {
		signal.Stop(winch)
		close(done)
	}
}
