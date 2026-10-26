package cmd

import (
	"errors"
	"fmt"
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

	c, err := container.FindContainer(client, container.ContainerName())
	if err != nil {
		return err
	}
	if c == nil || !c.Running() {
		return errors.New("no running container for this directory")
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

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <job>",
	Short: "Print the output of a detached job",
	Long: `logs prints everything a job started with "dv run --detach" has written.

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
