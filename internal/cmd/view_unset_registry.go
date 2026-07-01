package cmd

import (
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// unsetMessages maps each removal outcome to its canonical report line.
var unsetMessages = map[usecase.UnsetOutcome]string{
	usecase.UnsetNothing: "no registry configured; nothing was unset",
	usecase.UnsetPreview: "registry would be unset; installed artifacts preserved",
	usecase.UnsetRemoved: "registry unset; installed artifacts preserved",
}

// renderUnsetRegistry writes the report line for the removal outcome.
func renderUnsetRegistry(w io.Writer, result *usecase.UnsetRegistryResponse) error {
	ew := newErrWriter(w)
	ew.printf("%s\n", unsetMessages[result.Outcome])
	return ew.toIOError("write report")
}
