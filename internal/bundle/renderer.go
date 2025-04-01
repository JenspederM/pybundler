package bundle

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudflare/cfssl/log"
)

func RenderProject(bo *BundleOptions, errs ...error) error {
	if bo == nil {
		return fmt.Errorf("unable to render project: bundle options is nil")
	}
	cmdMod := "root"
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
	commands := make([]*Command, 0)
	only_one := len(bo.Commands.Scripts) + len(bo.Commands.GuiScripts) + len(bo.Commands.EntryPoints)
	if only_one == 0 {
		return fmt.Errorf("no commands found")
	}
	if only_one == 1 {
		switch {
		case len(bo.Commands.Scripts) == 1:
			fmt.Printf("Only one script found, creating a single command\n")
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
			fmt.Print("Only one entrypoint found, creating a single command\n")
			log.Fatalf("Not implemented yet")
		default:
			fmt.Print("Only one command found, creating a single command\n")
			log.Fatalf("Not implemented yet")
		}
	}
	if len(bo.Commands.Scripts) > 0 {
		root, err := RenderGroup("scripts", filepath.Join(bo.Output, "internal"), *bo, nil, bo.Commands.Scripts...)
		if err != nil {
			return fmt.Errorf("rendering script command group: %v", err)
		}
		commands = append(commands, root)
	}
	if len(bo.Commands.GuiScripts) > 0 {
		root, err := RenderGroup("gui", filepath.Join(bo.Output, "internal"), *bo, nil, bo.Commands.GuiScripts...)
		if err != nil {
			return fmt.Errorf("rendering gui command group: %v", err)
		}
		commands = append(commands, root)
	}
	if len(bo.Commands.EntryPoints) > 0 {
		root, err := RenderGroup("entrypoint", filepath.Join(bo.Output, "internal"), *bo, nil, bo.Commands.EntryPoints...)
		if err != nil {
			return fmt.Errorf("rendering entrypoint command group: %v", err)
		}
		for _, cmd := range bo.Commands.EntryPoints {
			_, err := RenderGroup(cmd.Module, filepath.Join(bo.Output, "internal", "entrypoint"), *bo, root, cmd.Commands...)
			if err != nil {
				return fmt.Errorf("rendering entrypoint command: %v", err)
			}
		}
		commands = append(commands, root)
	}
	rootCmd.Commands = commands
	return RenderCmd(rootCmd, filepath.Join(bo.Output, cmdMod, "root.go"))
}

func RenderGroup(module, output string, options BundleOptions, parent *Command, commands ...*Command) (*Command, error) {
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
	log.Infof("Rendering command '%s' at %s", c.Module, output)
	if c.Module == "cmd" {
		if len(c.Commands) > 0 {
			return fmt.Errorf("cmd module cannot have subcommands")
		}
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
