package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestFormattingReindentsBracedBlocks(t *testing.T) {
	in := "import QtQuick\nRectangle {\nwidth: 100\nText {\ntext: \"hi\"\n}\n}\n"
	want := "import QtQuick\nRectangle {\n    width: 100\n    Text {\n        text: \"hi\"\n    }\n}\n"

	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: true})
	if got != want {
		t.Errorf("formatQML mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormattingTrimsTrailingWhitespace(t *testing.T) {
	in := "Rectangle {   \n    width: 100   \n}   \n"
	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: true})
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			t.Errorf("trailing whitespace on line %q", line)
		}
	}
}

func TestFormattingCollapsesBlankRuns(t *testing.T) {
	in := "Rectangle {\n\n\n\n    width: 100\n\n\n}\n"
	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: true})
	// Allow at most one blank line between content lines.
	saw := 0
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n") {
		if line == "" {
			saw++
			if saw > 1 {
				t.Errorf("found a run of more than one blank line in:\n%s", got)
			}
		} else {
			saw = 0
		}
	}
}

func TestFormattingHonoursTabsOption(t *testing.T) {
	in := "Rectangle {\nwidth: 100\n}\n"
	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: false})
	if !strings.Contains(got, "\twidth: 100") {
		t.Errorf("expected tab-indented body, got:\n%s", got)
	}
}

func TestFormattingIgnoresBracesInsideStringsAndComments(t *testing.T) {
	in := "Rectangle {\ntext: \"a } b\"\n// comment with }\n/* block } closed */\nwidth: 100\n}\n"
	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: true})

	// All four interior lines should be indented exactly one level (4 spaces).
	for _, line := range strings.Split(strings.TrimRight(got, "\n"), "\n")[1:5] {
		if !strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "        ") {
			t.Errorf("line %q expected single-level indent", line)
		}
	}
}

func TestFormattingPreservesSemanticContent(t *testing.T) {
	in := "Rectangle {\nwidth: 100\ntext: \"hello world\"\n}\n"
	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: true})

	// Check that every non-whitespace token in the input survives.
	for _, want := range []string{"Rectangle", "width: 100", "text: \"hello world\""} {
		if !strings.Contains(got, want) {
			t.Errorf("output is missing %q:\n%s", want, got)
		}
	}
}

func TestFormattingEnsuresFinalNewline(t *testing.T) {
	in := "Rectangle {\n    width: 100\n}"
	got := formatQML(in, lsp.FormattingOptions{TabSize: 4, InsertSpaces: true})
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("output missing trailing newline: %q", got)
	}
	if strings.HasSuffix(got, "\n\n") {
		t.Errorf("output has extra trailing newline: %q", got)
	}
}

func TestFormattingNoChangeReturnsEmptyEdits(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n}\n"
	h := newTestHandler(t, "test://fmt.qml", doc)

	edits, err := h.Formatting(context.Background(), &lsp.DocumentFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://fmt.qml"},
		Options:      lsp.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	if err != nil {
		t.Fatalf("Formatting: %v", err)
	}
	if len(edits) != 0 {
		t.Errorf("already-formatted doc should produce no edits, got %d", len(edits))
	}
}

func TestFormattingHandlerReturnsFullReplacement(t *testing.T) {
	doc := "import QtQuick\nRectangle {\nwidth: 100\n}\n"
	h := newTestHandler(t, "test://fmt2.qml", doc)

	edits, err := h.Formatting(context.Background(), &lsp.DocumentFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://fmt2.qml"},
		Options:      lsp.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	if err != nil {
		t.Fatalf("Formatting: %v", err)
	}
	if len(edits) != 1 {
		t.Fatalf("expected exactly one edit (full replacement), got %d", len(edits))
	}
	if !strings.Contains(edits[0].NewText, "    width: 100") {
		t.Errorf("expected width to be indented; got:\n%s", edits[0].NewText)
	}
}
