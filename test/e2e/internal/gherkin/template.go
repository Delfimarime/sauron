package gherkin

import (
	"bytes"
	"text/template"
)

// render treats text as a Go template and renders it against the world, so a
// feature's docstrings and step args resolve {{.Environment.…}} and
// {{.Variables.…}}. Missing keys are an error, surfacing a feature typo rather
// than emitting "<no value>".
func render(text string, w *World) (string, error) {
	tmpl, err := template.New("step").Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, w); err != nil {
		return "", err
	}

	return out.String(), nil
}
