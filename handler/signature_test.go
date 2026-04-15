package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestSignatureHelpForKnownCallReportsActiveParameter(t *testing.T) {
	// `Qt.rect(1, 2, |` — cursor immediately after the second comma puts us on
	// the third parameter (index 2).
	doc := "import QtQuick\n\nRectangle {\n    property var r: Qt.rect(1, 2, )\n}\n"
	h := newTestHandler(t, "test://sig.qml", doc)

	sig, err := h.SignatureHelp(context.Background(), &lsp.SignatureHelpParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://sig.qml"},
			Position:     lsp.Position{Line: 3, Character: 33},
		},
	})
	if err != nil {
		t.Fatalf("SignatureHelp: %v", err)
	}
	if sig == nil {
		t.Fatal("expected non-nil SignatureHelp")
	}
	if len(sig.Signatures) != 1 {
		t.Fatalf("expected one signature, got %d", len(sig.Signatures))
	}
	if !strings.Contains(sig.Signatures[0].Label, "Qt.rect") {
		t.Errorf("signature label = %q, want it to mention Qt.rect", sig.Signatures[0].Label)
	}
	if sig.ActiveParameter == nil || *sig.ActiveParameter != 2 {
		got := -1
		if sig.ActiveParameter != nil {
			got = *sig.ActiveParameter
		}
		t.Errorf("ActiveParameter = %d, want 2", got)
	}
}

func TestSignatureHelpUnknownCalleeReturnsNil(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    property var x: notARealFunction(\n}\n"
	h := newTestHandler(t, "test://sig2.qml", doc)

	sig, err := h.SignatureHelp(context.Background(), &lsp.SignatureHelpParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://sig2.qml"},
			Position:     lsp.Position{Line: 3, Character: 38},
		},
	})
	if err != nil {
		t.Fatalf("SignatureHelp: %v", err)
	}
	if sig != nil {
		t.Errorf("expected nil signature for unknown callee, got %+v", sig)
	}
}

func TestFindActiveCall(t *testing.T) {
	cases := []struct {
		name      string
		line      string
		cursor    int
		wantName  string
		wantArg   int
	}{
		{"first arg", "Qt.rect(", 8, "Qt.rect", 0},
		{"third arg", "Qt.rect(1, 2, ", 14, "Qt.rect", 2},
		{"nested call", "console.log(String(x), ", 23, "console.log", 1},
		{"no open paren", "no parens here", 14, "", 0},
		{"open brace before unmatched paren is ignored", "{ Qt.rect(1, 2, ", 16, "Qt.rect", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, arg := findActiveCall(tc.line, tc.cursor)
			if name != tc.wantName || arg != tc.wantArg {
				t.Errorf("findActiveCall(%q, %d) = (%q, %d), want (%q, %d)",
					tc.line, tc.cursor, name, arg, tc.wantName, tc.wantArg)
			}
		})
	}
}

func TestExtractCallee(t *testing.T) {
	cases := []struct {
		line     string
		parenIdx int
		want     string
	}{
		{"Qt.rect(", 7, "Qt.rect"},
		{"  console.log(", 13, "console.log"},
		{"foo(", 3, "foo"},
		{"(", 0, ""},
		{" String(", 7, "String"},
	}
	for _, tc := range cases {
		got := extractCallee(tc.line, tc.parenIdx)
		if got != tc.want {
			t.Errorf("extractCallee(%q, %d) = %q, want %q", tc.line, tc.parenIdx, got, tc.want)
		}
	}
}
