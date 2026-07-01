package cmd

import (
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// renderSetRegistry writes the confirmation that the registry was configured.
func renderSetRegistry(w io.Writer, result *usecase.SetRegistryResponse) error {
	ew := newErrWriter(w)
	ew.printf("registry set to %s (%s)\n", result.Source, result.Transport)
	return ew.toIOError("write report")
}
