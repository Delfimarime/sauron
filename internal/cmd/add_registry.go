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

// addRegistry holds the cobra-free logic: it validates flags, builds the
// request, and lets the fx graph invoke the use case, returning the classified
// failure to the caller.
func addRegistry(ctx context.Context, flags *addRegistryFlags, args []string, stdout io.Writer) error {
	if err := flags.validate(); err != nil {
		return err
	}

	// runUseCase runs on a cancellable run context: adapters schedule deferred
	// work (e.g. the git clone cleanup) on the worker pool keyed to it, and the
	// cancel-before-Stop teardown lets that work finish so Stop does not deadlock
	// waiting on a task that is itself waiting on the context.
	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.AddRegistryUseCase) error {
		return uc.Execute(newAddRegistryRequest(runCtx, flags, args, stdout))
	})
}

// newAddRegistryRequest maps the parsed flags and positional arguments onto the
// use case's request, binding it to ctx and the command's output writer.
func newAddRegistryRequest(ctx context.Context, flags *addRegistryFlags, args []string, stdout io.Writer) *usecase.AddRegistryRequest {
	request := usecase.NewAddRegistryRequest(ctx, stdout)
	request.Name = args[0]
	request.URI = args[1]
	request.Transport = flags.Kind
	request.Ref = flags.Ref
	request.Username = flags.Username
	request.Password = flags.Password
	request.SSHKey = flags.SSHKey
	request.SkipTLSVerify = flags.SkipTLSVerify
	request.CACert = flags.CACert
	request.ClientCert = flags.ClientCert
	request.ClientKey = flags.ClientKey
	request.Timeout = flags.Timeout
	return request
}
