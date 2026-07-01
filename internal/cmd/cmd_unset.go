package cmd

import "github.com/spf13/cobra"

// Unset builds the `unset` command group and attaches its subcommands.
func Unset() *cobra.Command {
	return newCommand("unset", "Remove a configured resource",
		withLong("Unset removes a configured resource, such as the registry, from Sauron. Installed artifacts are preserved."),
		withSubcommands(
			UnsetRegistry(),
		),
	)
}
