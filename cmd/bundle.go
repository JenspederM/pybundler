package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cloudflare/cfssl/log"
	"github.com/jenspederm/pybundler/internal/tmpl"
	"github.com/jenspederm/pybundler/internal/types"
	"github.com/spf13/cobra"
)

type BundleOptions struct {
	Path    string
	Name    string
	Version string
	Output  string
}

func BundleCmd() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "bundle"
	cmd.Short = "Bundle a Python project"
	cmd.Long = `Bundle a Python project into a single executable file.`

	cmd.Flags().StringP("path", "p", ".", "Path to the Python project")
	cmd.Flags().StringP("output", "o", "dist", "Output directory for the bundle")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Implementation here
		path := cmd.Flag("path").Value.String()
		output := cmd.Flag("output").Value.String()
		bo, err := NewBundleOptions(path, output)
		cobra.CheckErr(err)
		data := map[string]interface{}{
			"Name":    bo.Name,
			"Version": bo.Version,
			"Path":    bo.Path,
		}

		err = runCmd(bo.Output, "go", "mod", "init", bo.Name)
		cobra.CheckErr(err)
		err = saveTemplate("main.go.tmpl", fmt.Sprintf("%s/%s", output, "main.go"), data)
		cobra.CheckErr(err)
		err = saveTemplate("generate.go.tmpl", fmt.Sprintf("%s/%s", output, "generate/main.go"), data)
		cobra.CheckErr(err)
		err = saveTemplate("root.go.tmpl", fmt.Sprintf("%s/%s", output, "cmd/root.go"), data)
		cobra.CheckErr(err)
		err = runCmd(bo.Output, "go", "mod", "tidy")
		cobra.CheckErr(err)
		err = runCmd(bo.Path, "uv", "build", "--wheel", "-o", bo.Output)
		cobra.CheckErr(err)
		err = runCmd(bo.Path, "uv", "export", "--no-emit-project", "--no-dev", "--no-hashes", "-o", bo.Output)
		cobra.CheckErr(err)
		err = runCmd(bo.Output, "go", "generate", "./...")
		cobra.CheckErr(err)
		err = runCmd(bo.Output, "go", "build", "-o", "main")
		cobra.CheckErr(err)
		fmt.Println("Bundle created successfully.")
	}

	return cmd
}

func NewBundleOptions(path string, output string) (*BundleOptions, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	if strings.TrimSpace(output) == "." {
		output = ""
	}
	if strings.TrimSpace(output) == "" {
		output = "dist"
	}

	if output[0] != '/' {
		cwd, err := os.Getwd()
		if err != nil {
			cobra.CheckErr(fmt.Errorf("getting current working directory: %v", err))
		}
		output = fmt.Sprintf("%s/%s", cwd, output)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", path, "pyproject.toml")); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("pyproject.toml not found in %s", path)
		}
	}

	pt, err := os.ReadFile(fmt.Sprintf("%s/%s", path, "pyproject.toml"))
	if err != nil {
		return nil, fmt.Errorf("error reading pyproject.toml: %v", err)
	}

	var pyproject types.PyProject
	_, err = toml.Decode(string(pt), &pyproject)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pyproject.toml: %v", err)
	}
	if pyproject.Project.Name == "" {
		return nil, fmt.Errorf("project name not found in pyproject.toml")
	}
	if pyproject.Project.Version == "" {
		return nil, fmt.Errorf("project version not found in pyproject.toml")
	}

	err = os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("creating output directory: %v", err)
	}
	return &BundleOptions{
		Name:    pyproject.Project.Name,
		Version: pyproject.Project.Version,
		Path:    path,
		Output:  output,
	}, nil
}

func saveTemplate(template string, output string, data map[string]interface{}) error {
	parent := filepath.Dir(output)
	if _, err := os.Stat(parent); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(parent, os.ModePerm)
		} else {
			return fmt.Errorf("creating output directory: %v", err)
		}
	}
	log.Infof("saving template %s to %s", template, output)
	if strings.TrimSpace(output) == "." {
		output = ""
	}

	f, err := tmpl.RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("rendering template: %v", err)
	}
	err = os.WriteFile(output, []byte(f), 0644)
	if err != nil {
		return fmt.Errorf("writing file: %v", err)
	}
	return nil
}

func runCmd(cwd string, args ...string) error {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	curCwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working directory: %v", err)
	}
	os.Chdir(cwd)
	defer os.Chdir(curCwd)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running command: %v", err)
	}
	return nil
}
