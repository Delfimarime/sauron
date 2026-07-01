package cmd

import "github.com/spf13/cobra"

// List builds the `list` command group and attaches its subcommands.
func List() *cobra.Command {
	return newGroup(
		"list",
		"List what is available to Sauron",
		"List shows what is available to Sauron, such as the registry's catalogue.",
		Catalogue(),
	)
}
