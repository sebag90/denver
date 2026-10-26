package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/sebag90/denver/internal/container"
)

func TestJobRows(t *testing.T) {
	now := time.Unix(10_000, 0)
	jobs := []container.Job{
		{ID: "bbbbbbbbbbbb", Status: "running", Started: now.Add(-90 * time.Second),
			Command: "sleep 100"},
		{ID: "aaaaaaaaaaaa", Status: "exited", Exit: 1, HasExit: true,
			Started: now.Add(-2 * time.Hour), Command: "false"},
		{ID: "cccccccccccc", Status: "lost"},
	}

	lines := strings.Split(strings.TrimSpace(jobRows(jobs, now)), "\n")
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4:\n%s", len(lines), strings.Join(lines, "\n"))
	}

	if got := strings.Fields(lines[0]); strings.Join(got, ",") != "JOB,STATUS,EXIT,AGE,COMMAND" {
		t.Errorf("header = %v", got)
	}
	if got := strings.Fields(lines[1]); strings.Join(got, ",") != "bbbbbbbbbbbb,running,-,1m,sleep,100" {
		t.Errorf("line 1 = %v", got)
	}
	if got := strings.Fields(lines[2]); strings.Join(got, ",") != "aaaaaaaaaaaa,exited,1,2h,false" {
		t.Errorf("line 2 = %v", got)
	}
	if got := strings.Fields(lines[3]); strings.Join(got, ",") != "cccccccccccc,lost,-,-,-" {
		t.Errorf("line 3 = %v", got)
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
