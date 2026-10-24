package container

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// versionTransport answers /version with a canned ApiVersion, the way a daemon
// would, without needing a socket.
type versionTransport struct{ apiVersion string }

func (v versionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path != "/version" {
		return nil, fmt.Errorf("unexpected request path %q", req.URL.Path)
	}
	body := fmt.Sprintf(`{"Version":"x","ApiVersion":%q}`, v.apiVersion)
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func negotiate(t *testing.T, apiVersion string) string {
	t.Helper()
	t.Cleanup(func() {
		apiVersion = preferredApiVersion
		baseApiUrl = versionedUrl(apiVersion)
	})

	client := &http.Client{Transport: versionTransport{apiVersion: apiVersion}}
	if err := NegotiateApiVersion(client); err != nil {
		t.Fatalf("NegotiateApiVersion(%q): %v", apiVersion, err)
	}
	return baseApiUrl
}

func TestNegotiateApiVersionLowersForOlderDocker(t *testing.T) {
	if got := negotiate(t, "1.47"); got != versionedUrl("1.47") {
		t.Errorf("baseApiUrl = %q, want %q", got, versionedUrl("1.47"))
	}
}

func TestNegotiateApiVersionKeepsPreferredForNewerDaemon(t *testing.T) {
	if got := negotiate(t, "1.99"); got != versionedUrl(preferredApiVersion) {
		t.Errorf("baseApiUrl = %q, want %q", got, versionedUrl(preferredApiVersion))
	}
}

// Podman reports its libpod API version, which is far ahead of the Docker
// compatible one, so its prefix must stay where it was.
func TestNegotiateApiVersionKeepsPreferredForPodman(t *testing.T) {
	if got := negotiate(t, "5.8.7"); got != versionedUrl(preferredApiVersion) {
		t.Errorf("baseApiUrl = %q, want %q", got, versionedUrl(preferredApiVersion))
	}
}

func TestNegotiateApiVersionIgnoresUnparsableVersion(t *testing.T) {
	for _, version := range []string{"", "dev", "1.47-beta"} {
		if got := negotiate(t, version); got != versionedUrl(preferredApiVersion) {
			t.Errorf("NegotiateApiVersion(%q): baseApiUrl = %q, want %q",
				version, got, versionedUrl(preferredApiVersion))
		}
	}
}

func TestCompareApiVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.54", "1.54", 0},
		{"1.47", "1.54", -1},
		{"1.54", "1.47", 1},
		{"1.9", "1.10", -1}, // numeric, not lexical, comparison
		{"5.8.7", "1.54", 1},
		{"1.54", "1.54.0", 0},
	}
	for _, tc := range tests {
		if got := compareApiVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("compareApiVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
