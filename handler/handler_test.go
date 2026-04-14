package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestCountParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"no params", "Qt.rect()", 0},
		{"one param", "Qt.rect(100)", 0},
		{"two params", "Qt.rect(100, 200)", 1},
		{"three params", "Qt.rect(100, 200, 50)", 2},
		{"params with strings", "console.log(\"hello\", \"world\")", 1},
		{"nested parens", "String(value)", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countParams(tt.input)
			if result != tt.expected {
				t.Errorf("countParams(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsColorValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"red", "red", true},
		{"green", "green", true},
		{"blue", "blue", true},
		{"white", "white", true},
		{"black", "black", true},
		{"yellow", "yellow", true},
		{"cyan", "cyan", true},
		{"magenta", "magenta", true},
		{"gray", "gray", true},
		{"grey", "grey", true},
		{"transparent", "transparent", true},
		{"purple", "purple", false},
		{"orange", "orange", false},
		{"empty", "", false},
		{"uppercase", "RED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isColorValue(tt.input)
			if result != tt.expected {
				t.Errorf("isColorValue(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsQuotedString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"double quoted", "\"hello\"", true},
		{"single quoted", "'hello'", false},
		{"unquoted", "hello", false},
		{"partial quote", "\"hello", false},
		{"empty quotes", "\"\"", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isQuotedString(tt.input)
			if result != tt.expected {
				t.Errorf("isQuotedString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "hello", "hello"},
		{"empty string", "", "<empty>"},
		{"spaces only", "   ", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeString(tt.input)
			if result != tt.expected {
				t.Errorf("safeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSafeSliceLen(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{"normal slice", []string{"a", "b", "c"}, 3},
		{"empty slice", []string{}, 0},
		{"nil slice", nil, 0},
		{"single element", []string{"a"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeSliceLen(tt.input)
			if result != tt.expected {
				t.Errorf("safeSliceLen(%v) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHandlerError(t *testing.T) {
	err := newHandlerError("TEST_ERROR", "test message", nil)

	if err.Code != "TEST_ERROR" {
		t.Errorf("err.Code = %q, want %q", err.Code, "TEST_ERROR")
	}

	if err.Message != "test message" {
		t.Errorf("err.Message = %q, want %q", err.Message, "test message")
	}

	if err.Error() != "TEST_ERROR: test message" {
		t.Errorf("err.Error() = %q, want %q", err.Error(), "TEST_ERROR: test message")
	}
}

func TestHandlerErrorWithCause(t *testing.T) {
	cause := newHandlerError("CAUSE", "cause message", nil)
	err := newHandlerError("WRAPPER", "wrapper message", cause)

	if err.Unwrap() != cause {
		t.Errorf("err.Unwrap() = %v, want %v", err.Unwrap(), cause)
	}

	expected := "WRAPPER: wrapper message (CAUSE: cause message)"
	if err.Error() != expected {
		t.Errorf("err.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestQMLTypeInfo(t *testing.T) {
	tests := []struct {
		name       string
		typeName   string
		wantOK     bool
		wantModule string
		wantType   string
	}{
		{"Rectangle", "Rectangle", true, "QtQuick", "Object"},
		{"Text", "Text", true, "QtQuick", "Object"},
		{"Item", "Item", true, "QtQuick", "Object"},
		{"ColumnLayout", "ColumnLayout", true, "QtQuick.Layouts", "Object"},
		{"ListElement", "ListElement", true, "QtQml.Models", "Object"},
		{"Unknown", "UnknownType", false, "", ""},
		{"Empty", "", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := getTypeInfo(tt.typeName)
			if ok != tt.wantOK {
				t.Errorf("getTypeInfo(%q) ok = %v, want %v", tt.typeName, ok, tt.wantOK)
			}
			if ok {
				if info.Module != tt.wantModule {
					t.Errorf("getTypeInfo(%q).Module = %q, want %q", tt.typeName, info.Module, tt.wantModule)
				}
				if info.Type != tt.wantType {
					t.Errorf("getTypeInfo(%q).Type = %q, want %q", tt.typeName, info.Type, tt.wantType)
				}
			}
		})
	}
}

func TestHoverOnKnownType(t *testing.T) {
	h := newTestHandler(t, "test://foo.qml", "import QtQuick\n\nRectangle {\n    width: 100\n}\n")

	cases := []struct {
		name     string
		pos      lsp.Position
		wantText string
	}{
		{"on type name Rectangle", lsp.Position{Line: 2, Character: 3}, "Rectangle"},
		{"at end of Rectangle", lsp.Position{Line: 2, Character: 9}, "Rectangle"},
		{"on property width", lsp.Position{Line: 3, Character: 6}, "width"},
		{"on import module", lsp.Position{Line: 0, Character: 10}, "QtQuick"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Hover(context.Background(), &lsp.HoverParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: "test://foo.qml"},
					Position:     tt.pos,
				},
			})
			if err != nil {
				t.Fatalf("Hover returned error: %v", err)
			}
			if got == nil {
				t.Fatal("Hover returned nil")
			}
			if !strings.Contains(got.Contents.Value, tt.wantText) {
				t.Errorf("Hover value %q does not mention %q", got.Contents.Value, tt.wantText)
			}
			if got.Range == nil {
				t.Error("Hover range is nil")
			}
		})
	}
}

func TestCompletionPopulatesDocumentation(t *testing.T) {
	h := newTestHandler(t, "test://foo.qml", "import QtQuick\n\nRectangle {\n    \n}\n")

	list, err := h.Completion(context.Background(), &lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "test://foo.qml"},
			Position:     lsp.Position{Line: 3, Character: 4},
		},
	})
	if err != nil {
		t.Fatalf("Completion returned error: %v", err)
	}
	if list == nil || len(list.Items) == 0 {
		t.Fatal("Completion returned no items")
	}
	// Every item we emit should ship with Detail or Documentation.
	for _, item := range list.Items {
		if item.Documentation == nil && item.Detail == "" {
			t.Errorf("Completion item %q has neither Detail nor Documentation", item.Label)
		}
	}
}

func newTestHandler(t *testing.T, uri lsp.DocumentURI, text string) *Handler {
	t.Helper()
	h := New(nil)
	if h.parser == nil {
		t.Skip("parser unavailable in test environment")
	}
	if err := h.DidOpen(context.Background(), &lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{URI: uri, Text: text},
	}); err != nil {
		t.Fatalf("DidOpen: %v", err)
	}
	return h
}

func TestQMLPropertyInfo(t *testing.T) {
	tests := []struct {
		name     string
		propName string
		wantOK   bool
		wantType string
	}{
		{"width", "width", true, "real"},
		{"color", "color", true, "color"},
		{"text", "text", true, "string"},
		{"visible", "visible", true, "bool"},
		{"id", "id", true, "string"},
		{"onClicked", "onClicked", true, "signal"},
		{"Unknown", "unknownProp", false, ""},
		{"Empty", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := getPropertyInfo(tt.propName)
			if ok != tt.wantOK {
				t.Errorf("getPropertyInfo(%q) ok = %v, want %v", tt.propName, ok, tt.wantOK)
			}
			if ok && info.Type != tt.wantType {
				t.Errorf("getPropertyInfo(%q).Type = %q, want %q", tt.propName, info.Type, tt.wantType)
			}
		})
	}
}
