package cmd

import "github.com/spf13/cobra"

// List builds the `list` command group and attaches its subcommands.
func List() *cobra.Command {
	return newCommand("list", "List what is available to Sauron",
		withLong("List shows what is available to Sauron, such as the registry's catalogue."),
		withSubcommands(Catalogue()),
	)
}
