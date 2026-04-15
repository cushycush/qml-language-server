package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestReferencesFindsAllOccurrences(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    id: root\n    width: root.height\n    Text {\n        text: root.x\n    }\n}\n"
	h := newTestHandler(t, "test://refs.qml", doc)

	// Cursor on the second `root` (line 4, column 11).
	locs, err := h.References(context.Background(), &lsp.ReferenceParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://refs.qml"},
			Position:     lsp.Position{Line: 4, Character: 11},
		},
	})
	if err != nil {
		t.Fatalf("References: %v", err)
	}
	if len(locs) < 3 {
		t.Errorf("expected at least 3 occurrences of `root`, got %d", len(locs))
	}
	for _, loc := range locs {
		if loc.URI != "test://refs.qml" {
			t.Errorf("unexpected URI %q in references", loc.URI)
		}
	}
}

func TestReferencesOnNonIdentifierReturnsNil(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n}\n"
	h := newTestHandler(t, "test://refs2.qml", doc)

	// Position on the `100` literal — not an identifier.
	locs, err := h.References(context.Background(), &lsp.ReferenceParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://refs2.qml"},
			Position:     lsp.Position{Line: 3, Character: 13},
		},
	})
	if err != nil {
		t.Fatalf("References: %v", err)
	}
	if len(locs) != 0 {
		t.Errorf("expected no references on a literal, got %d", len(locs))
	}
}
