package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestDefinitionResolvesImplicitKeyword(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    Text {\n        anchors.fill: parent\n    }\n}\n"
	h := newTestHandler(t, "test://parent.qml", doc)

	// Cursor on `parent` (line 4, col 22). Implicit keywords (`parent`, `this`,
	// `root`) resolve to the nearest enclosing ui_object_definition. The
	// keyword is a child of a ui_binding, so the immediate enclosing object —
	// Text on line 3 — is what gets returned.
	locs, err := h.Definition(context.Background(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://parent.qml"},
			Position:     lsp.Position{Line: 4, Character: 22},
		},
	})
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("expected one location, got %d", len(locs))
	}
	if locs[0].Range.Start.Line != 3 {
		t.Errorf("expected parent to point to enclosing Text on line 3, got line %d", locs[0].Range.Start.Line)
	}
}

func TestDefinitionResolvesIdReference(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    id: root\n    Text {\n        color: root.color\n    }\n}\n"
	h := newTestHandler(t, "test://id.qml", doc)

	// Cursor on `root` inside the binding `id: root` (line 3).
	locs, err := h.Definition(context.Background(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://id.qml"},
			Position:     lsp.Position{Line: 3, Character: 9},
		},
	})
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("expected one location for id binding, got %d", len(locs))
	}
}

func TestDefinitionOnUnknownReturnsEmpty(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n}\n"
	h := newTestHandler(t, "test://unknown.qml", doc)

	// Position on whitespace — node is not an identifier, no definition.
	locs, err := h.Definition(context.Background(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://unknown.qml"},
			Position:     lsp.Position{Line: 1, Character: 0},
		},
	})
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	if len(locs) != 0 {
		t.Errorf("expected no locations, got %d", len(locs))
	}
}

func TestExtractIdFromBinding(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"id: root", "root"},
		{"id : foo", "foo"},
		{"width: 100", ""},
		{"no colon here", ""},
	}
	for _, tc := range cases {
		if got := extractIdFromBinding(tc.in); got != tc.want {
			t.Errorf("extractIdFromBinding(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
