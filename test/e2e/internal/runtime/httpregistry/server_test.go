//go:build unit

package httpregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/marketplace"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// skill builds a content resource for one skill artifact, named by its directory.
func skill(name string) runtime.Resource {
	return runtime.Resource{Path: "skills/" + name + "/skill.yaml", Content: []byte("name: " + name)}
}

// start brings up a server exposing resources and returns it, stopped on cleanup.
func start(t *testing.T, resources ...runtime.Resource) *Server {
	t.Helper()
	s := New()
	s.Expose(resources...)
	require.NoError(t, s.Start())
	t.Cleanup(func() { _ = s.Stop(context.Background()) })
	return s
}

// listNames issues GET path against s and returns the artifact names in the page.
func listNames(t *testing.T, s *Server, path string) []string {
	t.Helper()
	list := decodeList(t, doGet(t, s, path, ""))
	names := make([]string, 0, len(list.Items))
	for _, it := range list.Items {
		names = append(names, it.Name)
	}
	return names
}

func doGet(t *testing.T, s *Server, path, _ string) *http.Response {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", s.Port(), path)) //nolint:noctx
	require.NoError(t, err)
	return resp
}

func decodeList(t *testing.T, resp *http.Response) marketplace.ArtifactList {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list marketplace.ArtifactList
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	return list
}

// TestListAppliesSortLimitOffset proves the server pages and sorts server-side, the
// behaviour the catalogue paging/sort scenarios depend on.
func TestListAppliesSortLimitOffset(t *testing.T) {
	s := start(t, skill("alpha"), skill("bravo"), skill("charlie"))

	assert.Equal(t, []string{"alpha", "bravo", "charlie"}, listNames(t, s, "/skills"))
	assert.Equal(t, []string{"charlie"}, listNames(t, s, "/skills?sort=-name&limit=1"))
	assert.Equal(t, []string{"bravo"}, listNames(t, s, "/skills?limit=1&offset=1"))
	assert.Empty(t, listNames(t, s, "/skills?offset=160&limit=20"), "paging past the end is an empty page")
}

// TestListFiltersBySubstring proves the q parameter filters case-insensitively.
func TestListFiltersBySubstring(t *testing.T) {
	s := start(t, skill("code-review"), skill("go-style"), skill("sql-review"))

	assert.Equal(t, []string{"code-review", "sql-review"}, listNames(t, s, "/skills?q=REV"))
	assert.Empty(t, listNames(t, s, "/skills?q=absent"))
}

// TestListItemShape proves the summary carries the contract fields and that agents and
// skills are separated by their content prefix.
func TestListItemShape(t *testing.T) {
	s := start(t,
		skill("go-style"),
		runtime.Resource{Path: "agents/code-reviewer/agent.yaml", Content: []byte("x")},
	)

	skills := decodeList(t, doGet(t, s, "/skills", "")).Items
	require.Len(t, skills, 1)
	require.NotNil(t, skills[0].Version)
	require.NotNil(t, skills[0].Size)
	assert.Equal(t, "go-style", skills[0].Name)
	assert.Positive(t, *skills[0].Size)

	agents := decodeList(t, doGet(t, s, "/agents", "")).Items
	require.Len(t, agents, 1)
	assert.Equal(t, "code-reviewer", agents[0].Name)
}

// TestBasicAuth proves a fixture declaring a ${env:VAR} credential rejects anonymous
// access, accepts the resolved secret, and reports the env binding the binary needs.
func TestBasicAuth(t *testing.T) {
	s := start(t, skill("go-style"), runtime.Resource{Username: "acme", Password: "${env:ACME_TOKEN}"})

	envVar, secret, ok := s.AuthEnv()
	require.True(t, ok)
	assert.Equal(t, "ACME_TOKEN", envVar)
	assert.Equal(t, Secret, secret)

	anon := doGet(t, s, "/skills", "")
	defer func() { _ = anon.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, anon.StatusCode)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		fmt.Sprintf("http://127.0.0.1:%d/skills", s.Port()), nil)
	require.NoError(t, err)
	req.SetBasicAuth("acme", Secret)
	authed, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, decodeListStatus(authed))
}

func decodeListStatus(resp *http.Response) int {
	_ = resp.Body.Close()
	return resp.StatusCode
}

// TestAuthEnvAbsentForPublicOrLiteral proves no env binding is reported without a
// ${env:VAR} reference.
func TestAuthEnvAbsentForPublicOrLiteral(t *testing.T) {
	public := New()
	_, _, ok := public.AuthEnv()
	assert.False(t, ok, "a public fixture binds no env")

	literal := New()
	literal.Expose(runtime.Resource{Username: "acme", Password: "s3cr3t"})
	_, _, ok = literal.AuthEnv()
	assert.False(t, ok, "a literal password binds no env")
}
