package cmd

import "github.com/spf13/cobra"

// Set builds the `set` command group and attaches its subcommands.
func Set() *cobra.Command {
	return newCommand("set", "Configure a resource",
		withLong("Set configures a resource, such as the registry, with Sauron."),
		withSubcommands(SetRegistry(), SetProvider()),
	)
}
