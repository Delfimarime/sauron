package cmd

import "github.com/spf13/cobra"

// Describe builds the `describe` command group and attaches its subcommands.
func Describe() *cobra.Command {
	return newCommand("describe", "Show the full detail of one registered resource",
		withLong("Describe shows a single resource's detail, such as one registry, in a vertical key-value view."),
		withSubcommands(
			DescribeRegistry(), DescribeProvider(),
		),
	)
}
