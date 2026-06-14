package gherkin

import (
	"errors"
	"fmt"
)

// capturingT adapts testify's assert.TestingT to a godog step, which signals
// failure by returning an error rather than holding a *testing.T. Assertions
// record their failure here; the step returns err() so godog reports it.
type capturingT struct {
	failures []string
}

// Errorf records a formatted assertion failure. It satisfies assert.TestingT so
// the testify assert helpers can be used inside steps.
func (t *capturingT) Errorf(format string, args ...any) {
	t.failures = append(t.failures, fmt.Sprintf(format, args...))
}

// err collapses recorded assertion failures into a single error, or nil when
// every assertion held.
func (t *capturingT) err() error {
	if len(t.failures) == 0 {
		return nil
	}

	errs := make([]error, len(t.failures))
	for i, f := range t.failures {
		errs[i] = errors.New(f)
	}

	return errors.Join(errs...)
}
