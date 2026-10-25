package cmd

import (
	"strings"
	"testing"

	"github.com/sebag90/denver/internal/container"
)

func TestListRows(t *testing.T) {
	containers := []container.Container{
		{
			Names:  []string{"/dv.22"},
			State:  "running",
			Labels: map[string]string{"denver.directory": "/src/b"},
		},
		{
			Names:  []string{"/denver.11"},
			State:  "exited",
			Labels: map[string]string{"denver.directory": "/src/a", "denver.sandbox": "true"},
		},
		{
			Names: []string{"/dv.33"},
			State: "created",
		},
	}

	lines := strings.Split(strings.TrimSpace(listRows(containers)), "\n")
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4:\n%s", len(lines), strings.Join(lines, "\n"))
	}

	if got := strings.Fields(lines[0]); strings.Join(got, ",") != "NAME,DIRECTORY,STATUS,SANDBOX" {
		t.Errorf("header = %v", got)
	}
	// Sorted by directory; the container without a directory comes first.
	if got := strings.Fields(lines[1]); strings.Join(got, ",") != "dv.33,-,created,no" {
		t.Errorf("line 1 = %v", got)
	}
	if got := strings.Fields(lines[2]); strings.Join(got, ",") != "denver.11,/src/a,exited,yes" {
		t.Errorf("line 2 = %v", got)
	}
	if got := strings.Fields(lines[3]); strings.Join(got, ",") != "dv.22,/src/b,running,no" {
		t.Errorf("line 3 = %v", got)
	}
}

func TestListRowsDoesNotReorderInput(t *testing.T) {
	containers := []container.Container{
		{Names: []string{"/dv.2"}, Labels: map[string]string{"denver.directory": "/b"}},
		{Names: []string{"/dv.1"}, Labels: map[string]string{"denver.directory": "/a"}},
	}

	listRows(containers)

	if containers[0].Name() != "dv.2" {
		t.Errorf("input was reordered: %v", containers[0].Name())
	}
}
