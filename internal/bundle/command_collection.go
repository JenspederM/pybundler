package bundle

import "fmt"

type CommandCollection struct {
	Scripts     []*Command
	GuiScripts  []*Command
	EntryPoints []*Command
}

func NewCommandCollection(pyproject PyProject) (*CommandCollection, error) {
	project_name := pyproject.Project.Name
	sc := CommandCollection{
		Scripts:     make([]*Command, 0),
		GuiScripts:  make([]*Command, 0),
		EntryPoints: make([]*Command, 0),
	}

	for group_name, group := range pyproject.Project.EntryPoints {
		if group_name == "console_scripts" || group_name == "gui_scripts" {
			continue
		}
		entry_cmds := make([]*Command, 0)
		for k, v := range group {
			s, err := NewCommand(project_name, k, v, group_name)
			if err != nil {
				return nil, fmt.Errorf("error creating entry point '%s': %v", k, err)
			}
			if s == nil {
				return nil, fmt.Errorf("entry point '%s' is nil", k)
			}
			entry_cmds = append(entry_cmds, s)
		}
		if len(entry_cmds) == 0 {
			return nil, fmt.Errorf("no entry points found in group '%s'", group_name)
		}
		group_root, err := NewRootCommand(project_name, group_name, entry_cmds...)
		if err != nil {
			return nil, fmt.Errorf("error creating entry point group '%s': %v", group_name, err)
		}
		if group_root == nil {
			return nil, fmt.Errorf("entry point group '%s' is nil", group_name)
		}
		sc.EntryPoints = append(sc.EntryPoints, group_root)
	}

	for k, v := range pyproject.Project.Scripts {
		s, err := NewCommand(project_name, k, v, "scripts")
		if err != nil {
			return nil, fmt.Errorf("error creating script '%s': %v", k, err)
		}
		if s == nil {
			return nil, fmt.Errorf("script '%s' is nil", k)
		}
		sc.Scripts = append(sc.Scripts, s)
	}

	for k, v := range pyproject.Project.GuiScripts {
		s, err := NewCommand(project_name, k, v, "gui-scripts")
		if err != nil {
			return nil, fmt.Errorf("error creating gui script '%s': %v", k, err)
		}
		if s == nil {
			return nil, fmt.Errorf("gui script '%s' is nil", k)
		}
		sc.GuiScripts = append(sc.GuiScripts, s)
	}

	return &sc, nil
}
