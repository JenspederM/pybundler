package bundle

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cloudflare/cfssl/log"
)

//go:embed templates
var templates embed.FS

func SaveTemplate(template string, output string, data interface{}) error {
	parent := filepath.Dir(output)
	if _, err := os.Stat(parent); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(parent, os.ModePerm)
		} else {
			return fmt.Errorf("creating output directory: %v", err)
		}
	}
	log.Infof("Saving template %s to %s", template, output)
	if strings.TrimSpace(output) == "." {
		output = ""
	}

	f, err := RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("rendering template: %v", err)
	}
	err = os.WriteFile(output, []byte(f), 0644)
	if err != nil {
		return fmt.Errorf("writing file: %v", err)
	}
	return nil
}

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
