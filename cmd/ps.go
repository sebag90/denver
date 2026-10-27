package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

// psAllFlag makes ps list the jobs of every running container instead of only
// the ones of the current directory's container.
var psAllFlag bool

func ps(cmd *cobra.Command, args []string) error {
	client, socket, err := connect()
	if err != nil {
		return err
	}

	targets, err := psTargets(client, psAllFlag)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		if psAllFlag {
			fmt.Printf("no running containers\n")
		} else {
			fmt.Printf("no running container for this directory\n")
		}
		return nil
	}

	var rows []jobRow
	for _, c := range targets {
		dir := c.Directory()
		if dir == "" && !psAllFlag {
			// The container for this directory was created here even when it
			// predates the directory label.
			dir, _ = os.Getwd()
		}

		jobs, err := container.ListJobs(client, socket, c.ID)
		if err != nil {
			return fmt.Errorf("%s: %w", c.Name(), err)
		}
		for _, job := range jobs {
			rows = append(rows, jobRow{Directory: dir, Job: job})
		}
	}
	if len(rows) == 0 {
		fmt.Printf("no jobs\n")
		return nil
	}

	fmt.Print(jobRows(rows, time.Now()))
	return nil
}

// psTargets returns the containers whose jobs ps lists: the container for the
// current directory, or every running container dv created when all is set.
// Jobs of stopped containers cannot be read, because their metadata lives
// inside the container.
func psTargets(client *http.Client, all bool) ([]container.Container, error) {
	if !all {
		c, err := container.FindContainer(client, container.ContainerName())
		if err != nil {
			return nil, err
		}
		if c == nil || !c.Running() {
			return nil, nil
		}
		return []container.Container{*c}, nil
	}

	containers, err := container.ListManagedContainers(client)
	if err != nil {
		return nil, err
	}
	running := make([]container.Container, 0, len(containers))
	for _, c := range containers {
		if c.Running() {
			running = append(running, c)
		}
	}
	return running, nil
}

// jobRow is a detached job together with the directory of the container it runs
// in, so that jobs from several containers can be listed in one table.
type jobRow struct {
	Directory string
	Job       container.Job
}

// jobRows renders detached jobs as an aligned table sorted by directory and
// start time, with the age relative to now so the output is stable in tests.
func jobRows(rows []jobRow, now time.Time) string {
	sorted := slices.Clone(rows)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Directory != sorted[j].Directory {
			return sorted[i].Directory < sorted[j].Directory
		}
		si, sj := sorted[i].Job.Started, sorted[j].Job.Started
		if !si.Equal(sj) {
			return si.Before(sj)
		}
		return sorted[i].Job.ID < sorted[j].Job.ID
	})

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JOB\tSTATUS\tEXIT\tAGE\tDIRECTORY\tCOMMAND")
	for _, r := range sorted {
		j := r.Job
		exit := "-"
		if j.HasExit {
			exit = strconv.Itoa(j.Exit)
		}
		age := "-"
		if !j.Started.IsZero() {
			age = formatAge(now.Sub(j.Started))
		}
		dir := r.Directory
		if dir == "" {
			dir = "-"
		}
		command := j.Command
		if command == "" {
			command = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", j.ID, j.Status, exit, age, dir, command)
	}
	w.Flush()
	return buf.String()
}

// formatAge renders a duration as a short, coarse age.
func formatAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:     "ps",
	Aliases: []string{"jobs"},
	Short:   "List detached jobs and the directory they run in",
	Long: `ps lists the commands started with "dv run --detach" in the container for
the current directory, together with their status, exit code, age and the
directory they run in.

A running job is one whose process is still alive; an exited job has finished
and reports its exit code; a lost job disappeared without recording one, for
example because the container was stopped while it ran.

With --all, ps lists the jobs of every running container dv created, not just
the current directory's. Jobs of stopped containers cannot be listed, because
their metadata lives inside the container.

Only jobs started by dv are listed. Use "dv run ps aux" to see every process
in the container, and "dv logs <job>" to read a job's output.`,
	Args: cobra.NoArgs,
	RunE: ps,
}

func init() {
	rootCmd.AddCommand(psCmd)

	psCmd.Flags().BoolVar(&psAllFlag, "all", false,
		"list the jobs of every running container, not just this directory's")
}
