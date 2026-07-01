package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// setRegistryFlags holds every flag the `set registry` subcommand binds.
type setRegistryFlags struct {
	transportFlags
	timeoutFlags
	Revision      string
	Username      string
	Password      string
	SSHKey        string
	SkipTLSVerify bool
	CACert        string
	ClientCert    string
	ClientKey     string
}

// SetRegistry builds the `registry` subcommand of `set`.
func SetRegistry() *cobra.Command {
	var flags setRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry <uri>",
		Short:         "Configure the source to draw artifacts from",
		Long:          "Registry validates a source is reachable and hosts artifacts, then records it as the single registry, replacing any already set.",
		Args:          usageArgs(cobra.ExactArgs(1)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setRegistry(cmd.Context(), &flags, args, cmd.OutOrStdout())
		},
	}
	silenceFlagErrors(cmd)
	bindTransportFlags(cmd, &flags.transportFlags)
	bindTimeoutFlags(cmd, &flags.timeoutFlags)
	set := cmd.Flags()
	set.StringVar(&flags.Revision, "revision", "", "git revision to pin (git sources only)")
	set.StringVar(&flags.Username, "username", "", "credential reference for the user, as ${env:VAR}")
	set.StringVar(&flags.Password, "password", "", "credential reference for the secret, as ${env:VAR}")
	set.StringVar(&flags.SSHKey, "ssh-key", "", "path to the SSH private key")
	set.BoolVar(&flags.SkipTLSVerify, "skip-tls-verify", false, "skip TLS certificate verification")
	set.StringVar(&flags.CACert, "ca-cert", "", "path to a CA certificate bundle")
	set.StringVar(&flags.ClientCert, "client-cert", "", "path to the client certificate")
	set.StringVar(&flags.ClientKey, "client-key", "", "path to the client private key")

	return cmd
}

// setRegistry holds the cobra-free logic: it validates flags, builds the input,
// lets the fx graph invoke the use case, and renders the returned result,
// returning any classified failure to the caller.
func setRegistry(ctx context.Context, flags *setRegistryFlags, args []string, stdout io.Writer) error {
	if err := flags.validate(); err != nil {
		return err
	}

	result, err := runUseCase(ctx, func(runCtx context.Context, uc *usecase.SetRegistryUseCase) (*usecase.SetRegistryResponse, error) {
		return uc.Execute(runCtx, usecase.SetRegistryRequest{
			Source:        args[0],
			Transport:     flags.Transport,
			Revision:      flags.Revision,
			Username:      flags.Username,
			Password:      flags.Password,
			SSHKey:        flags.SSHKey,
			SkipTLSVerify: flags.SkipTLSVerify,
			CACert:        flags.CACert,
			ClientCert:    flags.ClientCert,
			ClientKey:     flags.ClientKey,
			Timeout:       flags.Timeout,
		})
	})
	if err != nil {
		return err
	}

	return renderSetRegistry(stdout, result)
}
