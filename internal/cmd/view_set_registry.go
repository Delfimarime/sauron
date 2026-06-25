package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// renderSetRegistry writes the confirmation that the registry was configured.
func renderSetRegistry(w io.Writer, result *usecase.SetRegistryResult) error {
	if _, err := fmt.Fprintf(w, "registry set to %s (%s)\n", result.URI, result.Transport); err != nil {
		return usecase.NewIOError(fmt.Sprintf("write report: %v", err))
	}
	return nil
}
