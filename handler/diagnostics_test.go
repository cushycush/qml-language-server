package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func TestDiagnosticsCleanDocumentReturnsEmpty(t *testing.T) {
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n}\n"
	h := newTestHandler(t, "test://clean.qml", doc)

	report, err := h.DocumentDiagnostic(context.Background(), &lsp.DocumentDiagnosticParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://clean.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentDiagnostic: %v", err)
	}
	full, ok := report.(lsp.FullDocumentDiagnosticReport)
	if !ok {
		t.Fatalf("expected FullDocumentDiagnosticReport, got %T", report)
	}
	if full.Items == nil {
		t.Error("Items must be normalized to [], not nil")
	}
	if len(full.Items) != 0 {
		t.Errorf("expected no diagnostics on clean doc, got %d", len(full.Items))
	}
}

func TestDiagnosticsReportsSyntaxError(t *testing.T) {
	// Missing closing brace.
	doc := "import QtQuick\n\nRectangle {\n    width: 100\n"
	h := newTestHandler(t, "test://broken.qml", doc)

	report, err := h.DocumentDiagnostic(context.Background(), &lsp.DocumentDiagnosticParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://broken.qml"},
	})
	if err != nil {
		t.Fatalf("DocumentDiagnostic: %v", err)
	}
	full, ok := report.(lsp.FullDocumentDiagnosticReport)
	if !ok {
		t.Fatalf("expected FullDocumentDiagnosticReport, got %T", report)
	}
	if len(full.Items) == 0 {
		t.Fatal("expected at least one diagnostic for broken doc")
	}
	for _, d := range full.Items {
		if d.Source != "qml-language-server" {
			t.Errorf("unexpected diagnostic source %q", d.Source)
		}
		if d.Severity == nil || *d.Severity != lsp.SeverityError {
			t.Errorf("expected error severity, got %v", d.Severity)
		}
	}
}
