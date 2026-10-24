package cmd

import (
	"os"
	"path/filepath"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

var shellFlag string

func openShell(cmd *cobra.Command, args []string) error {
	client, socket, err := connect()
	if err != nil {
		return err
	}

	id, err := container.EnsureContainer(client, imageFlag, sandboxFlag)
	if err != nil {
		return err
	}

	code, err := container.Exec(client, socket, id, shellArgs(shellFlag), true)
	if err != nil {
		return err
	}
	// The terminal has already been restored by Exec; report the shell's exit
	// code so that scripts can rely on it.
	if code != 0 {
		os.Exit(code)
	}
	return nil
}

// shellArgs builds the command for the interactive session. The host's shell is
// tried first and /bin/sh is used as a fallback, since the host shell is not
// necessarily installed in the image.
func shellArgs(shell string) []string {
	return []string{
		"/bin/sh", "-c",
		`if command -v "$1" >/dev/null 2>&1; then exec "$1" -l; fi; exec /bin/sh -l`,
		"sh", shell,
	}
}

func defaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}
	return "bash"
}

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:     "shell",
	Aliases: []string{"ssh"},
	Short:   "Open an interactive shell in the project container",
	Long: `shell opens a new interactive session in the container for the current
directory, starting the container first if needed.

The session is a child of the container's keep-alive process, so exiting it
leaves the container running and you can open several sessions in parallel.`,
	Args: cobra.NoArgs,
	RunE: openShell,
}

func init() {
	rootCmd.AddCommand(shellCmd)

	// -S rather than -s, which belongs to the global --sandbox flag.
	shellCmd.Flags().StringVarP(&shellFlag, "shell", "S", defaultShell(),
		"shell to start inside the container")
}
