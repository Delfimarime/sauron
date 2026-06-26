package cmd

import "github.com/spf13/cobra"

// Set builds the `set` command group and attaches its subcommands.
func Set() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "set",
		Short:         "Configure a resource",
		Long:          "Set configures a resource, such as the registry, with Sauron.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(SetRegistry())
	cmd.AddCommand(SetProvider())
	return cmd
}
