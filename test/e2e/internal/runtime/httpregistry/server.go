// Package httpregistry is the e2e harness's http-transport fixture: an in-process
// implementation of the Sauron HTTP Registry API (spec/contracts/registry-http-api).
// It replaces the former nginx sidecar so the registry "server" is ordinary Go in the
// test process — giving faithful server-side paging/search/sort and full per-scenario
// control — reachable by the binary under test in both runtimes (127.0.0.1 on the
// host runtime; host.docker.internal on the docker runtime).
package httpregistry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/delfimarime/sauron/pkg/sauron/marketplace"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// Secret is the concrete password the fixture binds to a ${env:VAR} basic-auth
// reference: the server checks against it and the binary's environment exports it, so
// the round trip succeeds while the binary's stored state keeps only the reference
// (the auth scenario asserts the secret never leaks into state).
const Secret = "s3cr3t"

// Listing defaults mirror the registry-http-api contract: limit defaults to 50 and is
// capped at 200; a fixed opaque version is reported since no scenario asserts it.
const (
	defaultLimit = 50
	maxLimit     = 200
	fixedVersion = "1.0.0"
)

// envRefPattern matches a ${env:VAR} basic-auth password reference, mirroring the
// binary's own grammar, so the fixture binds the referenced variable to Secret.
var envRefPattern = regexp.MustCompile(`^\$\{env:([A-Za-z_][A-Za-z0-9_]*)\}$`)

// Server is an in-process implementation of the Sauron HTTP Registry API, used as the
// http fixture for every http scenario in both runtimes. It honors the listing
// contract (q/sort/limit/offset) faithfully so paging, search, and sort scenarios
// assert real page slices and labels. Its in-memory state is driven by the existing
// content steps (hosts a skill / the directory / the file / requires basic auth).
type Server struct {
	mu        sync.Mutex
	resources []runtime.Resource
	username  string
	password  string
	listener  net.Listener
	httpSrv   *http.Server
}

// New builds an unstarted Server.
func New() *Server { return &Server{} }

// Expose accumulates content resources and, for an auth resource, the basic-auth
// credentials the server enforces. Accumulation only; the artifact set is derived per
// request, so steps may keep exposing until the first need.
func (s *Server) Expose(resources ...runtime.Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range resources {
		if r.IsAuth() {
			s.username = r.Username
			s.password = r.Password
			continue
		}
		s.resources = append(s.resources, r)
	}
}

// Start binds 0.0.0.0:0 — not 127.0.0.1, so a container reaches the server through the
// host gateway — and serves the API. Repeated calls are a no-op; the listener backlog
// accepts connections the instant it binds, so the server is reachable as soon as
// Start returns (no separate readiness wait is needed).
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return nil
	}
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("http registry fixture: listen: %w", err)
	}
	s.listener = ln
	s.httpSrv = &http.Server{Handler: s.handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = s.httpSrv.Serve(ln) }()
	return nil
}

// Port is the bound TCP port; valid only after Start.
func (s *Server) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener.Addr().(*net.TCPAddr).Port
}

// Stop shuts the server down; a server that never started is a no-op.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	srv := s.httpSrv
	s.mu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

// AuthEnv returns the environment binding the binary needs to resolve a ${env:VAR}
// basic-auth password reference to the concrete Secret the server checks. A server
// with no auth, or a literal password, contributes nothing.
func (s *Server) AuthEnv() (envVar, secret string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m := envRefPattern.FindStringSubmatch(s.password); m != nil {
		return m[1], Secret, true
	}
	return "", "", false
}

// handler wires the registry endpoints: a listing per kind and a single-artifact
// (detail and content) handler under it.
func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	for _, kind := range []string{"skills", "agents"} {
		k := kind
		mux.HandleFunc("/"+k, s.guard(func(w http.ResponseWriter, r *http.Request) { s.list(w, r, k) }))
		mux.HandleFunc("/"+k+"/", s.guard(func(w http.ResponseWriter, r *http.Request) { s.artifact(w, r, k) }))
	}
	return mux
}

// guard enforces basic auth when the fixture declares credentials; a public fixture
// passes every request through.
func (s *Server) guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		user, want := s.username, s.expectedSecret()
		s.mu.Unlock()
		if user != "" && !validBasicAuth(r, user, want) {
			w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
			writeProblem(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r)
	}
}

// expectedSecret resolves the declared password to the concrete secret the server
// checks: a ${env:VAR} reference resolves to Secret; a literal is itself.
func (s *Server) expectedSecret() string {
	if envRefPattern.MatchString(s.password) {
		return Secret
	}
	return s.password
}

// list answers GET /{kind} with the kind's artifacts after applying q/sort/limit/
// offset, as the marketplace client decodes.
func (s *Server) list(w http.ResponseWriter, r *http.Request, kind string) {
	page := applyQuery(s.summaries(prefixOf(kind)), r.URL.Query())
	writeJSON(w, marketplace.ArtifactList{Items: page})
}

// artifact answers GET /{kind}/{name} (metadata) and GET /{kind}/{name}/content (a
// gzip archive of the artifact's file tree), each carrying the Artifact-Version
// header. These endpoints are contract-complete though the catalogue scenarios only
// exercise the listing.
func (s *Server) artifact(w http.ResponseWriter, r *http.Request, kind string) {
	name, content := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/"+kind+"/"), "/content")
	files := s.artifactFiles(prefixOf(kind), name)
	if name == "" || len(files) == 0 {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("no artifact named %q", name))
		return
	}
	w.Header().Set("Artifact-Version", fixedVersion)
	if content {
		writeArchive(w, files)
		return
	}
	version := fixedVersion
	size := totalSize(files)
	writeJSON(w, marketplace.ArtifactSummary{Name: name, Version: &version, Size: &size})
}

// summaries derives the distinct artifacts under prefix ("skills/" or "agents/"),
// one per first path segment, sized by the bytes exposed beneath it.
func (s *Server) summaries(prefix string) []marketplace.ArtifactSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	sizes := map[string]int64{}
	var names []string
	for _, res := range s.resources {
		rest, ok := strings.CutPrefix(res.Path, prefix)
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(rest, "/")
		if name == "" {
			continue
		}
		if _, seen := sizes[name]; !seen {
			names = append(names, name)
		}
		sizes[name] += int64(len(res.Content))
	}
	sort.Strings(names)
	out := make([]marketplace.ArtifactSummary, 0, len(names))
	for _, name := range names {
		version := fixedVersion
		size := sizes[name]
		out = append(out, marketplace.ArtifactSummary{Name: name, Version: &version, Size: &size})
	}
	return out
}

// artifactFiles returns the resources beneath a single artifact's directory.
func (s *Server) artifactFiles(prefix, name string) []runtime.Resource {
	s.mu.Lock()
	defer s.mu.Unlock()
	base := prefix + name + "/"
	var out []runtime.Resource
	for _, res := range s.resources {
		if strings.HasPrefix(res.Path, base) {
			out = append(out, res)
		}
	}
	return out
}

// applyQuery applies the registry-http-api listing parameters to items in contract
// order: filter by q, sort by ±name, then offset/limit.
func applyQuery(items []marketplace.ArtifactSummary, q url.Values) []marketplace.ArtifactSummary {
	out := filterByName(items, q.Get("q"))
	sortByName(out, strings.HasPrefix(q.Get("sort"), "-"))
	return paginate(out, atoiDefault(q.Get("offset"), 0), limitOf(q.Get("limit")))
}

// filterByName keeps items whose name contains term (case-insensitive); an empty term
// matches everything. It never mutates the input.
func filterByName(items []marketplace.ArtifactSummary, term string) []marketplace.ArtifactSummary {
	if term == "" {
		return append([]marketplace.ArtifactSummary(nil), items...)
	}
	needle := strings.ToLower(term)
	out := make([]marketplace.ArtifactSummary, 0, len(items))
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Name), needle) {
			out = append(out, it)
		}
	}
	return out
}

// sortByName orders items by name, descending when desc is set.
func sortByName(items []marketplace.ArtifactSummary, desc bool) {
	sort.SliceStable(items, func(i, j int) bool {
		if desc {
			return items[i].Name > items[j].Name
		}
		return items[i].Name < items[j].Name
	})
}

// paginate applies offset then limit, returning an empty (non-nil) page past the end.
func paginate(items []marketplace.ArtifactSummary, offset, limit int) []marketplace.ArtifactSummary {
	if offset >= len(items) {
		return items[:0]
	}
	items = items[offset:]
	if limit >= 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}

// limitOf parses the limit parameter, defaulting to 50 and clamping to [1, 200].
func limitOf(raw string) int {
	if raw == "" {
		return defaultLimit
	}
	limit := atoiDefault(raw, defaultLimit)
	if limit < 1 {
		return 1
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

// atoiDefault parses raw as an int, falling back to fallback on empty or invalid input.
func atoiDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

// totalSize sums the byte length of the files.
func totalSize(files []runtime.Resource) int64 {
	var size int64
	for _, f := range files {
		size += int64(len(f.Content))
	}
	return size
}

// validBasicAuth reports whether the request carries the expected credentials,
// compared in constant time.
func validBasicAuth(r *http.Request, user, secret string) bool {
	u, p, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(u), []byte(user)) == 1 &&
		subtle.ConstantTimeCompare([]byte(p), []byte(secret)) == 1
}

// writeArchive streams the files as a gzip-compressed tar archive.
func writeArchive(w http.ResponseWriter, files []runtime.Resource) {
	w.Header().Set("Content-Type", "application/gzip")
	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)
	for _, f := range files {
		_ = tw.WriteHeader(&tar.Header{Name: f.Path, Mode: 0o644, Size: int64(len(f.Content))})
		_, _ = tw.Write(f.Content)
	}
	_ = tw.Close()
	_ = gz.Close()
}

// writeJSON encodes body as an application/json response.
func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

// writeProblem encodes an application/problem+json error response.
func writeProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"title":  http.StatusText(status),
		"detail": detail,
	})
}

// prefixOf is the content-set prefix for an artifact kind.
func prefixOf(kind string) string { return kind + "/" }
