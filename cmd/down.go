package cmd

import (
	"fmt"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

func down(cmd *cobra.Command, args []string) error {
	client, _, err := connect()
	if err != nil {
		return err
	}
	name := container.ContainerName()

	removed, err := container.RemoveContainer(client, name)
	if err != nil {
		return err
	}
	if !removed {
		fmt.Printf("no container for this directory\n")
		return nil
	}

	fmt.Printf("%s removed\n", name)
	return nil
}

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove the container for the current directory",
	Long:  `down stops and removes the container that belongs to the current directory.`,
	Args:  cobra.NoArgs,
	RunE:  down,
}

func init() {
	rootCmd.AddCommand(downCmd)
}
