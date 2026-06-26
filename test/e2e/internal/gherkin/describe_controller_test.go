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
		"credentials:\n" +
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
	_, ok := descriptorValue("name: acme\n", "revision")
	assert.False(t, ok)
}

func TestBuildRegistryStreamCarriesCredentialColumns(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"name", "transport", "source", "username", "password"},
		[]string{"acme", "git", "git@github.com:acme/artifacts.git", "${env:ACME_USER}", "${env:ACME_TOKEN}"},
	))
	require.NoError(t, err)

	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)
	require.NotNil(t, regs[0].Spec.Credentials)
	assert.Equal(t, "${env:ACME_USER}", regs[0].Spec.Credentials.Username)
	assert.Equal(t, "${env:ACME_TOKEN}", regs[0].Spec.Credentials.Password)
}

func TestBuildRegistryStreamLeavesCredentialsNilWhenAbsent(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"name", "transport", "source"},
		[]string{"acme", "git", "git@github.com:acme/artifacts.git"},
	))
	require.NoError(t, err)

	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)
	var nilCredentials *types.Credentials
	assert.Equal(t, nilCredentials, regs[0].Spec.Credentials)
}
