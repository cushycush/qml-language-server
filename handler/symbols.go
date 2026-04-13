package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) DocumentSymbol(_ context.Context, params *lsp.DocumentSymbolParams) ([]lsp.DocumentSymbol, error) {
	doc, ok := h.documents[params.TextDocument.URI]
	if !ok || h.parser == nil {
		return nil, nil
	}

	tree := h.parser.GetTree(params.TextDocument.URI)
	if tree == nil {
		return nil, nil
	}

	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

	lang := h.parser.Language()
	content := []byte(doc)

	return collectDocumentSymbols(root, lang, content), nil
}

func collectDocumentSymbols(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.DocumentSymbol {
	var symbols []lsp.DocumentSymbol

	if node == nil {
		return symbols
	}

	nodeType := node.Type(lang)

	switch nodeType {
	case "ui_object_definition":
		symbol := createDocumentSymbol(node, lang, content)
		symbol.Children = collectChildSymbols(node, lang, content)
		symbol.Kind = lsp.SymbolKindClass
		symbols = append(symbols, symbol)

	case "ui_import":
		symbol := createDocumentSymbol(node, lang, content)
		symbol.Kind = lsp.SymbolKindModule
		symbols = append(symbols, symbol)
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			symbols = append(symbols, collectDocumentSymbols(child, lang, content)...)
		}
	}

	return symbols
}

func collectChildSymbols(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.DocumentSymbol {
	var symbols []lsp.DocumentSymbol

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		childType := child.Type(lang)

		switch childType {
		case "ui_object_definition":
			symbol := createDocumentSymbol(child, lang, content)
			symbol.Children = collectChildSymbols(child, lang, content)
			symbol.Kind = lsp.SymbolKindClass
			symbols = append(symbols, symbol)

		case "ui_property":
			symbol := createDocumentSymbol(child, lang, content)
			symbol.Kind = lsp.SymbolKindProperty
			symbols = append(symbols, symbol)

		case "ui_required":
			symbol := createDocumentSymbol(child, lang, content)
			symbol.Kind = lsp.SymbolKindProperty
			symbols = append(symbols, symbol)

		case "ui_binding":
			propName := extractPropertyName(child, lang, content)
			if propName != "" {
				kind := lsp.SymbolKindProperty
				if isSignalHandler(propName) {
					kind = lsp.SymbolKindEvent
				}
				symbol := lsp.DocumentSymbol{
					Name: propName,
					Kind: kind,
					Range: lsp.Range{
						Start: lsp.Position{Line: 0, Character: int(child.StartByte())},
						End:   lsp.Position{Line: 0, Character: int(child.EndByte())},
					},
					SelectionRange: lsp.Range{
						Start: lsp.Position{Line: 0, Character: int(child.StartByte())},
						End:   lsp.Position{Line: 0, Character: int(child.StartByte() + uint32(len(propName)))},
					},
				}
				symbols = append(symbols, symbol)
			}

		case "comment":
			commentText := string(content[child.StartByte():child.EndByte()])
			symbol := lsp.DocumentSymbol{
				Name: truncateString(commentText, 50),
				Kind: lsp.SymbolKindNamespace,
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(child.StartByte())},
					End:   lsp.Position{Line: 0, Character: int(child.EndByte())},
				},
				SelectionRange: lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(child.StartByte())},
					End:   lsp.Position{Line: 0, Character: int(child.EndByte())},
				},
			}
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

func createDocumentSymbol(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) lsp.DocumentSymbol {
	name := extractNodeName(node, lang, content)
	return lsp.DocumentSymbol{
		Name: name,
		Kind: lsp.SymbolKindClass,
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: int(node.StartByte())},
			End:   lsp.Position{Line: 0, Character: int(node.EndByte())},
		},
		SelectionRange: lsp.Range{
			Start: lsp.Position{Line: 0, Character: int(node.StartByte())},
			End:   lsp.Position{Line: 0, Character: int(node.StartByte() + uint32(len(name)))},
		},
		Children: []lsp.DocumentSymbol{},
	}
}

func extractNodeName(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Type(lang) == "identifier" {
			return string(content[child.StartByte():child.EndByte()])
		}
	}
	return "Unknown"
}

func extractPropertyName(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			childType := child.Type(lang)
			if childType == "identifier" || childType == "property_identifier" {
				return string(content[child.StartByte():child.EndByte()])
			}
			if childType == "nested_identifier" {
				return string(content[child.StartByte():child.EndByte()])
			}
		}
	}
	return ""
}

func isSignalHandler(name string) bool {
	return len(name) > 2 && name[:2] == "on"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
