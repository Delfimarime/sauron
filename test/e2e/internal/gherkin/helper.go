package gherkin

import "fmt"

// assertExpected returns an error unless actual equals expected. Generic so
// controllers can assert any comparable value, not just strings.
func assertExpected[T comparable](field string, expected, actual T) error {
	if expected != actual {
		return fmt.Errorf("%s: expected %v but got %v", field, expected, actual)
	}
	return nil
}
