package bundle

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudflare/cfssl/log"
)

func RunCmd(cwd string, verbose bool, args ...string) ([]byte, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	log.Infof("Running command: %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	if verbose {
		cmd.Stdout = mw
		cmd.Stderr = mw
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running command: %v", err)
	}
	res := stdBuffer.String()
	log.Infof("Command output:\n%s", string(res))
	return []byte(res), nil
}
