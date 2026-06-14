// Package docker holds the e2e harness's Runtime abstraction: where the
// command-under-test executes (the host OS or a Testcontainers sandbox), and the
// container declarations a scenario's Given steps build up.
package docker

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ContainerSpec declares one dependency a scenario needs. Given steps append a
// spec to the world; the runtime realizes specs as containers. Service is the
// logical (and compose) service name.
type ContainerSpec struct {
	Service    string
	Image      string
	Entrypoint string
	Ports      []string
	Mount      []FileSpec
	Env        map[string]string
}

type FileSpec struct {
	Path       string
	Content    []byte
	SourceFile string
}

// GenerateDockerComposeFile renders specs into a docker-compose.yml document. It is pure
// (specs -> bytes) so it is unit-tested without Docker; composeRuntime writes the
// bytes to a temp file and feeds them to the compose module.
func GenerateDockerComposeFile(specs []ContainerSpec) ([]byte, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("synthCompose: no container specs")
	}

	services := make(map[string]any, len(specs))
	for _, s := range specs {
		if s.Service == "" || s.Image == "" {
			return nil, fmt.Errorf("synthCompose: spec needs Service and Image, got %+v", s)
		}

		svc := map[string]any{"image": s.Image}
		if len(s.Ports) > 0 {
			svc["ports"] = s.Ports
		}
		if len(s.Env) > 0 {
			svc["environment"] = s.Env
		}
		//TODO support mount  here
		services[s.Service] = svc
	}

	return yaml.Marshal(map[string]any{"services": services})
}
