package container

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"testing"
)

// containerListTransport answers the container list with body and records the
// stop and remove requests, so the removal path can be tested without a
// daemon.
func containerListTransport(t *testing.T, body string) (*http.Client, *[]string) {
	t.Helper()

	var requests []string
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/containers/json"):
			return jsonResponse(http.StatusOK, body, r), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/stop"):
			requests = append(requests, path.Base(strings.TrimSuffix(r.URL.Path, "/stop")))
			return jsonResponse(http.StatusNoContent, "", r), nil
		case r.Method == http.MethodDelete:
			requests = append(requests, path.Base(r.URL.Path))
			return jsonResponse(http.StatusNoContent, "", r), nil
		default:
			return nil, fmt.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})}
	return client, &requests
}

func TestRemoveAllContainersRemovesOnlyManagedContainers(t *testing.T) {
	const body = `[
		{"Id":"1","Names":["/dv.11"],"State":"running","Labels":{"denver.project":"dv.11"}},
		{"Id":"2","Names":["/denver.22"],"State":"exited","Labels":{"denver.project":"denver.22"}},
		{"Id":"3","Names":["/dv.33"],"State":"running","Labels":{}},
		{"Id":"4","Names":["/unrelated"],"State":"running"},
		{"Id":"5","Names":["/dv.notahash"],"State":"running"}
	]
`
	client, requests := containerListTransport(t, body)

	names, err := RemoveAllContainers(client)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"dv.11", "denver.22", "dv.33"}; strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("removed names = %v, want %v", names, want)
	}
	// Running containers are stopped before they are removed; the exited one is
	// removed directly.
	if want := []string{"1", "1", "2", "3", "3"}; strings.Join(*requests, ",") != strings.Join(want, ",") {
		t.Errorf("stop and remove IDs = %v, want %v", *requests, want)
	}
}

func TestRemoveAllContainersWithoutManagedContainers(t *testing.T) {
	client, requests := containerListTransport(t, `[{"Id":"1","Names":["/unrelated"]}]`)

	names, err := RemoveAllContainers(client)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("removed names = %v, want none", names)
	}
	if len(*requests) != 0 {
		t.Errorf("requests = %v, want only the list request", *requests)
	}
}

func TestRemoveContainerFindsLegacyName(t *testing.T) {
	client, requests := containerListTransport(t,
		`[{"Id":"7","Names":["/denver.77"],"State":"running","Labels":{"denver.project":"denver.77"}}]`)

	removed, err := RemoveContainer(client, "denver.77")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Error("RemoveContainer() = false, want true")
	}
	if want := []string{"7", "7"}; strings.Join(*requests, ",") != strings.Join(want, ",") {
		t.Errorf("stop and remove IDs = %v, want %v", *requests, want)
	}
}
