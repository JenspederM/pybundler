package bundle

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func RenderProject(bo *BundleOptions, errs ...error) error {
	if bo == nil {
		return fmt.Errorf("unable to render project: bundle options is nil")
	}
	cmdMod := "cmd"
	rootCmd, err := NewRootCommand(bo.PyProject.Project.Name, cmdMod)
	if err != nil {
		return fmt.Errorf("creating root command: %v", err)
	}
	err = SaveTemplate("generate.go.tmpl", filepath.Join(bo.Output, "generate/main.go"), rootCmd)
	if err != nil {
		return fmt.Errorf("rendering generate.go: %v", err)
	}
	err = SaveTemplate("main.go.tmpl", filepath.Join(bo.Output, "main.go"), rootCmd)
	if err != nil {
		return fmt.Errorf("rendering main.go: %v", err)
	}
	err = SaveTemplate("dockerfile.tmpl", filepath.Join(bo.Output, "Dockerfile"), rootCmd)
	if err != nil {
		return fmt.Errorf("rendering Dockerfile: %v", err)
	}
	commands := make([]*Command, 0)
	only_one := len(bo.Commands.Scripts) + len(bo.Commands.GuiScripts) + len(bo.Commands.EntryPoints)
	if only_one == 0 {
		return fmt.Errorf("no commands found")
	}
	if only_one == 1 {
		slog.Info("Only one command found, creating a single command")
		switch {
		case len(bo.Commands.Scripts) == 1:
			bo.Commands.Scripts[0].Module = cmdMod
			err := RenderCmd(bo.Commands.Scripts[0], filepath.Join(bo.Output, cmdMod, "root.go"))
			if err != nil {
				return fmt.Errorf("rendering script command: %v", err)
			}
			return nil
		case len(bo.Commands.GuiScripts) == 1:
			bo.Commands.GuiScripts[0].Module = cmdMod
			err := RenderCmd(bo.Commands.GuiScripts[0], filepath.Join(bo.Output, cmdMod, "root.go"))
			if err != nil {
				return fmt.Errorf("rendering gui command: %v", err)
			}
			return nil
		case len(bo.Commands.EntryPoints) == 1:
			slog.Info("Only one entrypoint found, creating a single command")
			return fmt.Errorf("Not implemented. rendering entrypoint command: %v", err)
		default:
			return fmt.Errorf("no commands found")
		}
	}
	if len(bo.Commands.Scripts) > 0 {
		root, err := RenderGroup(*bo, "scripts", filepath.Join(bo.Output, "internal"), nil, bo.Commands.Scripts...)
		if err != nil {
			return fmt.Errorf("rendering script command group: %v", err)
		}
		commands = append(commands, root)
	}
	if len(bo.Commands.GuiScripts) > 0 {
		root, err := RenderGroup(*bo, "gui", filepath.Join(bo.Output, "internal"), nil, bo.Commands.GuiScripts...)
		if err != nil {
			return fmt.Errorf("rendering gui command group: %v", err)
		}
		commands = append(commands, root)
	}
	if len(bo.Commands.EntryPoints) > 0 {
		root, err := RenderGroup(*bo, "entrypoint", filepath.Join(bo.Output, "internal"), nil, bo.Commands.EntryPoints...)
		if err != nil {
			return fmt.Errorf("rendering entrypoint command group: %v", err)
		}
		for _, cmd := range bo.Commands.EntryPoints {
			_, err := RenderGroup(*bo, cmd.Module, filepath.Join(bo.Output, "internal", "entrypoint"), root, cmd.Commands...)
			if err != nil {
				return fmt.Errorf("rendering entrypoint command: %v", err)
			}
		}
		commands = append(commands, root)
	}
	rootCmd.Commands = commands
	return RenderCmd(rootCmd, filepath.Join(bo.Output, cmdMod, "root.go"))
}

func RenderGroup(options BundleOptions, module, output string, parent *Command, commands ...*Command) (*Command, error) {
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
		CmdVarName: fmt.Sprintf("%sCmd", ToPascalCase(module)),
		CmdUse:     module,
		Module:     module,
		Import:     imp,
		Commands:   commands,
	}
	for _, cmd := range root.Commands {
		if cmd.Cmd == "" {
			continue
		}
		slog.Debug("Rendering command module", "module", cmd.Module, "path", path)
		fp := filepath.Join(path, cmd.Module, fmt.Sprintf("%s.go", cmd.CmdVarName))
		err := RenderCmd(cmd, fp)
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

func RenderCmd(c *Command, output string) error {
	if c == nil {
		return fmt.Errorf("unable to render command: command is nil")
	}
	slog.Debug("Rendering command", "module", c.Module, "path", output)
	if c.Module == "cmd" {
		c.CmdVarName = "RootCmd"
		err := SaveTemplate("root-with-commands.go.tmpl", output, c)
		if err != nil {
			return fmt.Errorf("rendering root command: %v", err)
		}
	} else {
		c.CmdVarName = ToPascalCase(c.CmdVarName)
		err := SaveTemplate("command.go.tmpl", output, c)
		if err != nil {
			return fmt.Errorf("rendering command: %v", err)
		}
	}

	return nil
}
