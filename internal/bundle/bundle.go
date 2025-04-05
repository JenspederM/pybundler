package bundle

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const EXAMPLES_DIR = "../..examples"

type BundleOptions struct {
	Path      string
	Output    string
	PyProject *PyProject
	Commands  *CommandCollection
}

func New(path string, output string, overwrite bool) (*BundleOptions, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}
	if strings.TrimSpace(output) == "." {
		output = ""
	}

	pyproject, err := NewPyProject(path)
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

	bundle := &BundleOptions{
		Path:      path,
		Output:    output,
		PyProject: pyproject,
		Commands:  scripts,
	}

	err = os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("creating output directory: %v", err)
	}

	if _, err := os.Stat(bundle.Output); err == nil {
		isEmpty, err := IsEmpty(bundle.Output)
		cobra.CheckErr(err)
		if !isEmpty && !overwrite {
			fp := filepath.Join(bundle.Output, "main.go")
			slog.Info(fmt.Sprintf("File %s already exists. Use --overwrite to overwrite.", fp))
			return nil, fmt.Errorf("output directory %s already exists", bundle.Output)
		}
		err = os.RemoveAll(bundle.Output)
		cobra.CheckErr(err)
		err = os.MkdirAll(bundle.Output, os.ModePerm)
		cobra.CheckErr(err)
	}

	slog.Info("Creating bundle:", "source", bundle.Path, "target", bundle.Output)
	return bundle, nil
}

func (bo *BundleOptions) Run(verbose bool) error {
	_, err := RunCmd(bo.Output, verbose, "go", "mod", "init", bo.PyProject.Project.Name)
	cobra.CheckErr(err)
	err = RenderProject(bo)
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, verbose, "go", "mod", "tidy")
	cobra.CheckErr(err)

	_, err = RunCmd(bo.Path, verbose, "uv", "build", "--wheel", "-o", bo.Output)
	cobra.CheckErr(err)
	pkgReqs, err := RunCmd(bo.Path, verbose, "uv", "export", "--no-emit-project", "--no-dev", "--no-hashes")
	cobra.CheckErr(err)
	requirements, err := bo.parseRequirements(pkgReqs)
	cobra.CheckErr(err)
	err = os.WriteFile(filepath.Join(bo.Output, "requirements.txt"), requirements, 0644)
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, verbose, "go", "generate", "./...")
	cobra.CheckErr(err)

	_, err = RunCmd(bo.Output, verbose, "go", "fmt", "./...")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, verbose, "go", "mod", "tidy")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, verbose, "go", "build", "-o", "main")
	cobra.CheckErr(err)
	slog.Info("Bundle created successfully.")
	return nil
}

func (bo *BundleOptions) parseRequirements(pkgReqs []byte) ([]byte, error) {
	slog.Info("Getting module requirements")
	whl := fmt.Sprintf("%s-%s-py3-none-any.whl", strings.ReplaceAll(bo.PyProject.Project.Name, "-", "_"), bo.PyProject.Project.Version)
	reqLines := bytes.Split(pkgReqs, []byte("\n"))
	reqs := [][]byte{[]byte(whl)}
	for _, line := range reqLines {
		if !bytes.HasPrefix(line, []byte("#")) && bytes.Contains(line, []byte("==")) {
			reqs = append(reqs, line)
		}
	}
	return bytes.Join(reqs, []byte("\n")), nil
}
