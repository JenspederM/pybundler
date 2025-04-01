package bundle

import (
	"fmt"
	"log/slog"
	"strings"
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
