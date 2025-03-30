package bundle

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cloudflare/cfssl/log"
)

func RunCmd(cwd string, args ...string) ([]byte, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	log.Infof("Running command: %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	content, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running command: %v", err)
	}
	log.Infof("Command output:\n%s", string(content))
	return content, nil
}
