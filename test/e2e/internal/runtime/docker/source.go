package docker

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
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
)

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

// folderPath is the in-container directory a folder source is mounted at, inside the
// "main" service.
func folderPath(alias string) string { return registryRoot + "/" + alias }

// webserverService is the compose service name (and in-network DNS host) of a
// webserver source's nginx sidecar.
func webserverService(alias string) string { return "registry-http-" + alias }

// buildSpecs folds the accumulated sources into base (main + any option specs):
// folder sources add content mounts to "main"; webserver sources add an nginx
// sidecar. It is pure (specs in, specs out) so it is unit-tested without Docker.
func buildSpecs(
	base []ContainerSpec,
	folders map[string]*folderSource,
	webservers map[string]*webserverSource,
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

// webserverSpec builds the nginx sidecar serving a webserver source's content, with
// basic auth wired in when the source declares credentials.
func webserverSpec(src *webserverSource) ContainerSpec {
	mounts := make([]FileSpec, 0, len(src.resources)+2)
	var auth *runtime.Resource
	for i, r := range src.resources {
		if r.IsAuth() {
			auth = &src.resources[i]
			continue
		}
		mounts = append(mounts, FileSpec{Content: r.Content, Path: nginxHTMLRoot + "/" + r.Path})
	}
	if auth != nil {
		mounts = append(mounts,
			FileSpec{Content: []byte(htpasswdLine(auth.Username, auth.Password) + "\n"), Path: htpasswdPath},
			FileSpec{Content: []byte(nginxAuthConf()), Path: nginxConfPath},
		)
	}
	return ContainerSpec{Service: webserverService(src.alias), Image: nginxImage, Mount: mounts}
}

// htpasswdLine renders an htpasswd entry using nginx's supported {SHA} scheme
// (base64(sha1(password))), so no external hashing tool or dependency is needed.
func htpasswdLine(username, password string) string {
	sum := sha1.Sum([]byte(password))
	return username + ":{SHA}" + base64.StdEncoding.EncodeToString(sum[:])
}

// nginxAuthConf is a minimal server block that serves the content root behind basic
// auth, valid enough for nginx to come up cleanly.
func nginxAuthConf() string {
	return strings.Join([]string{
		"server {",
		"    listen 80;",
		"    root " + nginxHTMLRoot + ";",
		"    location / {",
		"        auth_basic \"registry\";",
		"        auth_basic_user_file " + htpasswdPath + ";",
		"    }",
		"}",
		"",
	}, "\n")
}
