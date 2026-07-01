package cmd

import "github.com/spf13/cobra"

// Set builds the `set` command group and attaches its subcommands.
func Set() *cobra.Command {
	return newGroup(
		"set",
		"Configure a resource",
		"Set configures a resource, such as the registry, with Sauron.",
		SetRegistry(), SetProvider(),
	)
}
