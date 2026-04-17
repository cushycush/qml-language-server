package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestBuildIDTypeIndexFindsNestedIds(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    id: root\n    Text {\n        id: label\n        text: \"hi\"\n    }\n}\n"
	h := newTestHandler(t, "test://idx.qml", doc)
	tree := h.parser.GetTree("test://idx.qml")
	if tree == nil {
		t.Fatal("no tree")
	}
	index := buildIDTypeIndex(tree.RootNode(), h.parser.Language(), []byte(doc))
	if got := index["root"]; got != "Rectangle" {
		t.Errorf("root type = %q, want Rectangle", got)
	}
	if got := index["label"]; got != "Text" {
		t.Errorf("label type = %q, want Text", got)
	}
}

func TestBuildIDTypeIndexIgnoresNonIdBindings(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n    color: \"red\"\n}\n"
	h := newTestHandler(t, "test://nid.qml", doc)
	tree := h.parser.GetTree("test://nid.qml")
	index := buildIDTypeIndex(tree.RootNode(), h.parser.Language(), []byte(doc))
	if len(index) != 0 {
		t.Errorf("expected empty index, got %v", index)
	}
}

func TestIdentifierBeforeDot(t *testing.T) {
	cases := []struct {
		text string
		pos  int
		want string
	}{
		{"    root.", 9, "root"},
		{"foo = root.", 11, "root"},
		{"anchors.fill.", 13, ""}, // multi-dot chain: rejected
		{"    .", 5, ""},          // nothing before dot
		{"    root", 8, ""},       // no dot at cursor
	}
	for _, tc := range cases {
		if got := identifierBeforeDot(tc.text, tc.pos); got != tc.want {
			t.Errorf("identifierBeforeDot(%q, %d) = %q, want %q", tc.text, tc.pos, got, tc.want)
		}
	}
}

func TestCompletionAfterIdDotReturnsTypeProperties(t *testing.T) {
	// Rectangle has `radius` as a type-specific property (via static
	// typeProperties). Pre-populate the typeProperties map so the test is
	// independent of whatever Qt modules happen to be installed.
	typeProperties["TestRectFoo"] = []QMLSymbol{
		{Label: "uniqueTestProp", Category: "property"},
	}
	baseTypes["TestRectFoo"] = nil
	defer delete(typeProperties, "TestRectFoo")
	defer delete(baseTypes, "TestRectFoo")

	doc := "import QtQuick\n\nTestRectFoo {\n    id: root\n    width: root.\n}\n"
	h := newTestHandler(t, "test://dot.qml", doc)

	list, err := h.Completion(context.Background(), &lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://dot.qml"},
			Position:     lsp.Position{Line: 4, Character: 16}, // just after `root.`
		},
	})
	if err != nil {
		t.Fatalf("Completion: %v", err)
	}
	found := false
	for _, item := range list.Items {
		if item.Label == "uniqueTestProp" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected uniqueTestProp from TestRectFoo in completions, got %d items", len(list.Items))
	}
}

func TestCompletionAfterUnknownDotFallsBack(t *testing.T) {
	// `foo.` where foo isn't an id should fall back to generic property
	// completions (e.g. `width`, `height`).
	doc := "import QtQuick\n\nRectangle {\n    width: foo.\n}\n"
	h := newTestHandler(t, "test://unk.qml", doc)

	list, err := h.Completion(context.Background(), &lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://unk.qml"},
			Position:     lsp.Position{Line: 3, Character: 15}, // line 3 = `    width: foo.`
		},
	})
	if err != nil {
		t.Fatalf("Completion: %v", err)
	}
	foundWidth := false
	for _, item := range list.Items {
		if item.Label == "width" {
			foundWidth = true
			break
		}
	}
	if !foundWidth {
		t.Errorf("expected generic `width` in fallback completions, got %d items", len(list.Items))
	}
}
