package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

const DEFAULT_BUNDLE_DIR = ".pybundler"

type Project struct {
	Name        string                       `toml:"name"`
	Version     string                       `toml:"version"`
	Scripts     map[string]string            `toml:"scripts"`
	GuiScripts  map[string]string            `toml:"gui-scripts"`
	EntryPoints map[string]map[string]string `toml:"entry-points"`
}

type PyProject struct {
	Project Project `toml:"project"`
}

type CommandCollection struct {
	Scripts     []*Command
	GuiScripts  []*Command
	EntryPoints []*Command
}

type BundleOptions struct {
	Path      string
	Output    string
	PyProject *PyProject
	Commands  *CommandCollection
}

func NewBundleOptions(path string, output string) (*BundleOptions, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	if strings.TrimSpace(output) == "." {
		output = ""
	}

	pyproject, err := decodePyproject(path)
	if err != nil {
		return nil, fmt.Errorf("error decoding pyproject.toml: %v", err)
	}

	scripts, err := collectScripts(*pyproject)
	if err != nil {
		return nil, fmt.Errorf("error collecting scripts: %v", err)
	}

	if strings.TrimSpace(output) == "" {
		output = filepath.Join(DEFAULT_BUNDLE_DIR, pyproject.Project.Name)
	}

	output = makePathAbsolute(output)

	err = os.MkdirAll(output, os.ModePerm)

	if err != nil {
		return nil, fmt.Errorf("creating output directory: %v", err)
	}

	return &BundleOptions{
		Path:      path,
		Output:    output,
		PyProject: pyproject,
		Commands:  scripts,
	}, nil
}

func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func makePathAbsolute(p string) string {
	if p[0] != '/' {
		cwd, err := os.Getwd()
		if err != nil {
			cobra.CheckErr(fmt.Errorf("getting current working directory: %v", err))
		}
		p = filepath.Join(cwd, p)
	}
	return p
}

func checkPyproject(p string) error {
	if _, err := os.Stat(filepath.Join(p, "pyproject.toml")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pyproject.toml not found in %s", p)
		}
	}
	return nil
}

func decodePyproject(p string) (*PyProject, error) {
	if err := checkPyproject(p); err != nil {
		return nil, err
	}

	fp := filepath.Join(p, "pyproject.toml")
	pt, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("reading pyproject.toml from %s", fp)
	}

	var pyproject PyProject
	_, err = toml.Decode(string(pt), &pyproject)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling pyproject.toml from %s", fp)
	}
	if pyproject.Project.Name == "" {
		return nil, fmt.Errorf("project name not found in %s", fp)
	}
	if pyproject.Project.Version == "" {
		return nil, fmt.Errorf("project version not found in %s", fp)
	}
	return &pyproject, nil
}

func collectScripts(pyproject PyProject) (*CommandCollection, error) {
	project_name := pyproject.Project.Name
	sc := CommandCollection{
		Scripts:     make([]*Command, 0),
		GuiScripts:  make([]*Command, 0),
		EntryPoints: make([]*Command, 0),
	}

	for group_name, group := range pyproject.Project.EntryPoints {
		if group_name == "console_scripts" || group_name == "gui_scripts" {
			continue
		}
		entry_cmds := make([]*Command, 0)
		for k, v := range group {
			s, err := NewCommand(project_name, k, v, group_name)
			if err != nil {
				return nil, fmt.Errorf("error creating entry point '%s': %v", k, err)
			}
			if s == nil {
				return nil, fmt.Errorf("entry point '%s' is nil", k)
			}
			entry_cmds = append(entry_cmds, s)
		}
		if len(entry_cmds) == 0 {
			return nil, fmt.Errorf("no entry points found in group '%s'", group_name)
		}
		group_root, err := NewRootCommand(project_name, group_name, entry_cmds...)
		if err != nil {
			return nil, fmt.Errorf("error creating entry point group '%s': %v", group_name, err)
		}
		if group_root == nil {
			return nil, fmt.Errorf("entry point group '%s' is nil", group_name)
		}
		sc.EntryPoints = append(sc.EntryPoints, group_root)
	}

	for k, v := range pyproject.Project.Scripts {
		s, err := NewCommand(project_name, k, v, "scripts")
		if err != nil {
			return nil, fmt.Errorf("error creating script '%s': %v", k, err)
		}
		if s == nil {
			return nil, fmt.Errorf("script '%s' is nil", k)
		}
		sc.Scripts = append(sc.Scripts, s)
	}

	for k, v := range pyproject.Project.GuiScripts {
		s, err := NewCommand(project_name, k, v, "gui-scripts")
		if err != nil {
			return nil, fmt.Errorf("error creating gui script '%s': %v", k, err)
		}
		if s == nil {
			return nil, fmt.Errorf("gui script '%s' is nil", k)
		}
		sc.GuiScripts = append(sc.GuiScripts, s)
	}

	return &sc, nil
}
