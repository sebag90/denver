package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/sebag90/denver/internal/container"
)

func TestJobRows(t *testing.T) {
	now := time.Unix(10_000, 0)
	rows := []jobRow{
		{Directory: "/src/a", Job: container.Job{ID: "bbbbbbbbbbbb", Status: "running",
			Started: now.Add(-90 * time.Second), Command: "sleep 100"}},
		{Directory: "/src/a", Job: container.Job{ID: "aaaaaaaaaaaa", Status: "exited",
			Exit: 1, HasExit: true, Started: now.Add(-2 * time.Hour), Command: "false"}},
		{Directory: "/src/a", Job: container.Job{ID: "cccccccccccc", Status: "lost"}},
	}

	lines := strings.Split(strings.TrimSpace(jobRows(rows, now)), "\n")
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4:\n%s", len(lines), strings.Join(lines, "\n"))
	}

	if got := strings.Fields(lines[0]); strings.Join(got, ",") != "JOB,STATUS,EXIT,AGE,DIRECTORY,COMMAND" {
		t.Errorf("header = %v", got)
	}
	if got := strings.Fields(lines[1]); strings.Join(got, ",") != "cccccccccccc,lost,-,-,/src/a,-" {
		t.Errorf("line 1 = %v", got)
	}
	if got := strings.Fields(lines[2]); strings.Join(got, ",") != "aaaaaaaaaaaa,exited,1,2h,/src/a,false" {
		t.Errorf("line 2 = %v", got)
	}
	if got := strings.Fields(lines[3]); strings.Join(got, ",") != "bbbbbbbbbbbb,running,-,1m,/src/a,sleep,100" {
		t.Errorf("line 3 = %v", got)
	}
}

// TestJobRowsSortsByDirectory covers the output of "ps --all", where jobs of
// several containers end up in one table.
func TestJobRowsSortsByDirectory(t *testing.T) {
	now := time.Unix(10_000, 0)
	rows := []jobRow{
		{Directory: "/src/b", Job: container.Job{ID: "b1", Status: "running",
			Started: now.Add(-time.Hour), Command: "make"}},
		{Directory: "/src/a", Job: container.Job{ID: "a2", Status: "running",
			Started: now.Add(-time.Minute), Command: "sleep 1"}},
		{Directory: "", Job: container.Job{ID: "z1", Status: "running",
			Started: now.Add(-time.Second), Command: "true"}},
		{Directory: "/src/a", Job: container.Job{ID: "a1", Status: "running",
			Started: now.Add(-time.Hour), Command: "sleep 2"}},
	}

	lines := strings.Split(strings.TrimSpace(jobRows(rows, now)), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5:\n%s", len(lines), strings.Join(lines, "\n"))
	}

	// Sorted by directory, then by start time; a container without a directory
	// label shows "-" and comes first.
	want := []string{
		"z1,running,-,1s,-,true",
		"a1,running,-,1h,/src/a,sleep,2",
		"a2,running,-,1m,/src/a,sleep,1",
		"b1,running,-,1h,/src/b,make",
	}
	for i, w := range want {
		if got := strings.Fields(lines[i+1]); strings.Join(got, ",") != w {
			t.Errorf("line %d = %v, want %s", i+1, got, w)
		}
	}
}

func TestJobRowsDoesNotReorderInput(t *testing.T) {
	rows := []jobRow{
		{Directory: "/b", Job: container.Job{ID: "b", Status: "running"}},
		{Directory: "/a", Job: container.Job{ID: "a", Status: "running"}},
	}

	jobRows(rows, time.Unix(1, 0))

	if rows[0].Directory != "/b" {
		t.Errorf("input was reordered: %v", rows[0].Directory)
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{-time.Second, "0s"},
		{0, "0s"},
		{59 * time.Second, "59s"},
		{time.Minute, "1m"},
		{59 * time.Minute, "59m"},
		{time.Hour, "1h"},
		{23 * time.Hour, "23h"},
		{24 * time.Hour, "1d"},
		{72 * time.Hour, "3d"},
	}
	for _, tt := range tests {
		if got := formatAge(tt.d); got != tt.want {
			t.Errorf("formatAge(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
