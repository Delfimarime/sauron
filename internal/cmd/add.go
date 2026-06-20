package cmd

import "github.com/spf13/cobra"

// Add builds the `add` command group and attaches its subcommands.
func Add() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "add",
		Short:         "Add a source to the catalogue",
		Long:          "Add registers a new source, such as a registry, with Sauron.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(AddRegistry())
	return cmd
}
