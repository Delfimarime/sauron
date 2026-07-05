//go:build unit

package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRowMatchesOnFields(t *testing.T) {
	output := "NAME        KIND\n" +
		"go-style    skill\n" +
		"sql-review  skill\n" +
		"showing 1–2 (page 1, limit 20)\n"

	assert.True(t, hasRow(output, "go-style skill"))
	assert.True(t, hasRow(output, "sql-review skill"))
	assert.False(t, hasRow(output, "go-style agent"))
	assert.False(t, hasRow(output, "missing skill"))
}

func TestHasLineMatchesTrimmedLine(t *testing.T) {
	output := "NAME  KIND\ncode-reviewer  agent\nshowing 1–1 (page 1, limit 20)\n"

	assert.True(t, hasLine(output, "showing 1–1 (page 1, limit 20)"))
	assert.False(t, hasLine(output, "showing 0 results (page 1, limit 20)"))
}
