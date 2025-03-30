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

func Run(bo *BundleOptions) error {
	name := bo.PyProject.Project.Name
	version := bo.PyProject.Project.Version

	data := map[string]interface{}{
		"Name":    name,
		"Version": version,
		"Path":    bo.Path,
		"Scripts": bo.Scripts,
	}

	_, err := RunCmd(bo.Output, "go", "mod", "init", name)
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

	_, err = RunCmd(bo.Output, "go", "mod", "tidy")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Path, "uv", "build", "--wheel", "-o", bo.Output)
	cobra.CheckErr(err)
	pkgReqs, err := RunCmd(bo.Path, "uv", "export", "--no-emit-project", "--no-dev", "--no-hashes")
	cobra.CheckErr(err)
	whl := fmt.Sprintf("%s-%s-py3-none-any.whl", strings.ReplaceAll(name, "-", "_"), version)
	requirements := bytes.Join([][]byte{[]byte(whl), pkgReqs}, []byte("\n"))
	err = os.WriteFile(filepath.Join(bo.Output, "requirements.txt"), requirements, 0644)
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, "go", "generate", "./...")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, "go", "build", "-o", "main")
	cobra.CheckErr(err)
	log.Info("Bundle created successfully.")
	return nil
}
