package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Hover(_ context.Context, params *lsp.HoverParams) (*lsp.Hover, error) {
	uri := params.TextDocument.URI
	pos := params.Position

	doc, ok := h.documents[uri]
	if !ok {
		return nil, nil
	}

	if h.parser == nil {
		return simpleHover(doc, pos), nil
	}

	node := h.parser.GetNodeAt(uri, pos)
	if node == nil {
		return simpleHover(doc, pos), nil
	}

	lang := h.parser.Language()
	content := []byte(doc)
	nodeText := string(content[node.StartByte():node.EndByte()])
	nodeType := node.Type(lang)

	var hoverContent string

	switch nodeType {
	case "identifier":
		parent := node.Parent()
		if parent != nil {
			parentType := parent.Type(lang)
			switch parentType {
			case "ui_object_definition":
				hoverContent = getTypeHoverInfo(nodeText)
			case "nested_identifier":
				if info, ok := getPropertyInfo(nodeText); ok {
					hoverContent = fmt.Sprintf("**Property:** %s\n\n%s", nodeText, info.Description)
				} else {
					hoverContent = getTypeHoverInfo(nodeText)
				}
			case "ui_binding":
				if info, ok := getPropertyInfo(nodeText); ok {
					hoverContent = fmt.Sprintf("**Property:** %s\n\n%s\n\n**Type:** %s", nodeText, info.Description, info.Type)
				} else {
					hoverContent = getTypeHoverInfo(nodeText)
				}
			default:
				hoverContent = getTypeHoverInfo(nodeText)
			}
		} else {
			hoverContent = getTypeHoverInfo(nodeText)
		}

	case "string", "string_fragment":
		hoverContent = "**String literal**\n\nA text value enclosed in quotes."

	case "number":
		hoverContent = "**Number literal**\n\nA numeric value."

	case "ui_import":
		hoverContent = "**QML Import**\n\nImports a module to use its types.\n\n```qml\n" + nodeText + "\n```"

	case "ui_object_definition":
		typeName := extractObjectType(node, lang, content)
		if info, ok := getTypeInfo(typeName); ok {
			hoverContent = fmt.Sprintf("**Component:** %s\n\n%s\n\n**Module:** %s", typeName, info.Description, info.Module)
		} else {
			hoverContent = fmt.Sprintf("**Component:** %s\n\nA QML component definition.", typeName)
		}

	case "comment":
		hoverContent = "**Comment**\n\n" + nodeText

	default:
		if info, ok := getTypeInfo(nodeText); ok {
			hoverContent = fmt.Sprintf("**Type:** %s\n\n%s\n\n**Module:** %s", nodeText, info.Description, info.Module)
		} else if info, ok := getPropertyInfo(nodeText); ok {
			hoverContent = fmt.Sprintf("**Property:** %s\n\n%s", nodeText, info.Description)
		} else {
			hoverContent = fmt.Sprintf("**%s**\n\n`%s`", strings.Title(nodeType), nodeText)
		}
	}

	return &lsp.Hover{
		Range: &lsp.Range{
			Start: lsp.Position{Line: pos.Line, Character: int(node.StartByte())},
			End:   lsp.Position{Line: pos.Line, Character: int(node.EndByte())},
		},
		Contents: lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: hoverContent,
		},
	}, nil
}

func simpleHover(doc string, pos lsp.Position) *lsp.Hover {
	lines := getLines(doc)
	line := int(pos.Line)
	char := int(pos.Character)

	if line >= len(lines) {
		return nil
	}

	lineText := lines[line]
	if char >= len(lineText) {
		char = len(lineText) - 1
	}

	word := extractWordAt(lineText, char)
	if word == "" {
		return nil
	}

	if info, ok := getTypeInfo(word); ok {
		return &lsp.Hover{
			Range: &lsp.Range{
				Start: lsp.Position{Line: pos.Line, Character: char - len(word)},
				End:   lsp.Position{Line: pos.Line, Character: char},
			},
			Contents: lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: fmt.Sprintf("**Type:** %s\n\n%s\n\n**Module:** %s", word, info.Description, info.Module),
			},
		}
	}

	return nil
}

func getTypeHoverInfo(name string) string {
	if info, ok := getTypeInfo(name); ok {
		return fmt.Sprintf("**Type:** %s\n\n%s\n\n**Module:** %s", name, info.Description, info.Module)
	}
	if info, ok := getPropertyInfo(name); ok {
		return fmt.Sprintf("**Property:** %s\n\n%s", name, info.Description)
	}
	return fmt.Sprintf("**Identifier:** `%s`", name)
}

func extractObjectType(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Type(lang) == "identifier" {
			return string(content[child.StartByte():child.EndByte()])
		}
	}
	return "Unknown"
}
