/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

func up(cmd *cobra.Command, args []string) error {
	client, _, err := connect()
	if err != nil {
		return err
	}

	id, err := container.EnsureContainer(client, imageFlag, sandboxFlag)
	if err != nil {
		return err
	}

	fmt.Printf("%s is running (%s)\n", container.ContainerName(), container.ShortID(id))
	return nil
}

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and start the container for the current directory",
	Long: `up makes sure the container for the current directory exists and is running.

It is idempotent: a container that is already running is reused, a stopped one
is started again.`,
	Args: cobra.NoArgs,
	RunE: up,
}

func init() {
	rootCmd.AddCommand(upCmd)
}
