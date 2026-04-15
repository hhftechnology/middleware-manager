package config

import (
	"os"
	"path/filepath"
	"testing"

	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
)

func TestFileStoreConfigDirPrecedence(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "alpha.yml")
	second := filepath.Join(dir, "beta.yaml")
	if err := os.WriteFile(first, []byte("http: {}\n"), 0o644); err != nil {
		t.Fatalf("write first config: %v", err)
	}
	if err := os.WriteFile(second, []byte("http: {}\n"), 0o644); err != nil {
		t.Fatalf("write second config: %v", err)
	}
	store, err := NewFileStore(tmtypes.RuntimeConfig{ConfigDir: dir, ConfigPaths: []string{"/ignored.yml"}})
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	paths := store.ConfigPaths()
	if len(paths) != 2 {
		t.Fatalf("expected config dir paths, got %v", paths)
	}
	if paths[0] != first || paths[1] != second {
		t.Fatalf("unexpected config order: %v", paths)
	}
}

func TestValidateBackupPathRejectsTraversal(t *testing.T) {
	store, err := NewFileStore(tmtypes.RuntimeConfig{ConfigPath: filepath.Join(t.TempDir(), "dynamic.yml"), BackupDir: t.TempDir()})
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	if _, err := store.ValidateBackupPath("../secret.yml.20260101_010101.bak"); err == nil {
		t.Fatalf("expected traversal backup name to be rejected")
	}
}
