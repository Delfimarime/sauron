package cmd

import "github.com/spf13/cobra"

// List builds the `list` command group and attaches its subcommands.
func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List what is registered with Sauron",
		Long:          "List shows the sources registered with Sauron, such as registries.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(ListRegistries())
	return cmd
}
