package bundle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

func Run(bo *BundleOptions, verbose bool) error {
	name := bo.PyProject.Project.Name
	version := bo.PyProject.Project.Version

	data := map[string]interface{}{
		"Name":    name,
		"Version": version,
		"Path":    bo.Path,
		"Scripts": bo.Scripts,
	}

	_, err := RunCmd(bo.Output, verbose, "go", "mod", "init", name)
	cobra.CheckErr(err)
	err = SaveTemplate("main.go.tmpl", filepath.Join(bo.Output, "main.go"), data)
	cobra.CheckErr(err)
	err = SaveTemplate("generate.go.tmpl", filepath.Join(bo.Output, "generate/main.go"), data)
	cobra.CheckErr(err)

	if len(bo.Scripts) > 1 {
		err = SaveTemplate("root.go.tmpl", filepath.Join(bo.Output, "cmd/root.go"), data)
	} else {
		err = SaveTemplate("root-single.go.tmpl", filepath.Join(bo.Output, "cmd/root.go"), data)
	}
	cobra.CheckErr(err)

	_, err = RunCmd(bo.Output, verbose, "go", "mod", "tidy")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Path, verbose, "uv", "build", "--wheel", "-o", bo.Output)
	cobra.CheckErr(err)
	requirements := getRequirements(bo.Path, name, version, verbose)
	err = os.WriteFile(filepath.Join(bo.Output, "requirements.txt"), requirements, 0644)
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, verbose, "go", "generate", "./...")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, verbose, "go", "build", "-o", "main")
	cobra.CheckErr(err)
	log.Info("Bundle created successfully.")
	return nil
}

func getRequirements(path string, name string, version string, verbose bool) []byte {
	pkgReqs, err := RunCmd(path, verbose, "uv", "export", "--no-emit-project", "--no-dev", "--no-hashes")
	cobra.CheckErr(err)
	whl := fmt.Sprintf("%s-%s-py3-none-any.whl", strings.ReplaceAll(name, "-", "_"), version)
	reqLines := bytes.Split(pkgReqs, []byte("\n"))
	reqs := make([][]byte, 0)
	for _, line := range reqLines {
		if bytes.Contains(line, []byte("==")) || bytes.HasSuffix(line, []byte(".whl")) {
			parts := bytes.Split(line, []byte("=="))
			if len(parts) == 2 {
				reqs = append(reqs, []byte(fmt.Sprintf("%s==%s", string(parts[0]), string(parts[1]))))
			}
		}
	}
	reqs = append(reqs, []byte(whl))
	return bytes.Join(reqs, []byte("\n"))
}
