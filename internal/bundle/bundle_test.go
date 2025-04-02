package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBundle(t *testing.T) {
	path := "../../examples/basic"
	output := "./.bundle_test"
	absOutput, err := filepath.Abs(output)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	bundle, err := New(path, output, false)
	if err != nil {
		t.Fatalf("Failed to create bundle: %v", err)
	}
	defer os.RemoveAll(absOutput)

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
}
