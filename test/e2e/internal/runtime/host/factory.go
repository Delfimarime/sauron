package host

import (
	"path/filepath"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

type Factory struct {
}

func (*Factory) GetHomeDirectory() (string, error) {
	return filepath.Abs("~/.sauron")
}

func (*Factory) New(binaryURI, _ string) (runtime.Runtime, error) {
	return New(binaryURI), nil
}
