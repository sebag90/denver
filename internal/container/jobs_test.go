package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStartJobStartsDetachedExec(t *testing.T) {
	var create map[string]any
	var start map[string]any

	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/containers/abc/exec"):
			if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
				t.Fatalf("decode create payload: %v", err)
			}
			return jsonResponse(http.StatusCreated, `{"Id":"exec1"}`, r), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/exec/exec1/start"):
			if err := json.NewDecoder(r.Body).Decode(&start); err != nil {
				t.Fatalf("decode start payload: %v", err)
			}
			return jsonResponse(http.StatusOK, "", r), nil
		default:
			return nil, fmt.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})}

	jobID, err := StartJob(client, "abc", []string{"sleep", "100"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !ValidJobID(jobID) {
		t.Errorf("job ID %q is not valid", jobID)
	}
	if start["Detach"] != true {
		t.Errorf("start payload Detach = %v, want true", start["Detach"])
	}
	for _, key := range []string{"AttachStdin", "AttachStdout", "AttachStderr", "Tty"} {
		if create[key] != false {
			t.Errorf("create payload %s = %v, want false", key, create[key])
		}
	}

	cmd, ok := create["Cmd"].([]any)
	if !ok {
		t.Fatalf("create payload Cmd = %T, want []any", create["Cmd"])
	}
	want := []any{
		"/bin/sh", "-c", startJobScript, "dv-job", jobID,
		JobsDir + "/" + jobID + "/log",
	}
	// The start time sits between the wrapper and the command.
	if len(cmd) != len(want)+3 {
		t.Fatalf("Cmd = %v, want %d entries", cmd, len(want)+3)
	}
	for i, value := range want {
		if cmd[i] != value {
			t.Errorf("Cmd[%d] = %v, want %v", i, cmd[i], value)
		}
	}
	if _, err := strconv.ParseInt(cmd[6].(string), 10, 64); err != nil {
		t.Errorf("Cmd[6] = %v, want a unix timestamp", cmd[6])
	}
	if cmd[7] != "sleep" || cmd[8] != "100" {
		t.Errorf("command was not appended: %v", cmd[7:])
	}
}

func TestStartJobUsesGivenLogPath(t *testing.T) {
	var create map[string]any
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/exec"):
			json.NewDecoder(r.Body).Decode(&create)
			return jsonResponse(http.StatusCreated, `{"Id":"exec1"}`, r), nil
		case strings.HasSuffix(r.URL.Path, "/start"):
			return jsonResponse(http.StatusOK, "", r), nil
		default:
			return nil, fmt.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})}

	if _, err := StartJob(client, "abc", []string{"true"}, "/tmp/out.log"); err != nil {
		t.Fatal(err)
	}
	cmd := create["Cmd"].([]any)
	if cmd[5] != "/tmp/out.log" {
		t.Errorf("Cmd[5] = %v, want /tmp/out.log", cmd[5])
	}
}

// TestJobScripts runs the wrapper, list and log scripts against a temporary
// jobs directory. They are plain sh, so this also catches syntax errors that
// the Go compiler cannot see.
func TestJobScripts(t *testing.T) {
	jobsDir := filepath.Join(t.TempDir(), "jobs")
	script := func(s string) string { return strings.ReplaceAll(s, JobsDir, jobsDir) }

	const jobID = "aaaaaaaaaaaa"
	logPath := filepath.Join(jobsDir, jobID, "log")
	const started = "1700000000"

	start := exec.Command("sh", "-c", script(startJobScript),
		"dv-job", jobID, logPath, started,
		"sh", "-c", "echo hello; exit 3")
	if out, err := start.CombinedOutput(); err != nil {
		t.Fatalf("start script: %v\n%s", err, out)
	}

	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(jobsDir, jobID, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return strings.TrimSpace(string(b))
	}

	if got := read("exit"); got != "3" {
		t.Errorf("exit = %q, want 3", got)
	}
	if got := read("started"); got != started {
		t.Errorf("started = %q, want %q", got, started)
	}
	if got := read("logpath"); got != logPath {
		t.Errorf("log = %q, want %q", got, logPath)
	}
	if got := strings.ReplaceAll(read("cmd"), "\n", " "); got != "sh -c echo hello; exit 3" {
		t.Errorf("cmd = %q", got)
	}
	if _, err := strconv.Atoi(read("pid")); err != nil {
		t.Errorf("pid is not numeric: %v", err)
	}
	if b, err := os.ReadFile(logPath); err != nil || string(b) != "hello\n" {
		t.Errorf("log contents = %q, %v", b, err)
	}

	list := exec.Command("sh", "-c", script(listJobsScript))
	out, err := list.Output()
	if err != nil {
		t.Fatalf("list script: %v", err)
	}
	jobs, err := parseJobs(string(out))
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1: %v", len(jobs), jobs)
	}
	if jobs[0].ID != jobID || jobs[0].Status != "exited" || jobs[0].Exit != 3 {
		t.Errorf("job = %+v", jobs[0])
	}

	for _, follow := range []bool{false, true} {
		mode := "print"
		if follow {
			mode = "follow"
		}
		logs := exec.Command("sh", "-c", script(logJobScript), "dv-logs", jobID, mode)
		out, err := logs.Output()
		if err != nil {
			t.Fatalf("log script (follow=%v): %v", follow, err)
		}
		if string(out) != "hello\n" {
			t.Errorf("log script (follow=%v) = %q", follow, out)
		}
	}
}

func TestParseJobs(t *testing.T) {
	out := strings.Join([]string{
		"bbbbbbbbbbbb\trunning\t\t1700000100\t/tmp/dv/jobs/bbbbbbbbbbbb/log\tsleep 100",
		"aaaaaaaaaaaa\texited\t1\t1700000000\t/tmp/dv/jobs/aaaaaaaaaaaa/log\tfalse",
		"cccccccccccc\tlost\t\t1700000200\t/tmp/dv/jobs/cccccccccccc/log\tsh",
		"",
	}, "\r\n")

	jobs, err := parseJobs(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 3 {
		t.Fatalf("got %d jobs, want 3", len(jobs))
	}

	// Sorted by start time, oldest first.
	if jobs[0].ID != "aaaaaaaaaaaa" || jobs[1].ID != "bbbbbbbbbbbb" || jobs[2].ID != "cccccccccccc" {
		t.Errorf("job order = %s, %s, %s", jobs[0].ID, jobs[1].ID, jobs[2].ID)
	}

	exited := jobs[0]
	if exited.Status != "exited" || !exited.HasExit || exited.Exit != 1 {
		t.Errorf("exited job = %+v", exited)
	}
	if !exited.Started.Equal(time.Unix(1700000000, 0)) {
		t.Errorf("started = %v", exited.Started)
	}
	if exited.Command != "false" || exited.Log != "/tmp/dv/jobs/aaaaaaaaaaaa/log" {
		t.Errorf("exited job = %+v", exited)
	}

	running := jobs[1]
	if running.Status != "running" || running.HasExit {
		t.Errorf("running job = %+v", running)
	}
	if running.Command != "sleep 100" {
		t.Errorf("command = %q", running.Command)
	}
}

func TestParseJobsRejectsMalformedRecords(t *testing.T) {
	for _, out := range []string{
		"only\ttwo\n",
		"aaaaaaaaaaaa\trunning\tx\t1700000000\t/tmp/log\tcmd\n",
		"aaaaaaaaaaaa\trunning\t\tnotatime\t/tmp/log\tcmd\n",
	} {
		if _, err := parseJobs(out); err == nil {
			t.Errorf("parseJobs(%q) succeeded, want an error", out)
		}
	}
}

func TestValidJobID(t *testing.T) {
	for _, id := range []string{"0123456789ab", "ffffffffffff", "0123456789AB"} {
		if !ValidJobID(id) {
			t.Errorf("ValidJobID(%q) = false, want true", id)
		}
	}
	for _, id := range []string{"", "0123456789a", "0123456789abc", "0123456789ag", "../etc/passwd"} {
		if ValidJobID(id) {
			t.Errorf("ValidJobID(%q) = true, want false", id)
		}
	}
}

func TestLogCommand(t *testing.T) {
	printCmd := LogCommand("0123456789ab", false)
	if printCmd[3] != "dv-logs" || printCmd[4] != "0123456789ab" || printCmd[5] != "print" {
		t.Errorf("LogCommand = %v", printCmd)
	}
	followCmd := LogCommand("0123456789ab", true)
	if followCmd[5] != "follow" {
		t.Errorf("follow LogCommand = %v", followCmd)
	}
}
