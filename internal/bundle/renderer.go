package bundle

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

type Command struct {
	Origin     string
	AppName    string
	Module     string
	Import     string
	CmdVarName string
	CmdUse     string
	Cmd        string
	Commands   []*Command
}

func NewCommand(appName, name, value, origin string, commands ...*Command) (*Command, error) {
	slog.Info("Creating command input", "appName", appName, "name", name, "value", value, "origin", origin)
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		slog.Error("Invalid script format", "appName", appName, "name", name, "value", value, "origin", origin)
		return nil, fmt.Errorf("invalid script format: %s", value)
	}
	import_module := strings.TrimSpace(parts[0])
	method := strings.TrimSpace(parts[1])
	method = strings.TrimPrefix(method, import_module+".")
	cmd := fmt.Sprintf("import %s; %s.%s()", import_module, import_module, method)

	parts = strings.Split(import_module, ".")
	module := parts[len(parts)-1]

	cmdUse := strings.TrimSpace(name)
	cmdUse = strings.ReplaceAll(cmdUse, " ", "-")
	cmdUse = strings.ReplaceAll(cmdUse, "_", "-")
	cmdVarName := strings.ReplaceAll(cmdUse, "-", "_")
	slog.Info("Creating command output",
		"AppName", appName,
		"Origin", origin,
		"Module", module+RandomString(5),
		"CmdVarName", toPascalCase(cmdVarName),
		"CmdUse", cmdUse,
		"Cmd", cmd,
		"Import", fmt.Sprintf("%s/cmd/%s", appName, cmdVarName),
		"Commands", commands,
	)
	m := module + RandomString(5)
	return &Command{
		AppName:    appName,
		Module:     m,
		Import:     m,
		CmdVarName: toPascalCase(cmdVarName),
		CmdUse:     cmdUse,
		Cmd:        cmd,
		Commands:   commands,
	}, nil
}

func NewRootCommand(appName, module string, commands ...*Command) (*Command, error) {
	root := &Command{
		AppName:    appName,
		Module:     strings.ReplaceAll(module, "-", "_"),
		Import:     strings.ReplaceAll(module, "-", "_"),
		CmdVarName: fmt.Sprintf("%sCmd", toPascalCase(module)),
		CmdUse:     module,
		Cmd:        "",
		Commands:   commands,
	}
	return root, nil
}

func (bo *BundleOptions) renderProject() error {
	data := map[string]interface{}{
		"Name":    bo.PyProject.Project.Name,
		"Version": bo.PyProject.Project.Version,
		"Path":    bo.Path,
		"Scripts": bo.Commands,
	}
	err := SaveTemplate("main.go.tmpl", filepath.Join(bo.Output, "main.go"), data)
	cobra.CheckErr(err)
	err = SaveTemplate("generate.go.tmpl", filepath.Join(bo.Output, "generate/main.go"), data)
	cobra.CheckErr(err)
	commands := make([]*Command, 0)
	kinds := []string{"scripts", "gui", "entrypoint"}
	for _, kind := range kinds {
		if len(bo.Commands.Scripts) > 0 {
			root, err := RenderModule(kind, filepath.Join(bo.Output, "internal"), *bo, nil, bo.Commands.Scripts...)
			cobra.CheckErr(err)
			if kind == "entrypoint" {
				for _, cmd := range bo.Commands.EntryPoints {
					_, err := RenderModule(cmd.Module, filepath.Join(bo.Output, "internal", "entrypoint"), *bo, root, cmd.Commands...)
					if err != nil {
						return fmt.Errorf("rendering entrypoint command: %v", err)
					}
				}
			}
			commands = append(commands, root)
		}
	}
	rootCmd, err := NewRootCommand(bo.PyProject.Project.Name, "root", commands...)
	if err != nil {
		return fmt.Errorf("creating root command: %v", err)
	}
	err = rootCmd.Render(filepath.Join(bo.Output, "cmd", "root.go"))
	if err != nil {
		return fmt.Errorf("rendering root command: %v", err)
	}
	return nil
}

func RenderModule(module, output string, options BundleOptions, parent *Command, commands ...*Command) (*Command, error) {
	path := filepath.Join(output, module)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("creating cmd directory: %v", err)
	}

	imp := module
	if parent != nil {
		imp = fmt.Sprintf("%s/%s", parent.Import, module)
	}

	root := &Command{
		AppName:    options.PyProject.Project.Name,
		CmdVarName: fmt.Sprintf("%sCmd", toPascalCase(module)),
		CmdUse:     module,
		Module:     module,
		Import:     imp,
		Commands:   commands,
	}
	for _, cmd := range root.Commands {
		if cmd.Cmd == "" {
			continue
		}
		log.Infof("Rendering command module '%s' at %s", cmd.Module, path)
		fp := filepath.Join(path, cmd.Module, fmt.Sprintf("%s.go", cmd.CmdVarName))
		err := cmd.Render(fp)
		if err != nil {
			return nil, fmt.Errorf("rendering command: %v", err)
		}
	}

	err = SaveTemplate("command-group.go.tmpl", filepath.Join(path, "root.go"), root)
	if err != nil {
		return nil, fmt.Errorf("rendering command group: %v", err)
	}
	return root, nil
}

func (c *Command) Render(output string) error {
	log.Infof("Rendering command '%s' at %s", c.Module, output)
	if c.Module == "root" && len(c.Commands) > 0 {
		c.CmdVarName = "RootCmd"
		err := SaveTemplate("root-with-commands.go.tmpl", output, c)
		if err != nil {
			return fmt.Errorf("rendering root command: %v", err)
		}
	} else {
		c.CmdVarName = toPascalCase(c.CmdVarName)
		err := SaveTemplate("command.go.tmpl", output, c)
		if err != nil {
			return fmt.Errorf("rendering command: %v", err)
		}
	}

	return nil
}
