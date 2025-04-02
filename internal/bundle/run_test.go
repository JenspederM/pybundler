package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

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
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("Running example: %s\n", file.Name())
		}
	}
	path := "../../examples/basic"
	output := "./.run_test"
	absOutput, err := filepath.Abs(output)

	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	bundle, err := New(path, output, false)
	if err != nil {
		t.Fatalf("Failed to create bundle: %v", err)
	}
	// defer os.RemoveAll(absOutput)

	if bundle.Path != path {
		t.Errorf("Expected path %s, got %s", path, bundle.Path)
	}
	if bundle.Output != absOutput {
		t.Errorf("Expected output %s, got %s", absOutput, bundle.Output)
	}
	if bundle.PyProject == nil {
		t.Errorf("Expected PyProject to be initialized, got nil")
	}
	if bundle.Commands == nil {
		t.Errorf("Expected Commands to be initialized, got nil")
	}
	if bundle.Commands.Scripts == nil {
		t.Errorf("Expected Scripts to be initialized, got nil")
	}
	if len(bundle.Commands.Scripts) == 0 {
		t.Errorf("Expected Scripts to have elements, got empty")
	}

	err = bundle.Run(true)
	if err != nil {
		t.Fatalf("Failed to run bundle: %v", err)
	}

	if _, err := os.Stat(filepath.Join(absOutput, "generate/main.go")); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("generate/main.go not found in %s", absOutput)
		}
	}
	if _, err := os.Stat(filepath.Join(absOutput, "main.go")); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("main.go not found in %s", absOutput)
		}
	}
	if _, err := os.Stat(filepath.Join(absOutput, "main")); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("main.go not found in %s", absOutput)
		}
	}
}
