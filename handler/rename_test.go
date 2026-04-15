package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestPrepareRenameOnIdentifier(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    id: root\n    width: root.height\n}\n"
	h := newTestHandler(t, "test://rename.qml", doc)

	res, err := h.PrepareRename(context.Background(), &lsp.PrepareRenameParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://rename.qml"},
			Position:     lsp.Position{Line: 3, Character: 9},
		},
	})
	if err != nil {
		t.Fatalf("PrepareRename: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil PrepareRenameResult on identifier")
	}
	if res.Placeholder != "root" {
		t.Errorf("placeholder = %q, want %q", res.Placeholder, "root")
	}
}

func TestPrepareRenameOnNonIdentifierIsNil(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n}\n"
	h := newTestHandler(t, "test://rename2.qml", doc)

	res, err := h.PrepareRename(context.Background(), &lsp.PrepareRenameParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://rename2.qml"},
			Position:     lsp.Position{Line: 3, Character: 13},
		},
	})
	if err != nil {
		t.Fatalf("PrepareRename: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result on number literal, got %+v", res)
	}
}

func TestRenameProducesEditsAtAllOccurrences(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    id: root\n    width: root.height\n    Text {\n        text: root.color\n    }\n}\n"
	h := newTestHandler(t, "test://rename3.qml", doc)

	edit, err := h.Rename(context.Background(), &lsp.RenameParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://rename3.qml"},
			Position:     lsp.Position{Line: 3, Character: 9},
		},
		NewName: "container",
	})
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if edit == nil {
		t.Fatal("expected non-nil WorkspaceEdit")
	}
	edits, ok := edit.Changes["test://rename3.qml"]
	if !ok {
		t.Fatal("expected edits for the document URI")
	}
	if len(edits) < 3 {
		t.Errorf("expected at least 3 edits for `root`, got %d", len(edits))
	}
	for _, e := range edits {
		if e.NewText != "container" {
			t.Errorf("edit NewText = %q, want %q", e.NewText, "container")
		}
	}
}
