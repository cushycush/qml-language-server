package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Definition(_ context.Context, params *lsp.DefinitionParams) ([]lsp.Location, error) {
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
	pos := params.Position

	byteOffset := positionToByte(content, pos)
	node := findSmallestNodeAt(root, byteOffset, lang)

	if node == nil {
		return nil, nil
	}

	return findDefinition(node, lang, content, root)
}

func positionToByte(content []byte, pos lsp.Position) uint32 {
	line := int(pos.Line)
	char := int(pos.Character)

	offset := uint32(0)
	for i := 0; i < line; i++ {
		idx := indexOf(content[offset:], '\n')
		if idx < 0 {
			return offset
		}
		offset += uint32(idx) + 1
	}

	currentLineStart := offset
	currentLineEnd := offset
	for currentLineEnd < uint32(len(content)) && content[currentLineEnd] != '\n' {
		currentLineEnd++
	}

	if char > int(currentLineEnd-currentLineStart) {
		char = int(currentLineEnd - currentLineStart)
	}

	return offset + uint32(char)
}

func indexOf(data []byte, ch byte) int {
	for i, b := range data {
		if b == ch {
			return i
		}
	}
	return -1
}

func findSmallestNodeAt(node *gotreesitter.Node, offset uint32, lang *gotreesitter.Language) *gotreesitter.Node {
	if node == nil {
		return nil
	}

	if offset < node.StartByte() || offset >= node.EndByte() {
		return nil
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && offset >= child.StartByte() && offset < child.EndByte() {
			smallest := findSmallestNodeAt(child, offset, lang)
			if smallest != nil {
				return smallest
			}
		}
	}

	return node
}

func findDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node) ([]lsp.Location, error) {
	nodeType := node.Type(lang)
	var locations []lsp.Location

	switch nodeType {
	case "identifier":
		parent := node.Parent()
		if parent != nil {
			parentType := parent.Type(lang)
			switch parentType {
			case "nested_identifier":
				locations = findNestedIdentifierTarget(node, lang, content, root)
			case "ui_binding":
				target := findBindingTarget(node, lang, content, root)
				if target != nil {
					locations = append(locations, *target)
				}
			case "ui_object_definition":
				locations = findComponentDefinition(node, lang, content, root)
			}
		}

	case "property_identifier":
		locations = findPropertyDefinition(node, lang, content)

	case "ui_object_definition":
		locations = findComponentDefinition(node, lang, content, root)
	}

	return locations, nil
}

func findNestedIdentifierTarget(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node) []lsp.Location {
	var locations []lsp.Location

	identifierText := string(content[node.StartByte():node.EndByte()])

	if identifierText == "parent" || identifierText == "this" || identifierText == "root" {
		if parent := findParentObject(node, lang, root); parent != nil {
			locations = append(locations, lsp.Location{
				URI: "",
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(parent.StartByte())},
					End:   lsp.Position{Line: 0, Character: int(parent.EndByte())},
				},
			})
		}
	}

	return locations
}

func findParentObject(node *gotreesitter.Node, lang *gotreesitter.Language, root *gotreesitter.Node) *gotreesitter.Node {
	current := node
	depth := 0

	for current != nil && depth < 20 {
		if current.Type(lang) == "ui_object_definition" && current != node {
			return current
		}
		current = current.Parent()
		depth++
	}

	return nil
}

func findBindingTarget(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node) *lsp.Location {
	identifierText := string(content[node.StartByte():node.EndByte()])

	if identifierText == "parent" {
		if parent := findParentObject(node, lang, root); parent != nil {
			return &lsp.Location{
				URI: "",
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(parent.StartByte())},
					End:   lsp.Position{Line: 0, Character: int(parent.EndByte())},
				},
			}
		}
	}

	return nil
}

func findComponentDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node) []lsp.Location {
	var locations []lsp.Location
	return locations
}

func findPropertyDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.Location {
	var locations []lsp.Location

	propertyName := string(content[node.StartByte():node.EndByte()])

	if _, ok := getPropertyInfo(propertyName); ok {
		return locations
	}

	return locations
}
