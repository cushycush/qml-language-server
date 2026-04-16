package handler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseQMLLSIni(t *testing.T) {
	dir := t.TempDir()
	ini := filepath.Join(dir, ".qmlls.ini")
	content := `[General]
no-cmake-calls=true
buildDir="/tmp/quickshell/vfs/abc123"
importPaths="/usr/lib/qt6/qml:/opt/custom/qml"
`
	if err := os.WriteFile(ini, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseQMLLSIni(ini)
	if err != nil {
		t.Fatalf("ParseQMLLSIni: %v", err)
	}
	if cfg.BuildDir != "/tmp/quickshell/vfs/abc123" {
		t.Errorf("BuildDir = %q, want /tmp/quickshell/vfs/abc123", cfg.BuildDir)
	}
	if len(cfg.ImportPaths) != 2 {
		t.Fatalf("expected 2 import paths, got %d: %v", len(cfg.ImportPaths), cfg.ImportPaths)
	}
	if cfg.ImportPaths[0] != "/usr/lib/qt6/qml" {
		t.Errorf("ImportPaths[0] = %q", cfg.ImportPaths[0])
	}
	if cfg.ImportPaths[1] != "/opt/custom/qml" {
		t.Errorf("ImportPaths[1] = %q", cfg.ImportPaths[1])
	}
}

func TestParseQMLLSIniSkipsSectionsAndComments(t *testing.T) {
	dir := t.TempDir()
	ini := filepath.Join(dir, ".qmlls.ini")
	content := `# comment
[General]
buildDir="/some/path"
[Other]
foo=bar
`
	if err := os.WriteFile(ini, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseQMLLSIni(ini)
	if err != nil {
		t.Fatalf("ParseQMLLSIni: %v", err)
	}
	if cfg.BuildDir != "/some/path" {
		t.Errorf("BuildDir = %q", cfg.BuildDir)
	}
}

func TestFindAndParseQMLLSIni(t *testing.T) {
	dir := t.TempDir()
	ini := filepath.Join(dir, ".qmlls.ini")
	if err := os.WriteFile(ini, []byte(`[General]
buildDir="/build"
`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := FindAndParseQMLLSIni([]string{"/nonexistent", dir})
	if cfg == nil {
		t.Fatal("expected to find .qmlls.ini")
	}
	if cfg.BuildDir != "/build" {
		t.Errorf("BuildDir = %q", cfg.BuildDir)
	}
}

func TestFindAndParseQMLLSIniReturnsNilWhenMissing(t *testing.T) {
	cfg := FindAndParseQMLLSIni([]string{"/nonexistent"})
	if cfg != nil {
		t.Error("expected nil when no .qmlls.ini found")
	}
}
