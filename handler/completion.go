package handler

import (
	"context"
	"fmt"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Completion(_ context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	doc, ok := h.getDocument(params.TextDocument.URI)
	if !ok {
		return &lsp.CompletionList{Items: []lsp.CompletionItem{}}, nil
	}

	pos := params.TextDocumentPositionParams.Position
	line := int(pos.Line)
	char := int(pos.Character)

	lines := getLines(doc)
	if line >= len(lines) {
		return &lsp.CompletionList{Items: []lsp.CompletionItem{}}, nil
	}

	lineText := lines[line]

	context := detectCompletionContext(lineText, char)
	var items []lsp.CompletionItem

	if h.parser != nil {
		node := h.parser.GetNodeAt(params.TextDocument.URI, pos, []byte(doc))
		if node != nil {
			items = append(items, getContextCompletions(node, h.parser.Language(), []byte(doc))...)
		}
	}

	switch context {
	case ContextImport:
		items = append(items, qmlImports()...)
	case ContextTypeName:
		items = append(items, getCompletionTypes()...)
		items = append(items, h.workspaceCompletions()...)
	case ContextProperty:
		items = append(items, qmlPropertyCompletions()...)
	case ContextId:
		items = append(items, qmlKeywords()...)
	case ContextAfterColon:
		items = append(items, getValueCompletions()...)
		items = append(items, completionItemsByCategory("js")...)
	default:
		items = append(items, getCompletionTypes()...)
		items = append(items, h.workspaceCompletions()...)
		items = append(items, qmlKeywords()...)
		items = append(items, completionItemsByCategory("js")...)
	}

	return &lsp.CompletionList{
		Items:        items,
		IsIncomplete: false,
	}, nil
}

func getContextCompletions(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.CompletionItem {
	var items []lsp.CompletionItem
	nodeType := node.Type(lang)

	switch nodeType {
	case "ui_import":
		items = append(items, qmlImports()...)

	case "ui_object_definition":
		items = append(items, getCompletionTypes()...)

	case "ui_binding":
		items = append(items, qmlPropertyCompletions()...)

	case "nested_identifier":
		parent := node.Parent()
		if parent != nil && parent.Type(lang) == "ui_binding" {
			parentText := string(content[parent.StartByte():parent.EndByte()])
			if len(parentText) > 0 && parentText[len(parentText)-1] == '.' {
				items = append(items, getAnchorCompletions()...)
			}
		}

	case "identifier":
		parent := node.Parent()
		if parent != nil {
			parentType := parent.Type(lang)
			if parentType == "ui_object_definition" || parentType == "ui_required" || parentType == "ui_property" {
				items = append(items, getCompletionTypes()...)
			}
		}

	case "property_identifier":
		items = append(items, qmlPropertyCompletions()...)
	}

	return items
}

func getValueCompletions() []lsp.CompletionItem {
	snippetKind := lsp.CompletionItemKindSnippet
	boolKind := lsp.CompletionItemKindValue
	colorKind := lsp.CompletionItemKindColor
	snippetFmt := lsp.InsertTextFormatSnippet

	return []lsp.CompletionItem{
		{Label: "true", Kind: &boolKind, Detail: "Boolean true value"},
		{Label: "false", Kind: &boolKind, Detail: "Boolean false value"},
		{Label: "parent", Detail: "Reference to the parent item"},
		{Label: "this", Detail: "Reference to the current item"},
		{Label: "Qt.rect()", Kind: &snippetKind, InsertText: "Qt.rect(${1:x}, ${2:y}, ${3:width}, ${4:height})", InsertTextFormat: &snippetFmt, Detail: "Create a rect value"},
		{Label: "Qt.size()", Kind: &snippetKind, InsertText: "Qt.size(${1:width}, ${2:height})", InsertTextFormat: &snippetFmt, Detail: "Create a size value"},
		{Label: "Qt.point()", Kind: &snippetKind, InsertText: "Qt.point(${1:x}, ${2:y})", InsertTextFormat: &snippetFmt, Detail: "Create a point value"},
		{Label: "\"red\"", Kind: &colorKind, Detail: "Red color"},
		{Label: "\"green\"", Kind: &colorKind, Detail: "Green color"},
		{Label: "\"blue\"", Kind: &colorKind, Detail: "Blue color"},
		{Label: "\"white\"", Kind: &colorKind, Detail: "White color"},
		{Label: "\"black\"", Kind: &colorKind, Detail: "Black color"},
	}
}

func getAnchorCompletions() []lsp.CompletionItem {
	return completionItemsByCategory("anchor")
}

func (h *Handler) ResolveCompletionItem(_ context.Context, item *lsp.CompletionItem) (*lsp.CompletionItem, error) {
	if item.Documentation != nil {
		return item, nil
	}
	if sym, ok := lookupSymbol(item.Label); ok {
		item.Documentation = &lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: sym.Render(),
		}
		if item.Detail == "" {
			item.Detail = sym.Detail
		}
		return item, nil
	}
	if info, ok := getTypeInfo(item.Label); ok {
		item.Documentation = &lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: fmt.Sprintf("## `%s`\n\n%s\n\n**Type:** %s\n\n**Module:** %s", item.Label, info.Description, info.Type, info.Module),
		}
	}
	return item, nil
}

type CompletionContext int

const (
	ContextDefault CompletionContext = iota
	ContextImport
	ContextTypeName
	ContextProperty
	ContextId
	ContextAfterColon
)

func detectCompletionContext(text string, pos int) CompletionContext {
	if pos > len(text) {
		pos = len(text)
	}

	// Look at the last non-whitespace byte before the cursor: a '.' means the
	// user is chaining into an anchor/nested member, a ':' means they're on
	// the value side of a binding.
	for i := pos - 1; i >= 0; i-- {
		c := text[i]
		if c == ' ' || c == '\t' {
			continue
		}
		switch c {
		case '.':
			return ContextProperty
		case ':':
			return ContextAfterColon
		}
		break
	}

	trimmed := trimLeadingWhitespace(text[:pos])
	if trimmed == "" {
		return ContextDefault
	}

	// If the first token on the line is "import", everything after is a
	// module name.
	if hasWordPrefix(trimmed, "import") {
		return ContextImport
	}

	if isUpperCase(trimmed) {
		return ContextTypeName
	}
	return ContextDefault
}

func hasWordPrefix(s, word string) bool {
	if len(s) < len(word) {
		return false
	}
	if s[:len(word)] != word {
		return false
	}
	if len(s) == len(word) {
		return true
	}
	next := s[len(word)]
	return next == ' ' || next == '\t'
}

func isUpperCase(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return false
		}
	}
	return len(s) > 0
}

func trimLeadingWhitespace(s string) string {
	i := 0
	for ; i < len(s) && (s[i] == ' ' || s[i] == '\t'); i++ {
	}
	return s[i:]
}

func qmlImports() []lsp.CompletionItem {
	return completionItemsByCategory("import")
}

func qmlPropertyCompletions() []lsp.CompletionItem {
	return completionItemsByCategory("property", "anchor")
}

func qmlKeywords() []lsp.CompletionItem {
	return completionItemsByCategory("keyword")
}
