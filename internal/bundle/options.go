package bundle

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type Script struct {
	Name    string
	Command string
	Module  string
}

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
	Scripts   []*Script
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

	scripts := make([]*Script, 0)
	for k, v := range pyproject.Project.Scripts {
		s, err := NewScript(k, v)
		cobra.CheckErr(err)
		if s == nil {
			cobra.CheckErr(fmt.Errorf("script %s is nil", k))
		}
		scripts = append(scripts, s)
	}
	if len(scripts) == 0 {
		cobra.CheckErr(fmt.Errorf("no scripts found in pyproject.toml"))
	}

	err = os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("creating output directory: %v", err)
	}
	return &BundleOptions{
		Path:      path,
		Output:    output,
		PyProject: pyproject,
		Scripts:   scripts,
	}, nil
}

func NewScript(name string, entrypoint string) (*Script, error) {
	_vals := strings.SplitN(entrypoint, ":", 2)
	if len(_vals) < 2 {
		return nil, fmt.Errorf("invalid script format: %s", entrypoint)
	}
	imp := _vals[0]
	fun := _vals[1]
	nn := strings.TrimSpace(name)
	cmd := fmt.Sprintf("import %s; %s.%s()", imp, imp, fun)
	script := &Script{
		Name:    strings.ReplaceAll(nn, "_", "-"),
		Command: cmd,
		Module:  strings.ReplaceAll(nn, "-", "_"),
	}
	return script, nil
}
