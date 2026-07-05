package gherkin

import "strings"

// hasRow reports whether some output line, read as whitespace-separated fields,
// equals the expected fields. Matching on fields (not a raw substring) keeps
// column alignment from defeating the assertion. Shared by every controller that
// asserts on a rendered table row (list catalogue's "NAME KIND" pair, list
// skills/agents' "name [version] [lastUpdatedAt]" row). Pure so it is
// unit-tested without a process.
func hasRow(output, row string) bool {
	want := strings.Fields(row)
	for raw := range strings.SplitSeq(output, "\n") {
		if equalFields(strings.Fields(raw), want) {
			return true
		}
	}
	return false
}

// hasLine reports whether some output line, trimmed, equals the expected line. Pure
// so it is unit-tested without a process.
func hasLine(output, line string) bool {
	want := strings.TrimSpace(line)
	for raw := range strings.SplitSeq(output, "\n") {
		if strings.TrimSpace(raw) == want {
			return true
		}
	}
	return false
}

// equalFields reports whether two field slices are element-wise equal.
func equalFields(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
