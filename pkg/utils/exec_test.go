package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCommand_EmptyCommandIsNoop(t *testing.T) {
	if err := RunCommand(""); err != nil {
		t.Fatalf("expected no error for empty command, got %v", err)
	}
}

func TestRunCommand_SuccessfulCommand(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")

	if err := RunCommand("touch " + marker); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("expected command to have run and created the marker file: %v", err)
	}
}

func TestRunCommand_FailingCommand(t *testing.T) {
	if err := RunCommand("false"); err == nil {
		t.Fatal("expected an error for a command that exits non-zero")
	}
}
