package cmd

import (
	"bytes"
	"fmt"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

func ps(cmd *cobra.Command, args []string) error {
	client, socket, err := connect()
	if err != nil {
		return err
	}

	c, err := container.FindContainer(client, container.ContainerName())
	if err != nil {
		return err
	}
	if c == nil || !c.Running() {
		fmt.Printf("no running container for this directory\n")
		return nil
	}

	jobs, err := container.ListJobs(client, socket, c.ID)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		fmt.Printf("no jobs\n")
		return nil
	}

	fmt.Print(jobRows(jobs, time.Now()))
	return nil
}

// jobRows renders detached jobs as an aligned table, with the age relative to
// now so the output is stable in tests.
func jobRows(jobs []container.Job, now time.Time) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JOB\tSTATUS\tEXIT\tAGE\tCOMMAND")
	for _, j := range jobs {
		exit := "-"
		if j.HasExit {
			exit = strconv.Itoa(j.Exit)
		}
		age := "-"
		if !j.Started.IsZero() {
			age = formatAge(now.Sub(j.Started))
		}
		command := j.Command
		if command == "" {
			command = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", j.ID, j.Status, exit, age, command)
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
	Short:   "List detached jobs of the project container",
	Long: `ps lists the commands started with "dv run --detach" in the container for
the current directory, together with their status, exit code and age.

A running job is one whose process is still alive; an exited job has finished
and reports its exit code; a lost job disappeared without recording one, for
example because the container was stopped while it ran.

Only jobs started by dv are listed. Use "dv run ps aux" to see every process
in the container, and "dv logs <job>" to read a job's output.`,
	Args: cobra.NoArgs,
	RunE: ps,
}

func init() {
	rootCmd.AddCommand(psCmd)
}
