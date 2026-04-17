package handler

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestDocumentLinkQualifiedModule(t *testing.T) {
	// Point a fake module name at a real qmldir path so resolution succeeds
	// regardless of what Qt installations are present on the machine.
	tmp := t.TempDir()
	qmldir := filepath.Join(tmp, "qmldir")
	if err := os.WriteFile(qmldir, []byte("module FakeMod\n"), 0o644); err != nil {
		t.Fatalf("write qmldir: %v", err)
	}
	recordModuleQMLDir("FakeMod", qmldir)

	doc := "import FakeMod\n\nRectangle {}\n"
	h := newTestHandler(t, "test://mod.qml", doc)

	links, err := h.DocumentLink(context.Background(), &lsp.DocumentLinkParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://mod.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentLink: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	if links[0].Target == nil {
		t.Fatal("link target was nil")
	}
	if !strings.Contains(string(*links[0].Target), "qmldir") {
		t.Errorf("target %q should reference qmldir", *links[0].Target)
	}
	// Range should cover the module name on the first line, not the `import`
	// keyword or the whole line.
	if links[0].Range.Start.Line != 0 || links[0].Range.Start.Character != 7 {
		t.Errorf("range start = %+v, want line 0 char 7", links[0].Range.Start)
	}
	if links[0].Range.End.Character != 14 {
		t.Errorf("range end char = %d, want 14 (end of 'FakeMod')", links[0].Range.End.Character)
	}
}

func TestDocumentLinkDottedModuleFallsBack(t *testing.T) {
	// `import QtQuick.Controls` when only QtQuick is registered should still
	// resolve, by stripping the trailing segment.
	tmp := t.TempDir()
	qmldir := filepath.Join(tmp, "qmldir")
	if err := os.WriteFile(qmldir, []byte("module FallbackMod\n"), 0o644); err != nil {
		t.Fatalf("write qmldir: %v", err)
	}
	recordModuleQMLDir("FallbackMod", qmldir)

	doc := "import FallbackMod.Sub\n\nRectangle {}\n"
	h := newTestHandler(t, "test://fallback.qml", doc)

	links, err := h.DocumentLink(context.Background(), &lsp.DocumentLinkParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://fallback.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentLink: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	if links[0].Target == nil || !strings.Contains(string(*links[0].Target), "qmldir") {
		t.Errorf("target %v should reference qmldir", links[0].Target)
	}
}

func TestDocumentLinkStringImportRelativeDir(t *testing.T) {
	tmp := t.TempDir()
	comp := filepath.Join(tmp, "components")
	if err := os.MkdirAll(comp, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(comp, "qmldir"), []byte("module local\n"), 0o644); err != nil {
		t.Fatalf("write qmldir: %v", err)
	}

	mainPath := filepath.Join(tmp, "main.qml")
	if err := os.WriteFile(mainPath, []byte(""), 0o644); err != nil {
		t.Fatalf("write main: %v", err)
	}
	uri := lsp.DocumentURI((&url.URL{Scheme: "file", Path: filepath.ToSlash(mainPath)}).String())

	doc := "import \"./components\"\n\nRectangle {}\n"
	h := newTestHandler(t, uri, doc)

	links, err := h.DocumentLink(context.Background(), &lsp.DocumentLinkParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("DocumentLink: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	if links[0].Target == nil {
		t.Fatal("link target was nil")
	}
	if !strings.HasSuffix(string(*links[0].Target), "components/qmldir") {
		t.Errorf("target %q should end with components/qmldir", *links[0].Target)
	}
}

func TestDocumentLinkUnknownModuleDropped(t *testing.T) {
	doc := "import ThisDefinitelyDoesNotExistAnywhere\n\nRectangle {}\n"
	h := newTestHandler(t, "test://unknown.qml", doc)

	links, err := h.DocumentLink(context.Background(), &lsp.DocumentLinkParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://unknown.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentLink: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected no links for unresolvable module, got %d", len(links))
	}
}

func TestResolveImportTargetStringPrefersQMLDir(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "qmldir"), []byte(""), 0o644); err != nil {
		t.Fatalf("write qmldir: %v", err)
	}

	target, _, ok := resolveImportTarget("string", "\".\"", tmp)
	if !ok {
		t.Fatal("expected resolution to succeed")
	}
	if filepath.Base(target) != "qmldir" {
		t.Errorf("target = %q, want qmldir", target)
	}
}
