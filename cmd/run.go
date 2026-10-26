package cmd

import (
	"fmt"
	"os"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

var (
	detachFlag bool
	logFlag    string
)

func runInContainer(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("log") && !detachFlag {
		return fmt.Errorf("--log requires --detach")
	}

	client, socket, err := connect()
	if err != nil {
		return err
	}

	id, err := container.EnsureContainer(client, imageFlag, sandboxFlag)
	if err != nil {
		return err
	}

	if detachFlag {
		jobID, err := container.StartJob(client, id, args, logFlag)
		if err != nil {
			return err
		}
		logPath := logFlag
		if logPath == "" {
			logPath = container.JobsDir + "/" + jobID + "/log"
		}
		fmt.Printf("job %s started (log: %s)\n", jobID, logPath)
		return nil
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
works without a "--" separator.

With --detach the command is started in the background and run returns
immediately. Its output is written to a log file inside the container, and the
job appears in "dv ps" and can be followed with "dv logs -f <job>". Detached
jobs keep running while the container is up; they end when it is stopped or
removed.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInContainer,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVarP(&detachFlag, "detach", "d", false,
		"start the command in the background and return immediately")
	runCmd.Flags().StringVar(&logFlag, "log", "",
		"path inside the container for the command's output (with --detach)")

	// Everything after the command name is an argument of that command.
	runCmd.Flags().SetInterspersed(false)
}
