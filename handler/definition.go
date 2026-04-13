package handler

import (
	"context"
	"strings"

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

	locations, _ := findDefinition(node, lang, content, root)

	if len(locations) == 0 {
		locations = findIdDefinition(node, lang, content)
	}

	return locations, nil
}

func findIdDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.Location {
	var locations []lsp.Location

	if node.Type(lang) == "identifier" {
		parent := node.Parent()
		if parent != nil && parent.Type(lang) == "ui_object_binding" {
			targetId := string(content[parent.StartByte():parent.EndByte()])
			locations = findIdDeclarations(targetId, lang, content)
		}
	}

	return locations
}

func findIdDeclarations(targetId string, lang *gotreesitter.Language, content []byte) []lsp.Location {
	var locations []lsp.Location
	return locations
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

func byteOffsetToPosition(content []byte, offset uint32) lsp.Position {
	line := 0
	char := 0
	for i := 0; i < int(offset) && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}
	return lsp.Position{Line: line, Character: char}
}

func nodeToLocation(node *gotreesitter.Node, uri lsp.DocumentURI) lsp.Location {
	content := []byte{}
	return lsp.Location{
		URI: uri,
		Range: lsp.Range{
			Start: byteOffsetToPosition(content, node.StartByte()),
			End:   byteOffsetToPosition(content, node.EndByte()),
		},
	}
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
			case "expression_statement":
				identifierText := string(content[node.StartByte():node.EndByte()])
				if identifierText == "parent" {
					target := findParentObject(node, lang, root)
					if target != nil {
						locations = append(locations, lsp.Location{
							URI: "",
							Range: lsp.Range{
								Start: byteOffsetToPosition(content, target.StartByte()),
								End:   byteOffsetToPosition(content, target.EndByte()),
							},
						})
					}
				}
			case "ui_object_definition":
				locations = findComponentDefinition(node, lang, content, root)
			case "ui_object_binding":
				bindingText := string(content[parent.StartByte():parent.EndByte()])
				idName := extractIdFromBinding(bindingText)
				if idName != "" {
					locations = findIdDeclaration(idName, root, lang, content)
				}
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
	current := node.Parent()
	depth := 0

	for current != nil && depth < 20 {
		if current.Type(lang) == "ui_object_definition" {
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
					Start: byteOffsetToPosition(content, parent.StartByte()),
					End:   byteOffsetToPosition(content, parent.EndByte()),
				},
			}
		}
	}

	return nil
}

func extractIdFromBinding(bindingText string) string {
	parts := strings.Split(bindingText, ":")
	if len(parts) >= 2 {
		idPart := strings.TrimSpace(parts[0])
		if strings.HasPrefix(idPart, "id:") {
			return strings.TrimSpace(strings.TrimPrefix(idPart, "id:"))
		}
	}
	return ""
}

func findIdDeclaration(idName string, root *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.Location {
	var locations []lsp.Location
	findIdDeclarationsRecursive(root, lang, content, idName, &locations)
	return locations
}

func findIdDeclarationsRecursive(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, targetId string, locations *[]lsp.Location) {
	if node == nil {
		return
	}

	if node.Type(lang) == "ui_object_definition" {
		for i := 0; i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child != nil && child.Type(lang) == "ui_object_binding" {
				bindingText := string(content[child.StartByte():child.EndByte()])
				if strings.HasPrefix(strings.TrimSpace(bindingText), "id:") {
					idPart := strings.TrimSpace(strings.TrimPrefix(bindingText, "id:"))
					if idPart == targetId {
						*locations = append(*locations, lsp.Location{
							URI: "",
							Range: lsp.Range{
								Start: byteOffsetToPosition(content, child.StartByte()),
								End:   byteOffsetToPosition(content, child.EndByte()),
							},
						})
						return
					}
				}
			}
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			findIdDeclarationsRecursive(child, lang, content, targetId, locations)
		}
	}
}

func findComponentDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, root *gotreesitter.Node) []lsp.Location {
	var locations []lsp.Location

	nodeTypeName := string(content[node.StartByte():node.EndByte()])
	if info, ok := getTypeInfo(nodeTypeName); ok && info.Module != "" {
		locations = findImportForModule(root, lang, content, info.Module)
	}

	return locations
}

func findImportForModule(root *gotreesitter.Node, lang *gotreesitter.Language, content []byte, module string) []lsp.Location {
	var locations []lsp.Location

	imports := findUiImports(root, lang)
	for _, imp := range imports {
		importText := string(content[imp.StartByte():imp.EndByte()])
		moduleName := module
		if idx := strings.Index(module, "."); idx > 0 {
			moduleName = module[:idx]
		}
		if strings.Contains(importText, moduleName) {
			locations = append(locations, lsp.Location{
				URI: "",
				Range: lsp.Range{
					Start: byteOffsetToPosition(content, imp.StartByte()),
					End:   byteOffsetToPosition(content, imp.EndByte()),
				},
			})
			return locations
		}
	}

	return locations
}

func findUiImports(node *gotreesitter.Node, lang *gotreesitter.Language) []*gotreesitter.Node {
	var imports []*gotreesitter.Node

	if node.Type(lang) == "ui_import" {
		imports = append(imports, node)
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			imports = append(imports, findUiImports(child, lang)...)
		}
	}

	return imports
}

func findPropertyDefinition(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.Location {
	var locations []lsp.Location

	propertyName := string(content[node.StartByte():node.EndByte()])

	if _, ok := getPropertyInfo(propertyName); ok {
		return locations
	}

	return locations
}
