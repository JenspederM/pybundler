package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates
var templates embed.FS

func RenderTemplate(name string, data interface{}) (string, error) {
	// Read the template file
	f, err := templates.ReadFile("templates/" + name)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse the template
	t, err := template.New(name).Parse(string(f))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with the provided data
	var output bytes.Buffer
	err = t.ExecuteTemplate(&output, name, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}
