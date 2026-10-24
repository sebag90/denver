package cmd

import (
	"os"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

func runInContainer(cmd *cobra.Command, args []string) error {
	client, socket, err := connect()
	if err != nil {
		return err
	}

	id, err := container.EnsureContainer(client, imageFlag, sandboxFlag)
	if err != nil {
		return err
	}

	code, err := container.Exec(client, socket, id, args, false)
	if err != nil {
		return err
	}
	if code != 0 {
		os.Exit(code)
	}
	return nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <command> [args...]",
	Short: "Run a command inside the project container",
	Long: `run executes a command inside the container for the current directory,
starting the container first if needed.

stdin, stdout and stderr are forwarded and the command's exit code is returned.
Flags after the command name belong to the command, so "dv run ls -la"
works without a "--" separator.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInContainer,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Everything after the command name is an argument of that command.
	runCmd.Flags().SetInterspersed(false)
}
