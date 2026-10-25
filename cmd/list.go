package cmd

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"text/tabwriter"

	"github.com/sebag90/denver/internal/container"
	"github.com/spf13/cobra"
)

func list(cmd *cobra.Command, args []string) error {
	client, _, err := connect()
	if err != nil {
		return err
	}

	containers, err := container.ListManagedContainers(client)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		fmt.Printf("no containers\n")
		return nil
	}

	fmt.Print(listRows(containers))
	return nil
}

// listRows renders the containers as an aligned table sorted by directory and
// name. Containers created before the directory label existed show "-" instead
// of a directory.
func listRows(containers []container.Container) string {
	sorted := slices.Clone(containers)
	sort.Slice(sorted, func(i, j int) bool {
		di, dj := sorted[i].Directory(), sorted[j].Directory()
		if di != dj {
			return di < dj
		}
		return sorted[i].Name() < sorted[j].Name()
	})

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDIRECTORY\tSTATUS\tSANDBOX")
	for _, c := range sorted {
		dir := c.Directory()
		if dir == "" {
			dir = "-"
		}
		status := c.State
		if status == "" {
			status = c.Status
		}
		if status == "" {
			status = "-"
		}
		sandbox := "no"
		if c.Sandbox() {
			sandbox = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name(), dir, status, sandbox)
	}
	w.Flush()
	return buf.String()
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the containers for all directories",
	Long: `list shows every container dv created, with the directory it belongs to and
whether it was started in sandbox mode.

Containers created by versions of dv that did not record the directory show
"-" instead.`,
	Args: cobra.NoArgs,
	RunE: list,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
