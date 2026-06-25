package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// addRegistryFlags holds every flag the `add registry` subcommand binds.
type addRegistryFlags struct {
	kindFlags
	timeoutFlags
	Ref           string
	Username      string
	Password      string
	SSHKey        string
	SkipTLSVerify bool
	CACert        string
	ClientCert    string
	ClientKey     string
}

// AddRegistry builds the `registry` subcommand of `add`.
func AddRegistry() *cobra.Command {
	var flags addRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry <name> <uri>",
		Short:         "Register a source to draw artifacts from",
		Long:          "Registry validates a source is reachable and hosts artifacts, then records it under the given name.",
		Args:          usageArgs(cobra.ExactArgs(2)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addRegistry(cmd.Context(), &flags, args, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	bindKindFlags(cmd, &flags.kindFlags)
	bindTimeoutFlags(cmd, &flags.timeoutFlags)
	set := cmd.Flags()
	set.StringVar(&flags.Ref, "ref", "", "git revision to pin (git sources only)")
	set.StringVar(&flags.Username, "username", "", "credential reference for the user, as ${env:VAR}")
	set.StringVar(&flags.Password, "password", "", "credential reference for the secret, as ${env:VAR}")
	set.StringVar(&flags.SSHKey, "ssh-key", "", "path to the SSH private key")
	set.BoolVar(&flags.SkipTLSVerify, "skip-tls-verify", false, "skip TLS certificate verification")
	set.StringVar(&flags.CACert, "ca-cert", "", "path to a CA certificate bundle")
	set.StringVar(&flags.ClientCert, "client-cert", "", "path to the client certificate")
	set.StringVar(&flags.ClientKey, "client-key", "", "path to the client private key")

	return cmd
}

// addRegistry holds the cobra-free logic: it validates flags, builds the input,
// invokes the use case through the fx graph, and writes the confirmation to
// stdout.
func addRegistry(ctx context.Context, flags *addRegistryFlags, args []string, stdout io.Writer) error {
	if err := flags.validate(); err != nil {
		return err
	}

	in := newAddRegistryInput(flags, args)

	// runUseCase runs on a cancellable run context: adapters schedule deferred
	// work (e.g. the git clone cleanup) on the worker pool keyed to it, and the
	// cancel-before-Stop teardown lets that work finish so Stop does not deadlock
	// waiting on a task that is itself waiting on the context.
	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.AddRegistryUseCase) error {
		result, err := uc.Execute(runCtx, in)
		if err != nil {
			return err
		}
		_, werr := fmt.Fprintf(stdout, "registered registry %q (%s)\n", result.Metadata.Name, result.Spec.Transport)
		return werr
	})
}

// newAddRegistryInput maps the parsed flags and positional arguments onto the
// use case's input.
func newAddRegistryInput(flags *addRegistryFlags, args []string) usecase.AddRegistryInput {
	return usecase.AddRegistryInput{
		Name:          args[0],
		URI:           args[1],
		Transport:     flags.Kind,
		Ref:           flags.Ref,
		Username:      flags.Username,
		Password:      flags.Password,
		SSHKey:        flags.SSHKey,
		SkipTLSVerify: flags.SkipTLSVerify,
		CACert:        flags.CACert,
		ClientCert:    flags.ClientCert,
		ClientKey:     flags.ClientKey,
		Timeout:       flags.Timeout,
	}
}
