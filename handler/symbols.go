package handler

import (
	"context"
	"strings"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) DocumentSymbol(_ context.Context, params *lsp.DocumentSymbolParams) ([]lsp.DocumentSymbol, error) {
	doc, ok := h.getDocument(params.TextDocument.URI)
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

	return collectTopLevelSymbols(root, h.parser.Language(), []byte(doc)), nil
}

func collectTopLevelSymbols(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.DocumentSymbol {
	var symbols []lsp.DocumentSymbol
	walkTree(node, func(n *gotreesitter.Node) bool {
		switch n.Type(lang) {
		case "ui_object_definition":
			symbols = append(symbols, objectSymbol(n, lang, content))
			return false // don't recurse; objectSymbol gathers children
		case "ui_import":
			symbols = append(symbols, leafSymbol(n, lang, content, lsp.SymbolKindModule))
			return false
		}
		return true
	})
	return symbols
}

func objectSymbol(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) lsp.DocumentSymbol {
	sym := createSymbol(node, lang, content)
	sym.Kind = lsp.SymbolKindClass
	sym.Children = collectChildSymbols(node, lang, content)
	return sym
}

func leafSymbol(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, kind lsp.SymbolKind) lsp.DocumentSymbol {
	sym := createSymbol(node, lang, content)
	sym.Kind = kind
	return sym
}

func collectChildSymbols(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.DocumentSymbol {
	var symbols []lsp.DocumentSymbol

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type(lang) {
		case "ui_object_definition":
			symbols = append(symbols, objectSymbol(child, lang, content))
		case "ui_property", "ui_required":
			symbols = append(symbols, leafSymbol(child, lang, content, lsp.SymbolKindProperty))
		case "ui_binding":
			if s, ok := bindingSymbol(child, lang, content); ok {
				symbols = append(symbols, s)
			}
		}
	}

	return symbols
}

func bindingSymbol(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) (lsp.DocumentSymbol, bool) {
	name := extractPropertyName(node, lang, content)
	if name == "" {
		return lsp.DocumentSymbol{}, false
	}
	kind := lsp.SymbolKindProperty
	if isSignalHandler(name) {
		kind = lsp.SymbolKindEvent
	}
	nameEnd := node.StartByte() + uint32(len(name))
	return lsp.DocumentSymbol{
		Name:  name,
		Kind:  kind,
		Range: nodeRange(content, node),
		SelectionRange: lsp.Range{
			Start: byteOffsetToPosition(content, node.StartByte()),
			End:   byteOffsetToPosition(content, nameEnd),
		},
	}, true
}

func createSymbol(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) lsp.DocumentSymbol {
	name := extractNodeName(node, lang, content)
	nameEnd := node.StartByte() + uint32(len(name))
	return lsp.DocumentSymbol{
		Name:  name,
		Range: nodeRange(content, node),
		SelectionRange: lsp.Range{
			Start: byteOffsetToPosition(content, node.StartByte()),
			End:   byteOffsetToPosition(content, nameEnd),
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
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "identifier", "property_identifier", "nested_identifier":
			return string(content[child.StartByte():child.EndByte()])
		}
	}
	return ""
}

func isSignalHandler(name string) bool {
	return strings.HasPrefix(name, "on") && len(name) > 2 && name[2] >= 'A' && name[2] <= 'Z'
}
