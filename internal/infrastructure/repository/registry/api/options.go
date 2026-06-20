package api

import "github.com/delfimarime/sauron/pkg/sauron/extension"

// Resolve folds opts into a populated Options value.
func Resolve(opts []extension.Option) extension.Options {
	options := extension.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// HasAuth reports whether any credential is set.
func HasAuth(o extension.Options) bool {
	return o.Username != "" || o.Password != ""
}

// HasTLS reports whether any transport-security option is set.
func HasTLS(o extension.Options) bool {
	return o.SkipTLSVerify || o.CACert != "" || o.ClientCert != "" || o.ClientKey != ""
}
