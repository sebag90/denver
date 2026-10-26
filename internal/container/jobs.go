package container

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// JobsDir is where a container keeps the metadata and logs of detached jobs.
// It lives in the container's writable layer, so jobs and their logs disappear
// when the container is removed.
const JobsDir = "/tmp/dv/jobs"

// Job is a command started by StartJob.
type Job struct {
	ID      string
	Status  string // "running", "exited" or "lost"
	Exit    int
	HasExit bool
	Started time.Time
	Command string
	Log     string
}

// startJobScript records the job metadata and then runs the command with its
// output redirected to the log file. The redirects matter: without them the
// detached command would keep the exec session's pipes open.
//
// Arguments: job ID, log path, start time (unix seconds), command and args.
// It is plain POSIX sh so it works in any image.
const startJobScript = `job=$1
log=$2
started=$3
shift 3
dir=/tmp/dv/jobs/$job
mkdir -p "$dir" || exit 1
printf '%s\n' "$@" > "$dir/cmd" || exit 1
printf '%s\n' "$$" > "$dir/pid" || exit 1
printf '%s\n' "$started" > "$dir/started" || exit 1
printf '%s\n' "$log" > "$dir/logpath" || exit 1
"$@" >"$log" 2>&1 </dev/null
code=$?
printf '%s\n' "$code" > "$dir/exit"
`

// listJobsScript prints one tab separated record per job:
// ID, status, exit code, start time, log path, command.
const listJobsScript = `dir=/tmp/dv/jobs
for jobdir in "$dir"/*/; do
	[ -d "$jobdir" ] || continue
	id=${jobdir%/}
	id=${id##*/}
	[ -f "$jobdir/pid" ] || continue
	pid=$(cat "$jobdir/pid" 2>/dev/null)
	started=$(cat "$jobdir/started" 2>/dev/null)
	log=$(cat "$jobdir/logpath" 2>/dev/null)
	cmd=$(tr '\t\n' '  ' < "$jobdir/cmd" 2>/dev/null)
	if [ -f "$jobdir/exit" ]; then
		status=exited
		code=$(cat "$jobdir/exit" 2>/dev/null)
	elif kill -0 "$pid" 2>/dev/null; then
		status=running
		code=
	else
		status=lost
		code=
	fi
	printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$id" "$status" "$code" "$started" "$log" "$cmd"
done
`

// logJobScript prints the log of a job, or follows it when it is still running
// and the second argument is "follow".
const logJobScript = `dir=/tmp/dv/jobs/$1
[ -d "$dir" ] || { echo "dv: no such job: $1" >&2; exit 1; }
log=$(cat "$dir/logpath" 2>/dev/null) || { echo "dv: job $1 has no log recorded" >&2; exit 1; }
[ -f "$log" ] || { echo "dv: log file is missing for job $1: $log" >&2; exit 1; }
if [ "$2" = follow ] && [ ! -f "$dir/exit" ]; then
	exec tail -n +1 -f "$log"
fi
exec cat "$log"
`

// ValidJobID reports whether id has the shape StartJob generates. It keeps
// user input out of the shell snippets that look up a job.
func ValidJobID(id string) bool {
	if len(id) != 12 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}

// StartJob starts argv as a detached exec session and returns the job ID. The
// command runs independently of the caller; its output goes to logPath (inside
// the container), which defaults to a file in the job's directory.
//
// ensureContainer must already have made sure the container is running.
func StartJob(client *http.Client, containerID string, argv []string, logPath string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("no command given")
	}

	jobID, err := newJobID()
	if err != nil {
		return "", err
	}
	if logPath == "" {
		logPath = JobsDir + "/" + jobID + "/log"
	}

	cmd := []string{
		"/bin/sh", "-c", startJobScript, "dv-job",
		jobID, logPath, strconv.FormatInt(time.Now().Unix(), 10),
	}
	cmd = append(cmd, argv...)

	execID, err := createExec(client, containerID, ExecConfig{Cmd: cmd})
	if err != nil {
		return "", err
	}
	if err := startExecDetached(client, execID); err != nil {
		return "", err
	}
	return jobID, nil
}

// startExecDetached starts an exec instance and returns without attaching to
// it. The daemon keeps running the process in the background; the caller gets
// no stream and no exit code, which is why StartJob records both in the job
// metadata.
func startExecDetached(client *http.Client, execID string) error {
	jsonData, err := json.Marshal(execStartConfig{Detach: true})
	if err != nil {
		return fmt.Errorf("marshal start config: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseApiUrl+"/exec/"+execID+"/start",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return fmt.Errorf("create start request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("start detached exec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("start detached exec: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

// ListJobs returns the detached jobs recorded in the container, sorted by
// start time (oldest first).
func ListJobs(client *http.Client, socketPath, containerID string) ([]Job, error) {
	out, code, err := Output(client, socketPath, containerID, []string{"/bin/sh", "-c", listJobsScript})
	if err != nil {
		return nil, err
	}
	if code != 0 {
		return nil, fmt.Errorf("list jobs: command exited with code %d", code)
	}
	return parseJobs(out)
}

// parseJobs turns the output of listJobsScript into jobs.
func parseJobs(out string) ([]Job, error) {
	var jobs []Job
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, "\t", 6)
		if len(fields) != 6 {
			return nil, fmt.Errorf("unexpected job record: %q", line)
		}

		job := Job{
			ID:      fields[0],
			Status:  fields[1],
			Log:     fields[4],
			Command: fields[5],
		}
		if fields[2] != "" {
			code, err := strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("parse exit code of job %s: %w", job.ID, err)
			}
			job.Exit, job.HasExit = code, true
		}
		if fields[3] != "" {
			seconds, err := strconv.ParseInt(fields[3], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse start time of job %s: %w", job.ID, err)
			}
			job.Started = time.Unix(seconds, 0)
		}
		jobs = append(jobs, job)
	}

	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].Started.Equal(jobs[j].Started) {
			return jobs[i].ID < jobs[j].ID
		}
		return jobs[i].Started.Before(jobs[j].Started)
	})
	return jobs, nil
}

// LogCommand builds the exec command that prints or follows a job's log.
func LogCommand(jobID string, follow bool) []string {
	mode := "print"
	if follow {
		mode = "follow"
	}
	return []string{"/bin/sh", "-c", logJobScript, "dv-logs", jobID, mode}
}

func newJobID() (string, error) {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate job ID: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
