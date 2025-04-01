package bundle

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const DEFAULT_BUNDLE_DIR = ".pybundler"

type ProjectSection struct {
	Name        string                       `toml:"name"`
	Version     string                       `toml:"version"`
	Scripts     map[string]string            `toml:"scripts"`
	GuiScripts  map[string]string            `toml:"gui-scripts"`
	EntryPoints map[string]map[string]string `toml:"entry-points"`
}

type PyProject struct {
	Project ProjectSection `toml:"project"`
}

func DecodePyproject(p string) (*PyProject, error) {
	if _, err := os.Stat(filepath.Join(p, "pyproject.toml")); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("pyproject.toml not found in %s", p)
		}
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
