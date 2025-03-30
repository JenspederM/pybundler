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
	err = createScripts(bo)
	cobra.CheckErr(err)

	_, err = RunCmd(bo.Output, verbose, "go", "mod", "tidy")
	cobra.CheckErr(err)
	_, err = RunCmd(bo.Path, verbose, "uv", "build", "--wheel", "-o", bo.Output)
	cobra.CheckErr(err)
	requirements, err := getRequirements(bo.Path, name, version, verbose)
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
	log.Info("Bundle created successfully.")
	return nil
}

func createScripts(bo *BundleOptions) error {
	allSripts := append(bo.Scripts.Scripts, bo.Scripts.GuiScripts...)
	for k, scripts := range bo.Scripts.EntryPoints {
		for _, script := range scripts {
			if script.Origin == "" {
				script.Origin = k
			}
			allSripts = append(allSripts, script)
		}
	}
	commands := make([]*Command, 0)
	for _, script := range allSripts {
		c, err := NewCommand(bo.PyProject.Project.Name, script.Name, script.Command, script.Origin)
		if err != nil {
			return fmt.Errorf("creating command: %v", err)
		}
		commands = append(commands, c)
	}
	if len(commands) == 1 {
		c := commands[0]
		c.Module = "root"
		err := c.Render(filepath.Join(bo.Output, "cmd", "root.go"))
		if err != nil {
			return fmt.Errorf("rendering command: %v", err)
		}
		return nil
	}
	for _, cmd := range commands {
		err := cmd.Render(filepath.Join(bo.Output, "internal", cmd.Module, fmt.Sprintf("%s.go", cmd.Module)))
		if err != nil {
			return fmt.Errorf("rendering command: %v", err)
		}
	}
	rootCmd, err := NewRootCommand(bo.PyProject.Project.Name, commands...)
	if err != nil {
		return fmt.Errorf("creating root command: %v", err)
	}
	err = rootCmd.Render(filepath.Join(bo.Output, "cmd", "root.go"))
	if err != nil {
		return fmt.Errorf("rendering root command: %v", err)
	}

	return nil
}

func getRequirements(path string, name string, version string, verbose bool) ([]byte, error) {
	pkgReqs, err := RunCmd(path, verbose, "uv", "export", "--no-emit-project", "--no-dev", "--no-hashes")
	if err != nil {
		return nil, err
	}
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
	return bytes.Join(reqs, []byte("\n")), nil
}
