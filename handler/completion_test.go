package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestDetectCompletionContext(t *testing.T) {
	tests := []struct {
		name string
		text string
		pos  int
		want CompletionContext
	}{
		{"import with trailing dot", "import QtQuick.", 15, ContextImport},
		{"import with module and version", "import QtQuick.Controls 2.", 26, ContextImport},
		{"bare import keyword", "import ", 7, ContextImport},
		{"property dot chain", "anchors.", 8, ContextProperty},
		{"value side after colon", "color: ", 7, ContextAfterColon},
		{"empty line is default", "", 0, ContextDefault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectCompletionContext(tt.text, tt.pos); got != tt.want {
				t.Fatalf("detectCompletionContext(%q, %d) = %v, want %v", tt.text, tt.pos, got, tt.want)
			}
		})
	}
}

// Regression: issue #25 — completion items were being added by both the
// AST-driven branch and the context-switch fallback, so clients displayed
// every type and keyword twice. The handler now dedupes by Label.
func TestCompletionNoDuplicateLabels(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		line int
		ch   int
	}{
		{
			name: "type position inside object body",
			doc:  "import QtQuick\n\nItem {\n    Sca\n}\n",
			line: 3,
			ch:   7,
		},
		{
			name: "blank line inside object body",
			doc:  "import QtQuick\n\nItem {\n    \n}\n",
			line: 3,
			ch:   4,
		},
		{
			name: "import position",
			doc:  "import \n",
			line: 0,
			ch:   7,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(t, "test://dup.qml", tc.doc)
			res, err := h.Completion(context.Background(), &lsp.CompletionParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: "test://dup.qml"},
					Position:     lsp.Position{Line: tc.line, Character: tc.ch},
				},
			})
			if err != nil {
				t.Fatalf("Completion: %v", err)
			}
			seen := map[string]int{}
			for _, item := range res.Items {
				seen[item.Label]++
			}
			for label, n := range seen {
				if n > 1 {
					t.Errorf("label %q appeared %d times, want 1", label, n)
				}
			}
		})
	}
}

func TestDedupeCompletionItems(t *testing.T) {
	first := lsp.CompletionItem{Label: "Item", Detail: "from AST"}
	dup := lsp.CompletionItem{Label: "Item", Detail: "from fallback"}
	other := lsp.CompletionItem{Label: "Rectangle"}

	got := dedupeCompletionItems([]lsp.CompletionItem{first, dup, other})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Detail != "from AST" {
		t.Errorf("first occurrence not preserved: got Detail=%q, want %q", got[0].Detail, "from AST")
	}
	if got[1].Label != "Rectangle" {
		t.Errorf("got[1].Label = %q, want Rectangle", got[1].Label)
	}
}
