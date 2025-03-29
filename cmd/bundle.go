package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/jenspederm/pybundler/internal/tmpl"
	"github.com/jenspederm/pybundler/internal/types"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

type BundleOptions struct {
	Path    string
	Name    string
	Version string
	Output  string
}

func NewBundleOptions(path string, output string) *BundleOptions {
	if strings.TrimSpace(output) == "." {
		output = ""
	}
	if strings.TrimSpace(output) == "" {
		output = "dist"
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s", path, "pyproject.toml")); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("pyproject.toml not found in %s\n", path)
			return nil
		}
	}
	pt, err := os.ReadFile(fmt.Sprintf("%s/%s", path, "pyproject.toml"))
	if err != nil {
		fmt.Printf("Error reading pyproject.toml: %v\n", err)
		return nil
	}
	var pyproject types.PyProject
	toml.Unmarshal(pt, pyproject)
	if pyproject.Project.Name == "" {
		fmt.Printf("Project name not found in pyproject.toml\n")
		return nil
	}
	if pyproject.Project.Version == "" {
		fmt.Printf("Project version not found in pyproject.toml\n")
		return nil
	}
	return &BundleOptions{
		Name:    pyproject.Project.Name,
		Version: pyproject.Project.Version,
		Path:    path,
		Output:  output,
	}
}

func BundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Bundle a Python project",
		Long:  `Bundle a Python project into a single executable file.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Implementation here
			saveTemplate("main.go.tmpl", "main.go", nil)

			fmt.Println("Bundle created successfully.")
		},
	}

	cmd.Flags().StringP("name", "n", "", "Name of the bundle")
	cmd.Flags().StringP("version", "v", "", "Version of the bundle")
	cmd.Flags().StringP("output", "o", "", "Output directory for the bundle")

	return cmd
}
func saveTemplate(template string, output string, data map[string]interface{}) error {
	log.Infof("saving template %s to %s", template, output)
	if strings.TrimSpace(output) == "." {
		output = ""
	}
	if exist, err := exists(output); err != nil || !exist {
		err := os.MkdirAll(output, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating directory: %v", err)
		}
	}
	f, err := tmpl.RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("rendering template: %v", err)
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", output, "main.go"), []byte(f), 0644)
	if err != nil {
		return fmt.Errorf("writing file: %v", err)
	}
	return nil
}
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
