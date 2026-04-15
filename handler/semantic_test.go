package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestSemanticTokensLegendOrderIsStable(t *testing.T) {
	legend := SemanticTokensLegend()
	if got := legend.TokenTypes[tokTypeNamespace]; got != "namespace" {
		t.Errorf("legend[tokTypeNamespace] = %q, want namespace", got)
	}
	if got := legend.TokenTypes[tokTypeProperty]; got != "property" {
		t.Errorf("legend[tokTypeProperty] = %q, want property", got)
	}
	if got := legend.TokenTypes[tokTypeEvent]; got != "event" {
		t.Errorf("legend[tokTypeEvent] = %q, want event", got)
	}
	if got := legend.TokenModifiers[0]; got != "declaration" {
		t.Errorf("legend.Modifiers[0] = %q, want declaration", got)
	}
}

func TestEncodeSemanticTokensDeltaEncodesPositions(t *testing.T) {
	tokens := []rawToken{
		{Line: 0, Char: 0, Length: 6, TokenType: tokTypeKeyword},
		{Line: 0, Char: 7, Length: 7, TokenType: tokTypeNamespace},
		{Line: 2, Char: 0, Length: 9, TokenType: tokTypeType, Modifiers: tokModDefaultLibrary},
		{Line: 3, Char: 4, Length: 5, TokenType: tokTypeProperty},
	}
	want := []int{
		0, 0, 6, tokTypeKeyword, 0,
		0, 7, 7, tokTypeNamespace, 0,
		2, 0, 9, tokTypeType, tokModDefaultLibrary,
		1, 4, 5, tokTypeProperty, 0,
	}
	got := encodeSemanticTokens(tokens)
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d, len(want)=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("data[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestEncodeSemanticTokensSortsBeforeEncoding(t *testing.T) {
	tokens := []rawToken{
		{Line: 1, Char: 5, Length: 3, TokenType: tokTypeNumber},
		{Line: 0, Char: 0, Length: 6, TokenType: tokTypeKeyword},
	}
	got := encodeSemanticTokens(tokens)
	if got[0] != 0 || got[1] != 0 {
		t.Errorf("first token should start at line 0 char 0, got delta-line %d delta-char %d", got[0], got[1])
	}
}

func TestSemanticTokensFullProducesTokensForCommonNodes(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n    Text {\n        text: \"hello\"\n    }\n}\n"
	h := newTestHandler(t, "test://sem.qml", doc)

	res, err := h.SemanticTokensFull(context.Background(), &lsp.SemanticTokensParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://sem.qml"},
	})
	if err != nil {
		t.Fatalf("SemanticTokensFull: %v", err)
	}
	if res == nil || len(res.Data) == 0 {
		t.Fatal("expected non-empty token data")
	}
	if len(res.Data)%5 != 0 {
		t.Fatalf("token data length must be a multiple of 5, got %d", len(res.Data))
	}

	// Decode and bucket the tokens by type to assert categories appeared.
	seen := map[int]int{}
	line, char := 0, 0
	for i := 0; i < len(res.Data); i += 5 {
		dLine := res.Data[i]
		dChar := res.Data[i+1]
		tokenType := res.Data[i+3]
		if dLine == 0 {
			char += dChar
		} else {
			line += dLine
			char = dChar
		}
		seen[tokenType]++
		_ = line
	}
	mustHave := []int{tokTypeKeyword, tokTypeNamespace, tokTypeType, tokTypeProperty, tokTypeNumber, tokTypeString}
	for _, want := range mustHave {
		if seen[want] == 0 {
			t.Errorf("expected at least one token of type %d, got none", want)
		}
	}
}

func TestSemanticTokensFullReturnsEmptyForUnknownDoc(t *testing.T) {
	h := New(nil)
	if h.parser == nil {
		t.Skip("parser unavailable")
	}
	res, err := h.SemanticTokensFull(context.Background(), &lsp.SemanticTokensParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://missing.qml"},
	})
	if err != nil {
		t.Fatalf("SemanticTokensFull: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response, got nil")
	}
	if len(res.Data) != 0 {
		t.Errorf("expected empty data, got %v", res.Data)
	}
}
