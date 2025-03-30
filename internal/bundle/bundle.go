package bundle

import (
	"bytes"
	"fmt"
	"os"

	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

func Run(bo *BundleOptions) error {
	name := bo.PyProject.Project.Name
	version := bo.PyProject.Project.Version
	modules := make([]string, 0)
	for _, script := range bo.Scripts {
		modules = append(modules, script.Module)
	}

	data := map[string]interface{}{
		"Name":    name,
		"Version": version,
		"Path":    bo.Path,
		"Modules": modules,
		"Scripts": bo.Scripts,
	}

	_, err := RunCmd(bo.Output, "go", "mod", "init", name)
	cobra.CheckErr(err)
	err = SaveTemplate("main.go.tmpl", fmt.Sprintf("%s/%s", bo.Output, "main.go"), data)
	cobra.CheckErr(err)
	err = SaveTemplate("generate.go.tmpl", fmt.Sprintf("%s/%s", bo.Output, "generate/main.go"), data)
	cobra.CheckErr(err)
	err = SaveTemplate("root.go.tmpl", fmt.Sprintf("%s/%s", bo.Output, "cmd/root.go"), data)
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, "go", "mod", "tidy")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Path, "uv", "build", "--wheel", "-o", bo.Output)
	cobra.CheckErr(err)
	pkgReqs, err := RunCmd(bo.Path, "uv", "export", "--no-emit-project", "--no-dev", "--no-hashes")
	cobra.CheckErr(err)
	whl := fmt.Sprintf("%s-%s-py3-none-any.whl", name, version)
	requirements := bytes.Join([][]byte{[]byte(whl), pkgReqs}, []byte("\n"))
	err = os.WriteFile(fmt.Sprintf("%s/%s", bo.Output, "requirements.txt"), requirements, 0644)
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, "go", "generate", "./...")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Output, "go", "build", "-o", "main")
	cobra.CheckErr(err)
	log.Info("Bundle created successfully.")
	return nil
}
