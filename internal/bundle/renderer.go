package bundle

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudflare/cfssl/log"
)

type Command struct {
	Origin     string
	AppName    string
	Module     string
	CmdVarName string
	CmdUse     string
	Cmd        string
	Import     string
	Commands   []*Command
}

func NewCommand(appName, name, value, origin string, commands ...*Command) (*Command, error) {
	if len(commands) > 0 {
		commands = append(commands, commands...)
	}
	log.Infof("Creating command '%s' with value '%s'", name, value)
	// import any_script.gui; any_script.gui.main()
	parts := strings.SplitN(value, ";", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid script format: %s", value)
	}
	im := strings.TrimSpace(parts[0])
	imp := strings.Split(im, " ")
	if len(imp) != 2 {
		return nil, fmt.Errorf("invalid import format: %s", im)
	}
	module := strings.TrimSpace(imp[1])
	method := strings.TrimSpace(parts[1])
	method = strings.TrimPrefix(method, module+".")
	slog.Info("Creating command", "module", module, "method", method)
	cmd := fmt.Sprintf("import %s; %s.%s", module, module, method)
	cmdUse := strings.TrimSpace(name)
	cmdUse = strings.ReplaceAll(cmdUse, " ", "-")
	cmdUse = strings.ReplaceAll(cmdUse, "_", "-")
	cmdVarName := strings.ReplaceAll(cmdUse, "-", "_")
	return &Command{
		AppName:    appName,
		Origin:     origin,
		Module:     cmdVarName,
		CmdVarName: toPascalCase(cmdVarName),
		CmdUse:     cmdUse,
		Cmd:        cmd,
		Import:     fmt.Sprintf("%s/cmd/%s", appName, cmdVarName),
		Commands:   commands,
	}, nil
}

func NewRootCommand(appName string, commands ...*Command) (*Command, error) {
	cmd := &Command{
		AppName:  appName,
		Module:   "root",
		Commands: commands,
	}
	return cmd, nil
}

func (c *Command) Render(output string) error {
	// Render the command using a template engine
	// This is a placeholder for the actual rendering logic
	// You can use any template engine of your choice
	// For example, using text/template from the standard library
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
