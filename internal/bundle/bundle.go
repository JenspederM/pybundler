package bundle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudflare/cfssl/log"
)

type BundleOptions struct {
	Path      string
	Output    string
	PyProject *PyProject
	Commands  *CommandCollection
}

func New(path string, output string) (*BundleOptions, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	if strings.TrimSpace(output) == "." {
		output = ""
	}

	pyproject, err := DecodePyproject(path)
	if err != nil {
		return nil, fmt.Errorf("error decoding pyproject.toml: %v", err)
	}

	scripts, err := NewCommandCollection(*pyproject)
	if err != nil {
		return nil, fmt.Errorf("error collecting scripts: %v", err)
	}

	if strings.TrimSpace(output) == "" {
		output = filepath.Join(DEFAULT_BUNDLE_DIR, pyproject.Project.Name)
	}

	if !filepath.IsAbs(output) {
		output, err = filepath.Abs(output)
		if err != nil {
			return nil, fmt.Errorf("getting absolute path for output directory: %v", err)
		}
	}

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

func (bo *BundleOptions) GetRequirements(pkgReqs []byte) ([]byte, error) {

	log.Infof("Package requirements: %s", pkgReqs)
	whl := fmt.Sprintf("%s-%s-py3-none-any.whl", strings.ReplaceAll(bo.PyProject.Project.Name, "-", "_"), bo.PyProject.Project.Version)
	reqLines := bytes.Split(pkgReqs, []byte("\n"))
	reqs := [][]byte{[]byte(whl)}
	for _, line := range reqLines {
		if !bytes.HasPrefix(line, []byte("#")) && bytes.Contains(line, []byte("==")) {
			reqs = append(reqs, line)
		}
	}
	// reqs = append(reqs, []byte(whl))
	return bytes.Join(reqs, []byte("\n")), nil
}
