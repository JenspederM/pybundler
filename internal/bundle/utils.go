package bundle

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"math/rand"
	"time"

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

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func RandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return RandomStringWithCharset(length, charset)
}
