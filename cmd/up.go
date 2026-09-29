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

func up(cmd *cobra.Command, args []string) {
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
	container.CreateContainer(client, "ghcr.io/sebag90/devenv")
}

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: up,
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
