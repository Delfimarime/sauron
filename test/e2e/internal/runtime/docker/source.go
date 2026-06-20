package docker

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

const (
	nginxImage    = "nginx:alpine"
	registryRoot  = "/opt/registry"         // folder sources live here inside "main"
	nginxHTMLRoot = "/usr/share/nginx/html" // webserver content root
	nginxConfPath = "/etc/nginx/conf.d/default.conf"
	htpasswdPath  = "/etc/nginx/.htpasswd"

	// webserverSecret is the concrete value the harness binds to a basic-auth
	// password declared as a ${env:VAR} reference: nginx's htpasswd is built from it
	// and the binary's container env exports it, so the round trip succeeds while the
	// stored state keeps only the reference. It is the fixed credential the auth
	// scenario asserts never leaks into state.
	webserverSecret = "s3cr3t"
)

// envRefPattern matches a ${env:VAR} credential reference (mirroring the binary's
// own grammar) so the harness can bind the referenced variable to webserverSecret.
var envRefPattern = regexp.MustCompile(`^\$\{env:([A-Za-z_][A-Za-z0-9_]*)\}$`)

// resolveAuthSecret returns the concrete password the htpasswd must hash and, when
// the declared password is a ${env:VAR} reference, the variable name to bind to that
// secret on the binary's container (so the binary resolves the reference at connect
// time). A literal password binds no variable.
func resolveAuthSecret(password string) (secret, envVar string) {
	if m := envRefPattern.FindStringSubmatch(password); m != nil {
		return webserverSecret, m[1]
	}
	return password, ""
}

// resourceSet accumulates the content a source exposes. It is embedded by each
// source type so Expose is written once.
type resourceSet struct {
	resources []runtime.Resource
}

func (s *resourceSet) Expose(resources ...runtime.Resource) {
	s.resources = append(s.resources, resources...)
}

// folderSource serves content from a directory inside the "main" service. Given
// steps accumulate resources (Expose); Path returns the deterministic in-container
// directory, valid once Start mounts the content there (the proxy forces that Start
// before any attribute access).
type folderSource struct {
	resourceSet
	alias string
}

func (s *folderSource) Path(context.Context) (string, error) { return folderPath(s.alias), nil }

func (s *folderSource) URL(context.Context) (string, error) {
	return "", fmt.Errorf("docker: folder source %q has no url; use its path", s.alias)
}

func (s *folderSource) SSHKey(context.Context) (string, error) {
	return "", fmt.Errorf("docker: folder source %q has no ssh key", s.alias)
}

// webserverSource serves content over http from an nginx sidecar. URL returns the
// sidecar's deterministic in-network address, live once Start brings it up.
type webserverSource struct {
	resourceSet
	alias string
}

func (s *webserverSource) URL(context.Context) (string, error) {
	return "http://" + webserverService(s.alias), nil
}

func (s *webserverSource) Path(context.Context) (string, error) {
	return "", fmt.Errorf("docker: webserver source %q has no path; use its url", s.alias)
}

func (s *webserverSource) SSHKey(context.Context) (string, error) {
	return "", fmt.Errorf("docker: webserver source %q has no ssh key; use its url", s.alias)
}

// folderPath is the in-container directory a folder source is mounted at, inside the
// "main" service.
func folderPath(alias string) string { return registryRoot + "/" + alias }

// webserverService is the compose service name (and in-network DNS host) of a
// webserver source's nginx sidecar.
func webserverService(alias string) string { return "registry-http-" + alias }

// buildSpecs folds the accumulated sources into base (main + any option specs):
// folder sources add content mounts to "main"; webserver sources add an nginx
// sidecar; git sources add an sshd sidecar serving a seeded bare repo and mount the
// matching client key material into "main". It is pure (specs in, specs out) so it
// is unit-tested without Docker.
func buildSpecs(
	base []ContainerSpec,
	folders map[string]*folderSource,
	webservers map[string]*webserverSource,
	gits map[string]*gitSource,
) ([]ContainerSpec, error) {
	out := make([]ContainerSpec, len(base))
	copy(out, base)

	for _, alias := range sortedKeys(folders) {
		updated, err := mountIntoMain(out, folderMounts(folders[alias]))
		if err != nil {
			return nil, err
		}
		out = updated
	}
	for _, alias := range sortedKeys(webservers) {
		out = append(out, webserverSpec(webservers[alias]))
		if envVar, secret, ok := webserverAuthEnv(webservers[alias]); ok {
			updated, err := setMainEnv(out, envVar, secret)
			if err != nil {
				return nil, err
			}
			out = updated
		}
	}
	for _, alias := range sortedKeys(gits) {
		updated, err := mountIntoMain(out, gitClientMounts(gits[alias]))
		if err != nil {
			return nil, err
		}
		out = append(updated, gitServerSpec(gits[alias]))
	}
	return out, nil
}

// sortedKeys returns a map's keys in lexical order, so buildSpecs is deterministic
// regardless of map iteration order.
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// mountIntoMain appends mounts to the reserved "main" service, replacing its Mount
// slice so the input base is not mutated.
func mountIntoMain(specs []ContainerSpec, mounts []FileSpec) ([]ContainerSpec, error) {
	for i, s := range specs {
		if s.Service != mainService {
			continue
		}
		merged := make([]FileSpec, 0, len(s.Mount)+len(mounts))
		merged = append(merged, s.Mount...)
		merged = append(merged, mounts...)
		s.Mount = merged
		specs[i] = s
		return specs, nil
	}
	return nil, fmt.Errorf("docker: %q service not found while mounting folder content", mainService)
}

// setMainEnv binds an env var on the reserved "main" service, copying its Env map so
// the input base is not mutated (buildSpecs stays pure). It lets a source thread a
// credential the binary resolves from its own environment at connect time.
func setMainEnv(specs []ContainerSpec, key, value string) ([]ContainerSpec, error) {
	for i, s := range specs {
		if s.Service != mainService {
			continue
		}
		env := make(map[string]string, len(s.Env)+1)
		for k, v := range s.Env {
			env[k] = v
		}
		env[key] = value
		s.Env = env
		specs[i] = s
		return specs, nil
	}
	return nil, fmt.Errorf("docker: %q service not found while setting env %q", mainService, key)
}

// folderMounts turns a folder source's content resources into mounts under its
// in-container directory.
func folderMounts(src *folderSource) []FileSpec {
	base := folderPath(src.alias)
	mounts := make([]FileSpec, 0, len(src.resources))
	for _, r := range src.resources {
		if r.IsAuth() {
			continue
		}
		mounts = append(mounts, FileSpec{Content: r.Content, Path: base + "/" + r.Path})
	}
	return mounts
}

// webserverSpec builds the nginx sidecar that serves a webserver source as the
// registry REST API: GET /skills and GET /agents return the JSON listing the
// exposed content, with basic auth wired in when the source declares credentials.
func webserverSpec(src *webserverSource) ContainerSpec {
	var auth *runtime.Resource
	var content []runtime.Resource
	for i, r := range src.resources {
		if r.IsAuth() {
			auth = &src.resources[i]
			continue
		}
		content = append(content, r)
	}

	mounts := []FileSpec{
		{Content: registryListingJSON(content, ".skills/"), Path: nginxHTMLRoot + "/skills"},
		{Content: registryListingJSON(content, ".agents/"), Path: nginxHTMLRoot + "/agents"},
		{Content: []byte(nginxAPIConf(auth != nil)), Path: nginxConfPath},
	}
	if auth != nil {
		secret, _ := resolveAuthSecret(auth.Password)
		mounts = append(mounts,
			FileSpec{Content: []byte(htpasswdLine(auth.Username, secret) + "\n"), Path: htpasswdPath},
		)
	}
	return ContainerSpec{Service: webserverService(src.alias), Image: nginxImage, Mount: mounts}
}

// webserverAuthEnv returns the env binding a webserver source's basic-auth password
// reference requires on the binary's container: the referenced ${env:VAR} mapped to
// the concrete secret the htpasswd was built from. A source with no auth, or a
// literal password, contributes nothing.
func webserverAuthEnv(src *webserverSource) (envVar, secret string, ok bool) {
	for i := range src.resources {
		r := src.resources[i]
		if !r.IsAuth() {
			continue
		}
		s, v := resolveAuthSecret(r.Password)
		if v == "" {
			continue
		}
		return v, s, true
	}
	return "", "", false
}

// registryItem is one row of the registry REST listing; the production presence
// scan reads name/version/size off it.
type registryItem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    int    `json:"size"`
}

// registryListing is the REST envelope the production client decodes from
// GET /skills and GET /agents.
type registryListing struct {
	Items []registryItem `json:"items"`
}

// registryListingJSON renders the REST listing for one artifact kind by collecting
// the distinct artifact names exposed under prefix (".skills/" or ".agents/"). It is
// pure (resources + prefix in, JSON out) so it is unit-tested without Docker.
func registryListingJSON(content []runtime.Resource, prefix string) []byte {
	seen := map[string]int{}
	var names []string
	for _, r := range content {
		rest, ok := strings.CutPrefix(r.Path, prefix)
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(rest, "/")
		if name == "" {
			continue
		}
		if _, dup := seen[name]; !dup {
			names = append(names, name)
		}
		seen[name] += len(r.Content)
	}
	sort.Strings(names)

	listing := registryListing{Items: make([]registryItem, 0, len(names))}
	for _, name := range names {
		listing.Items = append(listing.Items, registryItem{Name: name, Version: "1.0.0", Size: seen[name]})
	}
	out, _ := json.Marshal(listing) // a fixed-shape struct never fails to marshal
	return out
}

// htpasswdLine renders an htpasswd entry using nginx's supported {SHA} scheme
// (base64(sha1(password))), so no external hashing tool or dependency is needed.
func htpasswdLine(username, password string) string {
	sum := sha1.Sum([]byte(password))
	return username + ":{SHA}" + base64.StdEncoding.EncodeToString(sum[:])
}

// nginxAPIConf is the server block that exposes /skills and /agents as JSON. It
// returns each generated file with an application/json content type and, when auth
// is set, gates every path behind basic auth. The query string (the client's
// limit=1 presence scan) does not affect which file is served.
func nginxAPIConf(withAuth bool) string {
	lines := []string{
		"server {",
		"    listen 80;",
		"    root " + nginxHTMLRoot + ";",
		"    default_type application/json;",
		"    location / {",
	}
	if withAuth {
		lines = append(lines,
			"        auth_basic \"registry\";",
			"        auth_basic_user_file "+htpasswdPath+";",
		)
	}
	lines = append(lines,
		"    }",
		"}",
		"",
	)
	return strings.Join(lines, "\n")
}
