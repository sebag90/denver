package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

var logsFollowFlag bool

func logs(cmd *cobra.Command, args []string) error {
	jobID := args[0]
	if !container.ValidJobID(jobID) {
		return fmt.Errorf("invalid job ID %q", jobID)
	}

	client, socket, err := connect()
	if err != nil {
		return err
	}

	c, err := jobContainer(client, socket, jobID)
	if err != nil {
		return err
	}

	code, err := container.Exec(client, socket, c.ID, container.LogCommand(jobID, logsFollowFlag), false)
	if err != nil {
		return err
	}
	if code != 0 {
		os.Exit(code)
	}
	return nil
}

// jobContainer returns the container that holds the job. The container for the
// current directory is checked first, so the common case costs one cheap exec;
// after that every other running container dv created is searched, so a job
// listed by "dv ps --all" can be read from any directory.
//
// Job IDs are random, which dv relies on to treat them as unique across
// containers. A job in a stopped container is not found, because its metadata
// cannot be read without entering the container.
func jobContainer(client *http.Client, socket, jobID string) (*container.Container, error) {
	current, err := container.FindContainer(client, container.ContainerName())
	if err != nil {
		return nil, err
	}

	if current != nil && current.Running() {
		found, err := container.HasJob(client, socket, current.ID, jobID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", current.Name(), err)
		}
		if found {
			return current, nil
		}
	}

	containers, err := container.ListManagedContainers(client)
	if err != nil {
		return nil, err
	}
	for i := range containers {
		c := &containers[i]
		if !c.Running() || (current != nil && c.ID == current.ID) {
			continue
		}
		found, err := container.HasJob(client, socket, c.ID, jobID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", c.Name(), err)
		}
		if found {
			return c, nil
		}
	}

	return nil, fmt.Errorf("no job %s in a running container", jobID)
}

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <job>",
	Short: "Print the output of a detached job",
	Long: `logs prints everything a job started with "dv run --detach" has written.

The job is looked up in the current directory's container first and then in
every other running container dv created, so a job listed by "dv ps --all" can
be read from any directory. Jobs of stopped containers cannot be found.

With --follow it keeps printing as the job writes more, until the job ends.
The log file lives inside the container and is removed with it.`,
	Args: cobra.ExactArgs(1),
	RunE: logs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&logsFollowFlag, "follow", "f", false,
		"keep printing new output until the job ends")
}
