package handler

import "github.com/owenrumney/go-lsp/lsp"

type QMLTypeInfo struct {
	Description string
	Type        string
	Module      string
}

var qmlTypes = map[string]QMLTypeInfo{
	"Item":            {Description: "A basic visual QML type. All visual items in Qt Quick inherit from Item.", Type: "Object", Module: "QtQuick"},
	"Rectangle":       {Description: "A rectangle type that can specify geometry and color.", Type: "Object", Module: "QtQuick"},
	"Text":            {Description: "A text display type that can render plain or rich text.", Type: "Object", Module: "QtQuick"},
	"Image":           {Description: "A type that displays an image.", Type: "Object", Module: "QtQuick"},
	"MouseArea":       {Description: "A rectangular area that responds to mouse events.", Type: "Object", Module: "QtQuick"},
	"Column":          {Description: "Positions its children in a column.", Type: "Object", Module: "QtQuick"},
	"Row":             {Description: "Positions its children in a row.", Type: "Object", Module: "QtQuick"},
	"Grid":            {Description: "Positions its children in a grid.", Type: "Object", Module: "QtQuick"},
	"ListView":        {Description: "A list view type for displaying a list of items.", Type: "Object", Module: "QtQuick"},
	"ListModel":       {Description: "Defines a data model for ListView.", Type: "Object", Module: "QtQuick"},
	"Component":       {Description: "Encapsulates a QML component definition.", Type: "Object", Module: "QtQml"},
	"QtObject":        {Description: "A basic non-visual object type.", Type: "Object", Module: "QtQml"},
	"Timer":           {Description: "A timer type for triggering actions at intervals.", Type: "Object", Module: "QtQuick"},
	"States":          {Description: "A list of State types.", Type: "Object", Module: "QtQuick"},
	"State":           {Description: "Defines a configuration of properties for an object.", Type: "Object", Module: "QtQuick"},
	"PropertyChanges": {Description: "Describes property changes for a state.", Type: "Object", Module: "QtQuick"},
	"Transition":      {Description: "Defines animated transitions between states.", Type: "Object", Module: "QtQuick"},
	"Animation":       {Description: "The base type for animations.", Type: "Object", Module: "QtQuick"},
	"Behavior":        {Description: "Defines a default animation for property changes.", Type: "Object", Module: "QtQuick"},
	"Repeater":        {Description: "Creates multiple copies of items from a model.", Type: "Object", Module: "QtQuick"},
	"Loader":          {Description: "Allows dynamic loading of a subtree from a URL or Component.", Type: "Object", Module: "QtQuick"},
	"FocusScope":      {Description: "A scope that accepts keyboard focus.", Type: "Object", Module: "QtQuick"},
	"Keys":            {Description: "Provides key handling for items.", Type: "Object", Module: "QtQuick"},
	"ColumnLayout":    {Description: "Arranges items vertically.", Type: "Object", Module: "QtQuick.Layouts"},
	"RowLayout":       {Description: "Arranges items horizontally.", Type: "Object", Module: "QtQuick.Layouts"},
	"GridLayout":      {Description: "Arranges items in a grid.", Type: "Object", Module: "QtQuick.Layouts"},
	"StackView":       {Description: "Provides a stack-based navigation model.", Type: "Object", Module: "QtQuick.Controls"},
	"SwipeView":       {Description: "Enables swiping between views.", Type: "Object", Module: "QtQuick.Controls"},
	"TabBar":          {Description: "Implements a tab bar.", Type: "Object", Module: "QtQuick.Controls"},
	"Button":          {Description: "A push button control.", Type: "Object", Module: "QtQuick.Controls"},
	"TextField":       {Description: "A single-line text input field.", Type: "Object", Module: "QtQuick.Controls"},
	"TextArea":        {Description: "A multi-line text editor.", Type: "Object", Module: "QtQuick.Controls"},
	"ListElement":     {Description: "Defines a data element in a ListModel.", Type: "Object", Module: "QtQml.Models"},
}

func getTypeInfo(name string) (QMLTypeInfo, bool) {
	info, ok := qmlTypes[name]
	return info, ok
}

type PropertyInfo struct {
	Description string
	Type        string
}

var qmlPropertiesMap = map[string]PropertyInfo{
	"id":                {Description: "Unique identifier for referencing this component", Type: "string"},
	"x":                 {Description: "X position of the item", Type: "real"},
	"y":                 {Description: "Y position of the item", Type: "real"},
	"width":             {Description: "Width of the item", Type: "real"},
	"height":            {Description: "Height of the item", Type: "real"},
	"z":                 {Description: "Z-ordering depth", Type: "real"},
	"visible":           {Description: "Whether the item is visible", Type: "bool"},
	"enabled":           {Description: "Whether the item is enabled", Type: "bool"},
	"opacity":           {Description: "Opacity value from 0.0 to 1.0", Type: "real"},
	"rotation":          {Description: "Rotation angle in degrees", Type: "real"},
	"scale":             {Description: "Scale factor", Type: "real"},
	"anchors":           {Description: "Anchor properties for layout positioning", Type: "AnchorAnimation"},
	"color":             {Description: "Color value (e.g., \"red\", \"#RRGGBB\")", Type: "color"},
	"text":              {Description: "Text content", Type: "string"},
	"font":              {Description: "Font properties", Type: "Font"},
	"radius":            {Description: "Corner radius", Type: "real"},
	"source":            {Description: "Source URL or file path", Type: "url"},
	"model":             {Description: "Data model for list views", Type: "any"},
	"delegate":          {Description: "Item delegate for list views", Type: "Component"},
	"currentIndex":      {Description: "Current selected index", Type: "int"},
	"count":             {Description: "Number of items", Type: "int"},
	"onClicked":         {Description: "Signal handler for click events", Type: "signal"},
	"onPressed":         {Description: "Signal handler for press events", Type: "signal"},
	"onReleased":        {Description: "Signal handler for release events", Type: "signal"},
	"onEntered":         {Description: "Signal handler for mouse enter", Type: "signal"},
	"onExited":          {Description: "Signal handler for mouse exit", Type: "signal"},
	"Layout.fillWidth":  {Description: "Fill available horizontal space", Type: "bool"},
	"Layout.fillHeight": {Description: "Fill available vertical space", Type: "bool"},
	"Layout.alignment":  {Description: "Alignment within cell", Type: "alignment"},
	"Layout.columnSpan": {Description: "Number of columns to span", Type: "int"},
	"Layout.rowSpan":    {Description: "Number of rows to span", Type: "int"},
}

func getPropertyInfo(name string) (PropertyInfo, bool) {
	info, ok := qmlPropertiesMap[name]
	return info, ok
}

func getLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(text); i++ {
		if i == len(text) || text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	return lines
}

func extractWordAt(text string, pos int) string {
	if pos < 0 || pos > len(text) {
		return ""
	}

	start := pos
	end := pos

	for start > 0 && isIdentChar(text[start-1]) {
		start--
	}
	for end < len(text) && isIdentChar(text[end]) {
		end++
	}

	return text[start:end]
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func getCompletionTypes() []lsp.CompletionItem {
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
