package cmd

import (
	"fmt"
	"net/http"

	"github.com/sebag90/paguro/internal/container"
	"github.com/spf13/cobra"
)

// downAllFlag makes down act on the containers of every directory.
var downAllFlag bool

func down(cmd *cobra.Command, args []string) error {
	client, _, err := connect()
	if err != nil {
		return err
	}

	if downAllFlag {
		return downAll(client)
	}

	name := container.ContainerName()
	removed, err := container.RemoveContainer(client, name)
	if err != nil {
		return err
	}
	if removed {
		fmt.Printf("%s removed\n", name)
	} else {
		fmt.Printf("no container for this directory\n")
	}
	return nil
}

// downAll stops and removes the containers of every directory.
func downAll(client *http.Client) error {
	names, err := container.RemoveAllContainers(client)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Printf("no containers\n")
		return nil
	}
	for _, name := range names {
		fmt.Printf("%s removed\n", name)
	}
	return nil
}

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove the container for the current directory",
	Long: `down stops and removes the container that belongs to the current directory.

With --all it stops and removes every container paguro created, whatever
directory it belongs to.`,
	Args: cobra.NoArgs,
	RunE: down,
}

func init() {
	rootCmd.AddCommand(downCmd)

	downCmd.Flags().BoolVar(&downAllFlag, "all", false,
		"stop and remove the containers of every directory")
}
