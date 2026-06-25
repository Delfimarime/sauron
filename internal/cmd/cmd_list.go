package cmd

import "github.com/spf13/cobra"

// List builds the `list` command group and attaches its subcommands.
func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List what is available to Sauron",
		Long:          "List shows what is available to Sauron, such as the registry's catalogue.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(Catalogue())
	return cmd
}
