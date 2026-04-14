package handler

import (
	"context"
	"strings"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Definition(_ context.Context, params *lsp.DefinitionParams) ([]lsp.Location, error) {
	uri := params.TextDocument.URI
	doc, ok := h.getDocument(uri)
	if !ok || h.parser == nil {
		return nil, nil
	}

	tree := h.parser.GetTree(uri)
	if tree == nil {
		return nil, nil
	}
	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

	lang := h.parser.Language()
	content := []byte(doc)
	offset := positionToByte(content, params.Position)
	node := findSmallestNodeAt(root, offset, lang)
	if node == nil {
		return nil, nil
	}

	return findDefinition(node, lang, content, root, uri), nil
}

func findDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node, uri lsp.DocumentURI) []lsp.Location {
	nodeType := node.Type(lang)

	switch nodeType {
	case "identifier":
		return findIdentifierDefinition(node, lang, content, root, uri)
	case "ui_object_definition":
		return findComponentDefinition(node, lang, content, root, uri)
	}
	return nil
}

func findIdentifierDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node, uri lsp.DocumentURI) []lsp.Location {
	parent := node.Parent()
	if parent == nil {
		return nil
	}

	text := string(content[node.StartByte():node.EndByte()])

	switch parent.Type(lang) {
	case "nested_identifier":
		if target := findKeywordTarget(text, node, lang, root); target != nil {
			return []lsp.Location{nodeLocation(uri, content, target)}
		}
	case "ui_binding":
		if target := findKeywordTarget(text, node, lang, root); target != nil {
			return []lsp.Location{nodeLocation(uri, content, target)}
		}
	case "expression_statement":
		if target := findKeywordTarget(text, node, lang, root); target != nil {
			return []lsp.Location{nodeLocation(uri, content, target)}
		}
	case "ui_object_definition":
		return findComponentDefinition(node, lang, content, root, uri)
	case "ui_object_binding":
		bindingText := string(content[parent.StartByte():parent.EndByte()])
		if idName := extractIdFromBinding(bindingText); idName != "" {
			return findIdDeclarations(idName, root, lang, content, uri)
		}
	}
	return nil
}

// findKeywordTarget resolves implicit QML identifiers (`parent`, `this`,
// `root`) to their enclosing ui_object_definition.
func findKeywordTarget(text string, node *gotreesitter.Node, lang *gotreesitter.Language, root *gotreesitter.Node) *gotreesitter.Node {
	switch text {
	case "parent", "this", "root":
		return findEnclosingObject(node, lang)
	}
	return nil
}

func findEnclosingObject(node *gotreesitter.Node, lang *gotreesitter.Language) *gotreesitter.Node {
	const maxDepth = 64
	current := node.Parent()
	for i := 0; current != nil && i < maxDepth; i++ {
		if current.Type(lang) == "ui_object_definition" {
			return current
		}
		current = current.Parent()
	}
	return nil
}

func extractIdFromBinding(bindingText string) string {
	idx := strings.Index(bindingText, ":")
	if idx < 0 {
		return ""
	}
	key := strings.TrimSpace(bindingText[:idx])
	if key != "id" {
		return ""
	}
	return strings.TrimSpace(bindingText[idx+1:])
}

func findIdDeclarations(idName string, root *gotreesitter.Node, lang *gotreesitter.Language, content []byte, uri lsp.DocumentURI) []lsp.Location {
	var locations []lsp.Location
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) != "ui_object_binding" {
			return true
		}
		text := string(content[n.StartByte():n.EndByte()])
		if extractIdFromBinding(text) == idName {
			locations = append(locations, nodeLocation(uri, content, n))
			return false
		}
		return true
	})
	return locations
}

func findComponentDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node, uri lsp.DocumentURI) []lsp.Location {
	name := string(content[node.StartByte():node.EndByte()])
	sym, ok := lookupSymbol(name)
	if !ok || sym.Module == "" {
		return nil
	}
	return findImportForModule(root, lang, content, sym.Module, uri)
}

func findImportForModule(root *gotreesitter.Node, lang *gotreesitter.Language, content []byte, module string, uri lsp.DocumentURI) []lsp.Location {
	base := module
	if idx := strings.Index(module, "."); idx > 0 {
		base = module[:idx]
	}
	var locations []lsp.Location
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) != "ui_import" {
			return true
		}
		if strings.Contains(string(content[n.StartByte():n.EndByte()]), base) {
			locations = append(locations, nodeLocation(uri, content, n))
			return false
		}
		return true
	})
	return locations
}

// walkTree does a pre-order walk; visit returns false to stop descending into
// this subtree.
func walkTree(node *gotreesitter.Node, visit func(*gotreesitter.Node) bool) {
	if node == nil {
		return
	}
	if !visit(node) {
		return
	}
	for i := 0; i < node.ChildCount(); i++ {
		walkTree(node.Child(i), visit)
	}
}
