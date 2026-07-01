// Package docker holds the e2e harness's compose-backed Runtime: every sandbox
// (single dependency or a group) is realized as a docker-compose project with a
// baked-in "main" service that runs the binary under test. Where the binary
// executes (host vs sandbox) is decided by the caller; this package owns the
// sandbox case.
package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ContainerSpec declares one service in the compose project. Given steps append
// specs to the world; the runtime always prepends the "main" service. Service is
// the compose service name.
type ContainerSpec struct {
	Service    string
	Image      string
	Entrypoint string
	Ports      []string
	Mount      []FileSpec
	Env        map[string]string
	// ExtraHosts renders as the compose service's extra_hosts (e.g.
	// "host.docker.internal:host-gateway"), so a containerized binary can reach a
	// server the test process runs on the host gateway.
	ExtraHosts []string
}

// FileSpec mounts a file into a container. Provide exactly one source: SourceFile
// to bind an existing host file, or Content to have the runtime materialize the
// bytes to a file (under the compose directory) and bind that.
type FileSpec struct {
	Path       string
	Content    []byte
	SourceFile string
}

// materializeContent writes every Content-only FileSpec to a file under directory
// (via writeFile, injected so this is unit-testable without the real fs) and sets
// its SourceFile, so GenerateDockerComposeFile sees only concrete host paths.
func materializeContent(
	writeFile func(string, []byte, os.FileMode) error,
	directory string,
	specs []ContainerSpec,
) ([]ContainerSpec, error) {
	out := make([]ContainerSpec, len(specs))
	for i, s := range specs {
		mounts := make([]FileSpec, len(s.Mount))
		for j, m := range s.Mount {
			if len(m.Content) > 0 && m.SourceFile == "" {
				name := filepath.Join(directory, fmt.Sprintf("%s-%d-%s", s.Service, j, filepath.Base(m.Path)))
				// World-readable: these files are bind-mounted and read by
				// unprivileged container processes (e.g. nginx's worker, uid 101).
				// On Linux the host mode is preserved, so 0o600 would 403/500 nginx;
				// macOS Docker's FUSE masks it, hiding the bug locally.
				if err := writeFile(name, m.Content, 0o644); err != nil {
					return nil, fmt.Errorf("materialize content for %q: %w", s.Service, err)
				}
				m.SourceFile = name
			}
			mounts[j] = m
		}
		s.Mount = mounts
		out[i] = s
	}
	return out, nil
}

// GenerateDockerComposeFile renders specs into a docker-compose.yml document. It
// is pure (specs -> bytes) so it is unit-tested without Docker; every mount must
// already carry a concrete SourceFile (materializeContent handles inline Content).
func GenerateDockerComposeFile(specs []ContainerSpec) ([]byte, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("compose: no container specs")
	}

	services := make(map[string]any, len(specs))
	for _, s := range specs {
		if s.Service == "" || s.Image == "" {
			return nil, fmt.Errorf("compose: spec needs Service and Image, got %+v", s)
		}

		svc := map[string]any{"image": s.Image}
		if s.Entrypoint != "" {
			svc["entrypoint"] = strings.Fields(s.Entrypoint)
		}
		if len(s.Ports) > 0 {
			svc["ports"] = s.Ports
		}
		if len(s.Env) > 0 {
			svc["environment"] = s.Env
		}
		if len(s.ExtraHosts) > 0 {
			svc["extra_hosts"] = s.ExtraHosts
		}

		volumes, err := renderVolumes(s)
		if err != nil {
			return nil, err
		}
		if len(volumes) > 0 {
			svc["volumes"] = volumes
		}

		services[s.Service] = svc
	}

	return yaml.Marshal(map[string]any{"services": services})
}

// renderVolumes turns a spec's mounts into compose short-syntax bind entries
// ("host:container"). A mount without a concrete host source is a hard error —
// inline Content must be materialized first.
func renderVolumes(s ContainerSpec) ([]string, error) {
	if len(s.Mount) == 0 {
		return nil, nil
	}

	volumes := make([]string, 0, len(s.Mount))
	for _, m := range s.Mount {
		if m.SourceFile == "" {
			return nil, fmt.Errorf("compose: mount %q in %q has no host source (materialize inline Content first)", m.Path, s.Service)
		}
		volumes = append(volumes, m.SourceFile+":"+m.Path)
	}
	return volumes, nil
}
