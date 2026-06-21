//go:build unit

package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

func TestDescriptorValueReadsLabelledLine(t *testing.T) {
	output := "name:       acme\n" +
		"transport:  git\n" +
		"auth:\n" +
		"  username: ${env:ACME_USER}\n" +
		"  password: ${env:ACME_TOKEN}\n"

	got, ok := descriptorValue(output, "name")
	require.True(t, ok)
	assert.Equal(t, "acme", got)

	got, ok = descriptorValue(output, "transport")
	require.True(t, ok)
	assert.Equal(t, "git", got)

	got, ok = descriptorValue(output, "username")
	require.True(t, ok)
	assert.Equal(t, "${env:ACME_USER}", got, "a nested, indented field is read by its trimmed label")
}

func TestDescriptorValueMissingLabel(t *testing.T) {
	_, ok := descriptorValue("name: acme\n", "ref")
	assert.False(t, ok)
}

func TestBuildRegistryStreamCarriesAuthColumns(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"name", "transport", "uri", "username", "password"},
		[]string{"acme", "git", "git@github.com:acme/artifacts.git", "${env:ACME_USER}", "${env:ACME_TOKEN}"},
	))
	require.NoError(t, err)

	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)
	require.NotNil(t, regs[0].Spec.Auth)
	assert.Equal(t, "${env:ACME_USER}", regs[0].Spec.Auth.Username)
	assert.Equal(t, "${env:ACME_TOKEN}", regs[0].Spec.Auth.Password)
}

func TestBuildRegistryStreamLeavesAuthNilWhenAbsent(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"name", "transport", "uri"},
		[]string{"acme", "git", "git@github.com:acme/artifacts.git"},
	))
	require.NoError(t, err)

	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)
	var nilAuth *types.Auth
	assert.Equal(t, nilAuth, regs[0].Spec.Auth)
}
