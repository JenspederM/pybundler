package bundle_test

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jenspederm/pybundler/internal/bundle"
)

type TestCase struct {
	Name    string
	Command []string
}

func TestRun(t *testing.T) {
	examples := filepath.Join("..", "..", "examples")
	if _, err := os.Stat(examples); os.IsNotExist(err) {
		t.Fatalf("Examples directory does not exist: %v", err)
	}
	// for each example in examples
	files, err := os.ReadDir(examples)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}
	for _, f := range files {
		slog.Info("Running example", "name", f.Name())
		if f.IsDir() {
			fmt.Printf("Running example: %s\n in %s", f.Name(), f)
			path := filepath.Join(examples, f.Name())
			test_dir := fmt.Sprintf("./.test/%s", f.Name())
			b, err := bundle.New(path, test_dir, true)
			if err != nil {
				t.Fatalf("Failed to create bundle: %v", err)
			}
			err = b.Run(true)
			if err != nil {
				t.Fatalf("Failed to run bundle: %v", err)
			}
			args := []string{filepath.Join(test_dir, "main"), "scripts", "cli"}
			cmd := exec.Command(test_dir, args...)
			cmd.Dir = test_dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run command: %v", err)
			}
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Failed to get command output: %v", err)
			}
			if string(output) != "Hello from cli!" {
				t.Fatalf("Unexpected output: %s", output)
			}
		}
	}
}
