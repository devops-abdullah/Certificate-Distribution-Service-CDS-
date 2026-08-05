package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReloadNginx_EmptyCommandIsNoop(t *testing.T) {
	if err := ReloadNginx(""); err != nil {
		t.Fatalf("expected no error for empty command, got %v", err)
	}
}

func TestReloadNginx_SuccessfulCommand(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "reloaded")

	if err := ReloadNginx("touch " + marker); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("expected reload command to have run and created the marker file: %v", err)
	}
}

func TestReloadNginx_FailingCommand(t *testing.T) {
	if err := ReloadNginx("false"); err == nil {
		t.Fatal("expected an error for a command that exits non-zero")
	}
}
