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

type Script struct {
	Package  string
	Function string
	Origin   string
	Name     string
	Command  string
	Module   string
}

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

type ScriptCollection struct {
	Scripts     []*Script
	GuiScripts  []*Script
	EntryPoints map[string][]*Script
}

type BundleOptions struct {
	Path      string
	Output    string
	PyProject *PyProject
	Scripts   *ScriptCollection
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
		Scripts:   scripts,
	}, nil
}

func NewScript(name string, entrypoint string, origin string) (*Script, error) {
	_vals := strings.SplitN(entrypoint, ":", 2)
	if len(_vals) < 2 {
		return nil, fmt.Errorf("invalid script format: %s", entrypoint)
	}
	imp := _vals[0]
	fun := _vals[1]
	nn := strings.TrimSpace(name)
	cmd := fmt.Sprintf("import %s; %s.%s()", imp, imp, fun)
	script := &Script{
		Package:  strings.ReplaceAll(origin, "-", "_"),
		Function: toPascalCase(name),
		Origin:   origin,
		Name:     nn,
		Command:  cmd,
		Module:   strings.ReplaceAll(nn, "-", "_"),
	}
	return script, nil
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

func collectScripts(pyproject PyProject) (*ScriptCollection, error) {
	sc := ScriptCollection{
		Scripts:     make([]*Script, 0),
		GuiScripts:  make([]*Script, 0),
		EntryPoints: make(map[string][]*Script),
	}
	for k := range pyproject.Project.EntryPoints {
		sc.EntryPoints[k] = make([]*Script, 0)
	}
	for name, group := range pyproject.Project.EntryPoints {
		for k, v := range group {
			s, err := NewScript(k, v, "entry-points")
			if err != nil {
				return nil, fmt.Errorf("error creating entry point '%s': %v", k, err)
			}
			if s == nil {
				return nil, fmt.Errorf("entry point '%s' is nil", k)
			}
			sc.EntryPoints[name] = append(sc.EntryPoints[name], s)
		}
	}
	for k, v := range pyproject.Project.Scripts {
		s, err := NewScript(k, v, "scripts")
		if err != nil {
			return nil, fmt.Errorf("error creating script '%s': %v", k, err)
		}
		if s == nil {
			return nil, fmt.Errorf("script '%s' is nil", k)
		}
		sc.Scripts = append(sc.Scripts, s)
	}
	for k, v := range pyproject.Project.GuiScripts {
		s, err := NewScript(k, v, "gui-scripts")
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
