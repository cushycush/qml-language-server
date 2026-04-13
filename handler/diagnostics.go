package handler

import (
	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func collectDiagnostics(node *gotreesitter.Node, lang *gotreesitter.Language, diagnostics *[]lsp.Diagnostic) {
	if node == nil {
		return
	}

	if node.Type(lang) == "ERROR" {
		severity := lsp.SeverityError
		*diagnostics = append(*diagnostics, lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: int(node.StartByte())},
				End:   lsp.Position{Line: 0, Character: int(node.EndByte())},
			},
			Severity: &severity,
			Message:  "Syntax error",
			Source:   "qml-language-server",
		})
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			collectDiagnostics(child, lang, diagnostics)
		}
	}
}
