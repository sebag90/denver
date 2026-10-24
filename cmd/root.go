package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

// imageFlag is shared by every command that can create a container.
var imageFlag string

// sandboxFlag starts containers without bind-mounting the working directory.
var sandboxFlag bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dv",
	Short: "Per-directory development containers on top of Podman",
	Long: `dv manages one container per working directory.

The container's name is derived from the current directory, and its main
process is an idle keep-alive, so it stays up after you leave a shell. Run
"dv shell" to open a new session at any time.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "dv: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&imageFlag, "image", container.DefaultImage,
		"image used when creating a new project container")
	rootCmd.PersistentFlags().BoolVarP(&sandboxFlag, "sandbox", "s", false,
		"start the container without bind-mounting the current directory")
}

// connect opens the daemon connection and negotiates the API version, so that
// an older Docker daemon is not asked for an API version newer than it speaks.
// The socket path is returned for the callers that need it for hijacked
// sessions.
func connect() (*http.Client, string, error) {
	socket := container.SocketPath()
	client := container.NewClient(socket)
	if err := container.NegotiateApiVersion(client); err != nil {
		return nil, "", err
	}
	return client, socket, nil
}
