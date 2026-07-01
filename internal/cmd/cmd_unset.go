package cmd

import "github.com/spf13/cobra"

// Unset builds the `unset` command group and attaches its subcommands.
func Unset() *cobra.Command {
	return newGroup(
		"unset",
		"Remove a configured resource",
		"Unset removes a configured resource, such as the registry, from Sauron. Installed artifacts are preserved.",
		UnsetRegistry(),
	)
}
