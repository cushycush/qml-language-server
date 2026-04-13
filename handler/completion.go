package handler

import (
	"context"
	"fmt"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Completion(_ context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	doc, ok := h.documents[params.TextDocument.URI]
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
	case ContextProperty:
		items = append(items, qmlPropertyCompletions()...)
	case ContextId:
		items = append(items, qmlKeywords()...)
	case ContextAfterColon:
		items = append(items, getValueCompletions()...)
	default:
		items = append(items, getCompletionTypes()...)
		items = append(items, qmlKeywords()...)
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

	return []lsp.CompletionItem{
		{Label: "true", Kind: &boolKind, Detail: "Boolean true value"},
		{Label: "false", Kind: &boolKind, Detail: "Boolean false value"},
		{Label: "parent", Detail: "Reference to parent item"},
		{Label: "this", Detail: "Reference to current item"},
		{Label: "Qt.rect()", Kind: &snippetKind, InsertText: "Qt.rect(${1:x}, ${2:y}, ${3:width}, ${4:height})", Detail: "Create a rect value"},
		{Label: "Qt.size()", Kind: &snippetKind, InsertText: "Qt.size(${1:width}, ${2:height})", Detail: "Create a size value"},
		{Label: "Qt.point()", Kind: &snippetKind, InsertText: "Qt.point(${1:x}, ${2:y})", Detail: "Create a point value"},
		{Label: "\"red\"", Kind: &colorKind, Detail: "Red color"},
		{Label: "\"green\"", Kind: &colorKind, Detail: "Green color"},
		{Label: "\"blue\"", Kind: &colorKind, Detail: "Blue color"},
		{Label: "\"white\"", Kind: &colorKind, Detail: "White color"},
		{Label: "\"black\"", Kind: &colorKind, Detail: "Black color"},
	}
}

func getAnchorCompletions() []lsp.CompletionItem {
	propertyKind := lsp.CompletionItemKindProperty

	return []lsp.CompletionItem{
		{Label: "fill", Kind: &propertyKind, InsertText: "fill: parent", Detail: "Fill parent area"},
		{Label: "centerIn", Kind: &propertyKind, InsertText: "centerIn: parent", Detail: "Center in parent"},
		{Label: "top", Kind: &propertyKind, InsertText: "top: parent.top", Detail: "Align to parent top"},
		{Label: "bottom", Kind: &propertyKind, InsertText: "bottom: parent.bottom", Detail: "Align to parent bottom"},
		{Label: "left", Kind: &propertyKind, InsertText: "left: parent.left", Detail: "Align to parent left"},
		{Label: "right", Kind: &propertyKind, InsertText: "right: parent.right", Detail: "Align to parent right"},
		{Label: "horizontalCenter", Kind: &propertyKind, InsertText: "horizontalCenter: parent.horizontalCenter", Detail: "Center horizontally"},
		{Label: "verticalCenter", Kind: &propertyKind, InsertText: "verticalCenter: parent.verticalCenter", Detail: "Center vertically"},
		{Label: "margins", Kind: &propertyKind, InsertText: "margins: ${1:10}", Detail: "Add margins around anchor"},
		{Label: "topMargin", Kind: &propertyKind, InsertText: "topMargin: ${1:10}", Detail: "Top margin"},
		{Label: "bottomMargin", Kind: &propertyKind, InsertText: "bottomMargin: ${1:10}", Detail: "Bottom margin"},
		{Label: "leftMargin", Kind: &propertyKind, InsertText: "leftMargin: ${1:10}", Detail: "Left margin"},
		{Label: "rightMargin", Kind: &propertyKind, InsertText: "rightMargin: ${1:10}", Detail: "Right margin"},
	}
}

func (h *Handler) ResolveCompletionItem(_ context.Context, item *lsp.CompletionItem) (*lsp.CompletionItem, error) {
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
	for i := pos - 1; i >= 0; i-- {
		if text[i] == ' ' || text[i] == '\t' {
			continue
		}
		if i >= 5 && text[i-5:i] == "import" {
			return ContextImport
		}
		if text[i] == '.' {
			return ContextProperty
		}
		if text[i] == ':' {
			return ContextAfterColon
		}
		break
	}

	trimmed := trimLeadingWhitespace(text[:pos])
	if trimmed == "" || trimmed == "import " {
		return ContextImport
	}

	if isUpperCase(trimmed) {
		return ContextTypeName
	}

	return ContextDefault
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
	return []lsp.CompletionItem{
		{Label: "QtQuick", Detail: "Qt Quick module"},
		{Label: "QtQuick.Controls", Detail: "Qt Quick Controls module"},
		{Label: "QtQuick.Layouts", Detail: "Qt Quick Layouts module"},
		{Label: "QtQuick.Window", Detail: "Qt Quick Window module"},
		{Label: "QtQuick.Dialogs", Detail: "Qt Quick Dialogs module"},
		{Label: "QtQuick.Shapes", Detail: "Qt Quick Shapes module"},
		{Label: "QtQuick.Templates", Detail: "Qt Quick Templates module"},
		{Label: "QtQml", Detail: "Qt QML module"},
		{Label: "QtQml.Models", Detail: "Qt QML Models module"},
	}
}

func qmlCompletionTypes() []lsp.CompletionItem {
	classKind := lsp.CompletionItemKindClass
	types := []lsp.CompletionItem{
		{Label: "Item", Kind: &classKind, Detail: "QtQuick - Basic visual QML type"},
		{Label: "Rectangle", Kind: &classKind, Detail: "QtQuick - A rectangle type"},
		{Label: "Text", Kind: &classKind, Detail: "QtQuick - Text display type"},
		{Label: "Image", Kind: &classKind, Detail: "QtQuick - Image display type"},
		{Label: "MouseArea", Kind: &classKind, Detail: "QtQuick - Mouse event handling"},
		{Label: "Column", Kind: &classKind, Detail: "QtQuick - Vertical layout"},
		{Label: "Row", Kind: &classKind, Detail: "QtQuick - Horizontal layout"},
		{Label: "Grid", Kind: &classKind, Detail: "QtQuick - Grid layout"},
		{Label: "ListView", Kind: &classKind, Detail: "QtQuick - List view"},
		{Label: "ListModel", Kind: &classKind, Detail: "QtQuick - List data model"},
		{Label: "Component", Kind: &classKind, Detail: "QtQml - Component definition"},
		{Label: "QtObject", Kind: &classKind, Detail: "QtQml - Basic non-visual object"},
		{Label: "Timer", Kind: &classKind, Detail: "QtQuick - Timer for intervals"},
		{Label: "State", Kind: &classKind, Detail: "QtQuick - State definition"},
		{Label: "PropertyChanges", Kind: &classKind, Detail: "QtQuick - Property changes for states"},
		{Label: "Transition", Kind: &classKind, Detail: "QtQuick - Animated transitions"},
		{Label: "Behavior", Kind: &classKind, Detail: "QtQuick - Default property animation"},
		{Label: "Repeater", Kind: &classKind, Detail: "QtQuick - Item repetition"},
		{Label: "Loader", Kind: &classKind, Detail: "QtQuick - Dynamic loading"},
		{Label: "FocusScope", Kind: &classKind, Detail: "QtQuick - Keyboard focus scope"},
		{Label: "Keys", Kind: &classKind, Detail: "QtQuick - Key handling"},
		{Label: "ColumnLayout", Kind: &classKind, Detail: "QtQuick.Layouts - Vertical layout"},
		{Label: "RowLayout", Kind: &classKind, Detail: "QtQuick.Layouts - Horizontal layout"},
		{Label: "GridLayout", Kind: &classKind, Detail: "QtQuick.Layouts - Grid layout"},
	}

	return types
}

func qmlPropertyCompletions() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "id", InsertText: "id: ", Detail: "Unique identifier"},
		{Label: "property", InsertText: "property var ", Detail: "Property declaration"},
		{Label: "readonly property", InsertText: "readonly property var ", Detail: "Read-only property"},
		{Label: "required property", InsertText: "required property var ", Detail: "Required property"},
		{Label: "signal", InsertText: "signal ", Detail: "Signal declaration"},
		{Label: "function", InsertText: "function() {\n\t\n}", Detail: "Function declaration"},
		{Label: "x", InsertText: "x: ", Detail: "Item x position"},
		{Label: "y", InsertText: "y: ", Detail: "Item y position"},
		{Label: "width", InsertText: "width: ", Detail: "Item width"},
		{Label: "height", InsertText: "height: ", Detail: "Item height"},
		{Label: "color", InsertText: "color: ", Detail: "Color property"},
		{Label: "text", InsertText: "text: ", Detail: "Text content"},
		{Label: "anchors", InsertText: "anchors.", Detail: "Anchoring system"},
		{Label: "states", InsertText: "states: [", Detail: "List of states"},
		{Label: "transitions", InsertText: "transitions: [", Detail: "List of transitions"},
	}
}

func qmlKeywords() []lsp.CompletionItem {
	keywordKind := lsp.CompletionItemKindKeyword
	variableKind := lsp.CompletionItemKindVariable
	classKind := lsp.CompletionItemKindClass
	return []lsp.CompletionItem{
		{Label: "import", Kind: &keywordKind, InsertText: "import ", Detail: "Import statement"},
		{Label: "id", Kind: &variableKind, InsertText: "id: ", Detail: "Identifier assignment"},
		{Label: "property", Kind: &keywordKind, InsertText: "property ", Detail: "Property declaration"},
		{Label: "signal", Kind: &keywordKind, InsertText: "signal ", Detail: "Signal declaration"},
		{Label: "function", Kind: &keywordKind, InsertText: "function ", Detail: "Function declaration"},
		{Label: "on", Kind: &keywordKind, InsertText: "on ", Detail: "Signal handler prefix"},
		{Label: "Item", Kind: &classKind, InsertText: "Item", Detail: "QtQuick - Base visual type"},
		{Label: "ListElement", Kind: &classKind, InsertText: "ListElement", Detail: "QtQml.Models - List element"},
	}
}
