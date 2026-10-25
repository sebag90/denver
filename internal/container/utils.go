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

// Directory returns the working directory the container was created in, or an
// empty string for containers created before the directory label existed.
func (c Container) Directory() string {
	return c.Labels[directoryLabel]
}

// Sandbox reports whether the container was created in sandbox mode.
// Containers from before the sandbox label existed are treated as normal
// containers.
func (c Container) Sandbox() bool {
	return c.Labels[sandboxLabel] == "true"
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
	return listContainers(client, name)
}

// listContainers returns the containers matching name, or every container when
// name is empty.
func listContainers(client *http.Client, name string) ([]Container, error) {
	req, err := http.NewRequest(http.MethodGet, baseApiUrl+"/containers/json", nil)
	if err != nil {
		return nil, fmt.Errorf("create list request: %w", err)
	}

	q := req.URL.Query()
	q.Set("all", "true")
	if name != "" {
		filters, err := json.Marshal(map[string][]string{"name": {name}})
		if err != nil {
			return nil, fmt.Errorf("marshal filters: %w", err)
		}
		q.Set("filters", string(filters))
	}
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

// containerNamePrefix starts the name of every container dv creates for a
// project directory.
const containerNamePrefix = "dv."

// legacyContainerNamePrefix was used before the names were shortened from
// denver.<hash> to dv.<hash>. It is still recognised so that containers
// created by older versions can be found, replaced and removed.
const legacyContainerNamePrefix = "denver."

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

// ContainerName derives a stable container name from the current directory, so
// every directory gets its own container.
func ContainerName() string {
	return nameForDirectory(containerNamePrefix)
}

// LegacyContainerName returns the container name that dv versions before the
// dv.<hash> rename derived from the current directory.
func LegacyContainerName() string {
	return nameForDirectory(legacyContainerNamePrefix)
}

func nameForDirectory(prefix string) string {
	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dv: cannot determine working directory: %v\n", err)
	}
	return prefix + hash(path)
}

// ListManagedContainers returns every container dv created, whatever directory
// it belongs to, including containers that still carry the legacy
// denver.<hash> name. Containers made before dv added labels are matched by
// name, so the whole list is fetched and filtered here.
func ListManagedContainers(client *http.Client) ([]Container, error) {
	containers, err := listContainers(client, "")
	if err != nil {
		return nil, err
	}

	managed := make([]Container, 0, len(containers))
	for _, c := range containers {
		if isManagedContainer(c) {
			managed = append(managed, c)
		}
	}
	return managed, nil
}

// isManagedContainer reports whether dv created the container: it carries one
// of dv's labels or has one of dv's name prefixes.
func isManagedContainer(c Container) bool {
	if _, ok := c.Labels[projectLabel]; ok {
		return true
	}
	if _, ok := c.Labels[sandboxLabel]; ok {
		return true
	}
	return isManagedName(c.Name())
}

// isManagedName reports whether name looks like a container dv derived from a
// directory, in either the current or the legacy naming scheme.
func isManagedName(name string) bool {
	for _, prefix := range []string{containerNamePrefix, legacyContainerNamePrefix} {
		if suffix, ok := strings.CutPrefix(name, prefix); ok && isDigits(suffix) {
			return true
		}
	}
	return false
}

// isDigits reports whether s is a non-empty run of decimal digits, which is
// what the directory hash looks like.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
