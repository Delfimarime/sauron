package host

import (
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

type Factory struct {
}

func (*Factory) New(binaryURI, directoryURI string) (runtime.Runtime, error) {
	return New(binaryURI, directoryURI), nil
}
