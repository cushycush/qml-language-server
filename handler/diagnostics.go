package handler

import (
	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

// collectDiagnostics walks the tree for ERROR and MISSING nodes and reports
// them as syntax errors.
func collectDiagnostics(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, diagnostics *[]lsp.Diagnostic) {
	walkTree(node, func(n *gotreesitter.Node) bool {
		if n.IsError() || n.IsMissing() {
			severity := lsp.SeverityError
			msg := "Syntax error"
			if n.IsMissing() {
				msg = "Missing " + n.Type(lang)
			}
			*diagnostics = append(*diagnostics, lsp.Diagnostic{
				Range:    nodeRange(content, n),
				Severity: &severity,
				Message:  msg,
				Source:   "qml-language-server",
			})
		}
		return true
	})
}
