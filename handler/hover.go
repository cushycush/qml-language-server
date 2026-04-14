package handler

import (
	"context"
	"fmt"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

// Hover resolves the node under the cursor and renders rich markdown from the
// symbol registry when we recognize the identifier.
func (h *Handler) Hover(_ context.Context, params *lsp.HoverParams) (*lsp.Hover, error) {
	uri := params.TextDocument.URI
	doc, ok := h.getDocument(uri)
	if !ok {
		return nil, nil
	}
	content := []byte(doc)

	if h.parser == nil {
		return simpleHover(doc, params.Position), nil
	}

	node := h.parser.GetNodeAt(uri, params.Position, content)
	if node == nil {
		return simpleHover(doc, params.Position), nil
	}
	lang := h.parser.Language()

	body := hoverBody(node, lang, content)
	if body == "" {
		return nil, nil
	}
	return &lsp.Hover{
		Range:    ptrRange(nodeRange(content, node)),
		Contents: lsp.MarkupContent{Kind: lsp.Markdown, Value: body},
	}, nil
}

func ptrRange(r lsp.Range) *lsp.Range { return &r }

// hoverBody produces markdown for the node under the cursor, or "" to tell
// the caller there's nothing useful to show.
func hoverBody(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	nodeText := string(content[node.StartByte():node.EndByte()])

	switch node.Type(lang) {
	case "identifier", "property_identifier", "nested_identifier":
		return identifierHover(nodeText, node, lang, content)
	case "ui_object_definition":
		return typeHoverOrGeneric(extractObjectType(node, lang, content))
	case "ui_import":
		return "**QML import**\n\n```qml\n" + nodeText + "\n```"
	case "string", "string_fragment":
		return "**String literal**\n\n```\n" + nodeText + "\n```"
	case "number":
		return "**Number literal** — `" + nodeText + "`"
	case "comment":
		return "**Comment**\n\n" + nodeText
	}
	// Unknown node — still try a registry lookup on its text.
	if sym, ok := lookupSymbol(nodeText); ok {
		return sym.Render()
	}
	return ""
}

func identifierHover(nodeText string, node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	if sym, ok := lookupSymbol(nodeText); ok {
		return sym.Render()
	}
	// Workspace component reference?
	if info := workspaceComponentHover(nodeText); info != "" {
		return info
	}
	// Fall back to parent-context text. e.g. an identifier inside an import
	// that's not in the registry still deserves a line.
	parent := node.Parent()
	if parent != nil {
		switch parent.Type(lang) {
		case "ui_object_definition":
			return "**Component:** `" + nodeText + "`"
		case "ui_binding":
			return "**Property:** `" + nodeText + "`"
		}
	}
	return "**Identifier:** `" + nodeText + "`"
}

func workspaceComponentHover(name string) string {
	// Avoid a hard dependency on the handler receiver by doing a cheap lookup
	// through the shared registry first. Workspace components only show up if
	// they've been explicitly registered.
	sym, ok := lookupSymbol(name)
	if !ok || sym.Category != "workspace" {
		return ""
	}
	return sym.Render()
}

func typeHoverOrGeneric(name string) string {
	if sym, ok := lookupSymbol(name); ok {
		return sym.Render()
	}
	if info, ok := getTypeInfo(name); ok {
		return fmt.Sprintf("**Type:** `%s`\n\n%s\n\n**Module:** %s", name, info.Description, info.Module)
	}
	return fmt.Sprintf("**Component:** `%s`\n\nA QML component definition.", name)
}

// simpleHover runs when the parser is unavailable (e.g. grammar didn't load).
// It does a best-effort lookup on the word under the cursor.
func simpleHover(doc string, pos lsp.Position) *lsp.Hover {
	lines := getLines(doc)
	if int(pos.Line) >= len(lines) {
		return nil
	}
	lineText := lines[int(pos.Line)]
	char := int(pos.Character)
	if char > len(lineText) {
		char = len(lineText)
	}
	word := extractWordAt(lineText, char)
	if word == "" {
		return nil
	}
	if sym, ok := lookupSymbol(word); ok {
		return &lsp.Hover{
			Contents: lsp.MarkupContent{Kind: lsp.Markdown, Value: sym.Render()},
		}
	}
	return nil
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
