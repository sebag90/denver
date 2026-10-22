/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

func ssh(cmd *cobra.Command, args []string) {
	// socket := "/run/podman/podman.sock" // + "/podman/podman.sock"
	socket := os.Getenv("XDG_RUNTIME_DIR") + "/podman/podman.sock"
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.DialUnix("unix", nil, &net.UnixAddr{
					Net:  "unix",
					Name: socket,
				})
			},
		},
		Timeout: 10 * time.Second,
	}

	name := container.NameContainer()
	containerId := container.ListContainers(client, name)
	container.AttachToContainerHijack(socket, containerId)
}

// upCmd represents the up command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: ssh,
}

func init() {
	rootCmd.AddCommand(sshCmd)
}
