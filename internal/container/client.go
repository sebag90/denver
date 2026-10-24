package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SocketPath resolves the Podman (or Docker-compatible) API socket.
//
// A unix:// override from CONTAINER_HOST / DOCKER_HOST wins, otherwise the
// rootless Podman socket is used (falling back to /run/user/<uid> when
// XDG_RUNTIME_DIR is not set).
func SocketPath() string {
	for _, key := range []string{"CONTAINER_HOST", "DOCKER_HOST"} {
		if v := os.Getenv(key); strings.HasPrefix(v, "unix://") {
			return strings.TrimPrefix(v, "unix://")
		}
	}

	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "podman", "podman.sock")
	}
	return fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
}

// NewClient returns an http.Client that talks to the given unix socket.
func NewClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

// NegotiateApiVersion lowers the API version used by every request to the one
// the daemon advertises, and leaves it alone when the daemon is newer or
// reports something that is not a dotted number.
func NegotiateApiVersion(client *http.Client) error {
	// /version is requested without a version prefix: it is the one endpoint
	// every daemon answers, whatever version the client would otherwise ask
	// for.
	req, err := http.NewRequest(http.MethodGet, apiHost+"/version", nil)
	if err != nil {
		return fmt.Errorf("create version request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to container engine: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("detect API version: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var version struct {
		ApiVersion string `json:"ApiVersion"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return fmt.Errorf("decode version response: %w", err)
	}

	if !validApiVersion(version.ApiVersion) {
		return nil // keep the newest version dv knows
	}
	if compareApiVersions(version.ApiVersion, preferredApiVersion) < 0 {
		apiVersion = version.ApiVersion
		baseApiUrl = versionedUrl(apiVersion)
	}
	return nil
}

// validApiVersion reports whether v looks like a dotted numeric API version.
func validApiVersion(v string) bool {
	for _, part := range strings.Split(v, ".") {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return v != ""
}

// compareApiVersions compares two dotted numeric API versions, returning -1, 0
// or 1 when a is older than, equal to, or newer than b.
func compareApiVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(as), len(bs)); i++ {
		av, bv := 0, 0
		if i < len(as) {
			av, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bv, _ = strconv.Atoi(bs[i])
		}
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		}
	}
	return 0
}
