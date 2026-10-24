package container

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// DefaultImage is used when no --image flag is given.
const DefaultImage = "ghcr.io/sebag90/devenv:latest"

// NormalizeImage adds the implicit :latest tag the way the CLIs do, so that
// "ghcr.io/sebag90/devenv" and "ghcr.io/sebag90/devenv:latest" refer to the
// same image.
func NormalizeImage(ref string) string {
	if ref == "" {
		return DefaultImage
	}
	if strings.Contains(ref, "@") {
		return ref // pinned by digest
	}
	if strings.Contains(ref[strings.LastIndex(ref, "/")+1:], ":") {
		return ref // already tagged (avoids mistaking a registry port for a tag)
	}
	return ref + ":latest"
}

// EnsureImage makes sure the image is available locally, pulling it if needed.
func EnsureImage(client *http.Client, image string) error {
	req, err := http.NewRequest(http.MethodGet, baseApiUrl+"/images/"+image+"/json", nil)
	if err != nil {
		return fmt.Errorf("create image inspect request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("inspect image %s: %w", image, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode == http.StatusNotFound:
		fmt.Fprintf(os.Stderr, "pulling %s\n", image)
		return pullImage(client, image, os.Stderr)
	default:
		return fmt.Errorf("inspect image %s: %s", image, resp.Status)
	}
}
