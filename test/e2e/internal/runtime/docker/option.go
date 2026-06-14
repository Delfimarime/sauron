package docker

import "github.com/testcontainers/testcontainers-go/log"

type Options struct {
	directory string
	logger    log.Logger
	specs     []ContainerSpec
	mount     []FileSpec
}

func WithDirectory(value string) func(*Options) {
	return func(o *Options) {
		o.directory = value
	}
}

func WithLogger(value log.Logger) func(*Options) {
	return func(o *Options) {
		o.logger = value
	}
}

func WithContainer(value ...ContainerSpec) func(*Options) {
	return func(o *Options) {
		if o.specs == nil {
			o.specs = make([]ContainerSpec, 0, len(value))
		}
		o.specs = append(o.specs, value...)
	}
}

func WithFile(value ...FileSpec) func(*Options) {
	return func(o *Options) {
		if o.mount == nil {
			o.mount = make([]FileSpec, 0, len(value))
		}
		o.mount = append(o.mount, value...)
	}
}
