// Package e2e contains sauron's black-box integration suite. The godog entrypoint
// in integration_test.go runs by default (plain `go test ./...`, the gate). The
// in-process tests under internal/ are tagged `unit` and run only with
// `go test -tags unit ./...`; that tag excludes the integration suite, so the two
// never run together. This file exists so the package still builds under -tags
// unit, where integration_test.go is excluded.
package e2e
