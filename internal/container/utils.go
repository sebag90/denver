package container

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Container is the subset of the /containers/json payload we care about.
type Container struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	Image  string            `json:"Image"`
	State  string            `json:"State"`
	Status string            `json:"Status"`
	Labels map[string]string `json:"Labels"`
}

// Name returns the primary container name without its leading slash.
func (c Container) Name() string {
	if len(c.Names) == 0 {
		return ""
	}
	return strings.TrimPrefix(c.Names[0], "/")
}

// Running reports whether the container is currently running.
func (c Container) Running() bool {
	return c.State == "running"
}

const (
	// preferredApiVersion is the newest API version dv is written against.
	preferredApiVersion = "1.54"

	// apiHost is ignored by the unix socket transport, but requests need a
	// well formed URL to be built from.
	apiHost = "http://container"
)

// apiVersion and baseApiUrl describe the API version in use for every
// request. They start at the newest version dv knows and are lowered together
// by NegotiateApiVersion when the daemon speaks something older.
var (
	apiVersion = preferredApiVersion
	baseApiUrl = versionedUrl(apiVersion)
)

// versionedUrl renders the API root for one API version.
func versionedUrl(version string) string {
	return apiHost + "/v" + version
}

// ShortID trims a container ID to the 12 character form the CLIs use.
func ShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// ListContainers returns every container whose name matches name. The daemon
// treats the name filter as a substring match, so callers that need an exact
// name should use FindContainer.
func ListContainers(client *http.Client, name string) ([]Container, error) {
	req, err := http.NewRequest(http.MethodGet, baseApiUrl+"/containers/json", nil)
	if err != nil {
		return nil, fmt.Errorf("create list request: %w", err)
	}

	filters, err := json.Marshal(map[string][]string{"name": {name}})
	if err != nil {
		return nil, fmt.Errorf("marshal filters: %w", err)
	}

	q := req.URL.Query()
	q.Set("all", "true")
	q.Set("filters", string(filters))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list containers: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("decode container list: %w", err)
	}
	return containers, nil
}

// FindContainer returns the container with exactly the given name, or nil when
// no such container exists.
func FindContainer(client *http.Client, name string) (*Container, error) {
	containers, err := ListContainers(client, name)
	if err != nil {
		return nil, err
	}

	for i := range containers {
		if containers[i].Name() == name {
			return &containers[i], nil
		}
	}
	return nil, nil
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

// ContainerName derives a stable container name from the current directory, so
// every directory gets its own container.
func ContainerName() string {
	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dv: cannot determine working directory: %v\n", err)
	}
	return fmt.Sprintf("denver.%s", hash(path))
}
