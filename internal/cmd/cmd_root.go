package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/config"
)

// New builds the root cobra command with the given build identity.
func New(appName, appVersion, appHash string) (*cobra.Command, error) {
	home, err := config.GetHomeDirectory()
	if err != nil {
		return nil, fmt.Errorf("determine home directory. caused by: %w", err)
	}
	banner := fmt.Sprintf("%s v%s\nHash %s\nHome: %s\n", appName, appVersion, appHash, home)
	root := &cobra.Command{
		Use:           appName,
		Version:       appVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprint(cmd.OutOrStdout(), banner)
			return err
		},
	}
	root.SetVersionTemplate(banner)
	root.AddCommand(
		Set(), List(), Describe(), Unset(), Install(),
	)
	return root, nil
}

// ExitCode maps the error returned by the root command's Execute to the process
// exit code the binary should terminate with.
func ExitCode(err error) int {
	return exitCode(err)
}
