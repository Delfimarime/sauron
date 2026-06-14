package docker

import "github.com/testcontainers/testcontainers-go/log"

// Options collects the optional inputs to New. The compose directory is a
// required positional argument of New, not an option.
type Options struct {
	mount  []FileSpec
	logger log.Logger
	specs  []ContainerSpec
}

func WithLogger(value log.Logger) func(*Options) {
	return func(o *Options) {
		o.logger = value
	}
}

// WithContainer adds dependency services alongside the reserved "main" service.
func WithContainer(value ...ContainerSpec) func(*Options) {
	return func(o *Options) {
		if o.specs == nil {
			o.specs = make([]ContainerSpec, 0, len(value))
		}
		o.specs = append(o.specs, value...)
	}
}

// WithFile mounts extra host files into the "main" service.
func WithFile(value ...FileSpec) func(*Options) {
	return func(o *Options) {
		if o.mount == nil {
			o.mount = make([]FileSpec, 0, len(value))
		}
		o.mount = append(o.mount, value...)
	}
}
