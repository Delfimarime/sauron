package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// Install builds the `install` command group and attaches its per-kind
// subcommands. A bare invocation prints help and exits 0; a kind noun selects
// the leaf that installs.
func Install() *cobra.Command {
	return newCommand("install", "Install artifacts from the registry",
		withLong("Install fetches and records named skills or agents from the configured registry into the active provider."),
		withSubcommands(
			InstallSkill(), InstallAgent(),
		),
	)
}

// InstallSkill builds the `skill` subcommand of `install`.
func InstallSkill() *cobra.Command {
	return newInstallCommand(
		types.KindSkill, "skill",
		"Install named skills from the registry",
		"Skill fetches each named skill from the configured registry and records it in the provider's track file.",
	)
}

// InstallAgent builds the `agent` subcommand of `install`.
func InstallAgent() *cobra.Command {
	return newInstallCommand(
		types.KindAgent, "agent",
		"Install named agents from the registry",
		"Agent fetches each named agent from the configured registry and records it in the provider's track file.",
	)
}

// newInstallCommand builds one install leaf command bound to a kind; the leaves
// differ only by the kind they delegate to install.
func newInstallCommand(kind, use, short, long string) *cobra.Command {
	return newCommand(use, short,
		withLong(long),
		withArgs(cobra.MinimumNArgs(1)),
		withRunE(func(ctx context.Context, args []string, stdout io.Writer) error {
			return install(ctx, kind, args, stdout)
		}),
	)
}

// install is the cobra-free handler shared by both install leaf commands: it
// builds the use-case input from the kind and arg names, runs the use case via
// the fx graph, renders the presentation-agnostic result, and then checks for
// persist failures. Benign skips (not-offered, no-version) do not change the
// exit code; only persist failures (Fatal == true) make the process exit 1.
func install(ctx context.Context, kind string, names []string, stdout io.Writer) error {
	result, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.InstallRequest, usecase.InstallResponse]) (*usecase.InstallResponse, error) {
		return uc.Execute(runCtx, usecase.InstallRequest{Kind: kind, Names: names})
	})
	if err != nil {
		return err
	}

	if err := renderInstall(stdout, kind, result); err != nil {
		return err
	}

	// A fatal failure means at least one artifact could not be fetched, written,
	// or recorded; exit non-zero so the caller can detect and retry.
	// Benign skips (Fatal == false: not offered or no declared version) leave the
	// exit code 0 — the run succeeded for every offered artifact.
	for _, f := range result.Failures {
		if f.Fatal {
			return usecase.NewIOError(fmt.Sprintf("install %q: %s", f.Name, f.Reason))
		}
	}

	return nil
}
