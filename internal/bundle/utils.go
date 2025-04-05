package bundle

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"math/rand"
	"time"
)

func RunCmd(cwd string, verbose bool, args ...string) ([]byte, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	if verbose {
		slog.Info("Running command", "args", strings.Join(args, " "))
		cmd.Stdout = mw
		cmd.Stderr = mw
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running command: %v", err)
	}
	res := stdBuffer.String()
	if verbose {
		slog.Info("Command output", "output", res)
	}
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

func ToPascalCase(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
