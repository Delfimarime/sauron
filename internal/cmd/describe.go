package cmd

import "github.com/spf13/cobra"

// Describe builds the `describe` command group and attaches its subcommands.
func Describe() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "describe",
		Short:         "Show the full detail of one registered resource",
		Long:          "Describe shows a single resource's detail, such as one registry, in a vertical key-value view.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(DescribeRegistry())
	return cmd
}
