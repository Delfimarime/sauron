package cmd

import "github.com/spf13/cobra"

// Delete builds the `delete` command group and attaches its subcommands.
func Delete() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "delete",
		Short:         "Remove a source from the catalogue",
		Long:          "Delete unregisters a source, such as a registry, from Sauron and cascade-uninstalls its artifacts.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(DeleteRegistry())
	return cmd
}
