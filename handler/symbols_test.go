package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestDocumentSymbolReturnsTopLevelObjectsAndImports(t *testing.T) {
	doc := "import QtQuick\nimport QtQuick.Controls\n\nRectangle {\n    id: root\n    width: 100\n    Text {\n        text: \"hi\"\n    }\n}\n"
	h := newTestHandler(t, "test://syms.qml", doc)

	syms, err := h.DocumentSymbol(context.Background(), &lsp.DocumentSymbolParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://syms.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol: %v", err)
	}
	if len(syms) == 0 {
		t.Fatal("expected at least one top-level symbol")
	}

	var imports, classes int
	for _, s := range syms {
		switch s.Kind {
		case lsp.SymbolKindModule:
			imports++
		case lsp.SymbolKindClass:
			classes++
		}
	}
	if imports != 2 {
		t.Errorf("expected 2 import symbols, got %d", imports)
	}
	if classes != 1 {
		t.Errorf("expected 1 top-level class symbol, got %d", classes)
	}
}

func TestDocumentSymbolNestsChildrenAndProperties(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n    height: 50\n    Text {\n        text: \"hi\"\n    }\n}\n"
	h := newTestHandler(t, "test://nested.qml", doc)

	syms, err := h.DocumentSymbol(context.Background(), &lsp.DocumentSymbolParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://nested.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol: %v", err)
	}

	var rect *lsp.DocumentSymbol
	for i := range syms {
		if syms[i].Kind == lsp.SymbolKindClass && syms[i].Name == "Rectangle" {
			rect = &syms[i]
			break
		}
	}
	if rect == nil {
		t.Fatal("expected a Rectangle class symbol")
	}

	wantNames := map[string]bool{"width": false, "height": false, "Text": false}
	for _, child := range rect.Children {
		if _, ok := wantNames[child.Name]; ok {
			wantNames[child.Name] = true
		}
	}
	for name, found := range wantNames {
		if !found {
			t.Errorf("expected child symbol %q under Rectangle", name)
		}
	}
}

func TestDocumentSymbolMarksSignalHandlerAsEvent(t *testing.T) {
	doc := "import QtQuick\n\nMouseArea {\n    onClicked: console.log(\"hi\")\n}\n"
	h := newTestHandler(t, "test://sig.qml", doc)

	syms, err := h.DocumentSymbol(context.Background(), &lsp.DocumentSymbolParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://sig.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol: %v", err)
	}
	var mouse *lsp.DocumentSymbol
	for i := range syms {
		if syms[i].Kind == lsp.SymbolKindClass && syms[i].Name == "MouseArea" {
			mouse = &syms[i]
			break
		}
	}
	if mouse == nil {
		t.Fatal("expected a MouseArea class symbol")
	}

	for _, child := range mouse.Children {
		if child.Name == "onClicked" {
			if child.Kind != lsp.SymbolKindEvent {
				t.Errorf("expected onClicked to be SymbolKindEvent, got %v", child.Kind)
			}
			return
		}
	}
	t.Error("did not find onClicked among MouseArea children")
}

func TestIsSignalHandler(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"onClicked", true},
		{"onTextChanged", true},
		{"on", false},
		{"only", false},
		{"onlowercase", false},
		{"width", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isSignalHandler(tc.in); got != tc.want {
			t.Errorf("isSignalHandler(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
