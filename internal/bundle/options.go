package bundle

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type Project struct {
	Name    string            `toml:"name"`
	Version string            `toml:"version"`
	Scripts map[string]string `toml:"scripts"`
}

type PyProject struct {
	Project Project `toml:"project"`
}

type BundleOptions struct {
	Path      string
	Output    string
	PyProject PyProject
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

	var pyproject PyProject
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
		Path:      path,
		Output:    output,
		PyProject: pyproject,
	}, nil
}
