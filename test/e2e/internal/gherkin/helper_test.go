//go:build unit

package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertExpected(t *testing.T) {
	assert.NoError(t, assertExpected("version", "0.0.0", "0.0.0"))
	assert.Error(t, assertExpected("version", "0.0.0", "9.9.9"))

	// Generic over any comparable type, not just string.
	assert.NoError(t, assertExpected("count", 3, 3))
	assert.Error(t, assertExpected("count", 3, 4))
}
