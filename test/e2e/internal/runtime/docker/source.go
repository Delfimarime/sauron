package docker

import (
	"context"
	"fmt"
	"maps"
	"sort"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/httpregistry"
)

const (
	registryRoot = "/opt/registry" // folder sources live here inside "main"

	// hostGatewayExtraHost lets the containerized binary reach the in-process http
	// registry fixture, which runs in the test process on the host. host-gateway is
	// required on Linux/CI (not just Docker Desktop), so it is wired explicitly onto
	// "main" whenever a webserver source is declared.
	hostGatewayExtraHost = "host.docker.internal:host-gateway"
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

func (s *folderSource) SSHKey(context.Context) (string, error) {
	return "", fmt.Errorf("docker: folder source %q has no ssh key", s.alias)
}

func (s *folderSource) Revision(context.Context) (string, error) {
	return "", fmt.Errorf("docker: folder source %q has no revision; use its path", s.alias)
}

// folderPath is the in-container directory a folder source is mounted at, inside the
// "main" service.
func folderPath(alias string) string { return registryRoot + "/" + alias }

// buildSpecs folds the accumulated sources into base (main + any option specs):
// folder sources add content mounts to "main"; webserver sources wire the host
// gateway and any basic-auth credential onto "main" (the server itself runs in the
// test process, not a container); git sources add an sshd sidecar serving a seeded
// bare repo and mount the matching client key material into "main". It is pure (specs
// in, specs out) so it is unit-tested without Docker.
func buildSpecs(
	base []ContainerSpec,
	folders map[string]*folderSource,
	webservers map[string]*httpregistry.Source,
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
	updated, err := wireWebservers(out, webservers)
	if err != nil {
		return nil, err
	}
	out = updated
	for _, alias := range sortedKeys(gits) {
		updated, err := mountIntoMain(out, gitClientMounts(gits[alias]))
		if err != nil {
			return nil, err
		}
		out = append(updated, gitServerSpec(gits[alias]))
	}
	return out, nil
}

// wireWebservers attaches the host gateway to "main" (once) and binds each
// webserver's ${env:VAR} basic-auth secret on "main", so the containerized binary can
// reach the host-run server and resolve its credential reference at connect time.
func wireWebservers(specs []ContainerSpec, webservers map[string]*httpregistry.Source) ([]ContainerSpec, error) {
	if len(webservers) == 0 {
		return specs, nil
	}
	out, err := addExtraHostToMain(specs, hostGatewayExtraHost)
	if err != nil {
		return nil, err
	}
	for _, alias := range sortedKeys(webservers) {
		envVar, secret, ok := webservers[alias].Server().AuthEnv()
		if !ok {
			continue
		}
		out, err = setMainEnv(out, envVar, secret)
		if err != nil {
			return nil, err
		}
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
		maps.Copy(env, s.Env)
		env[key] = value
		s.Env = env
		specs[i] = s
		return specs, nil
	}
	return nil, fmt.Errorf("docker: %q service not found while setting env %q", mainService, key)
}

// addExtraHostToMain appends an extra_hosts entry to the reserved "main" service,
// copying the slice so the input base is not mutated.
func addExtraHostToMain(specs []ContainerSpec, entry string) ([]ContainerSpec, error) {
	for i, s := range specs {
		if s.Service != mainService {
			continue
		}
		hosts := make([]string, 0, len(s.ExtraHosts)+1)
		hosts = append(hosts, s.ExtraHosts...)
		hosts = append(hosts, entry)
		s.ExtraHosts = hosts
		specs[i] = s
		return specs, nil
	}
	return nil, fmt.Errorf("docker: %q service not found while adding extra host", mainService)
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
