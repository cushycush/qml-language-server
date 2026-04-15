package handler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/owenrumney/go-lsp/lsp"
)

// QMLSymbol is the unified knowledge entry used for both hover and completion.
// One source of truth, so completion items ship with rich Documentation instead
// of relying on ResolveCompletionItem to find docs that may never be asked for.
type QMLSymbol struct {
	Label         string
	Kind          lsp.CompletionItemKind
	Detail        string // short one-liner shown next to the label
	Signature     string // optional; rendered as a fenced code block above the description
	Description   string // prose shown below the signature
	Module        string // e.g. "QtQuick", "builtin", empty for keywords
	Category      string // "type", "property", "keyword", "import", "anchor", "js", "workspace"
	InsertText    string // optional; implies snippet format when set and containing $
	InsertSnippet bool   // forces Snippet insert text format
}

// Render produces the markdown body used for both hover and CompletionItem.Documentation.
func (s QMLSymbol) Render() string {
	var b strings.Builder

	header := s.Label
	switch s.Category {
	case "type":
		fmt.Fprintf(&b, "**%s** — type", header)
	case "property":
		fmt.Fprintf(&b, "**%s** — property", header)
	case "keyword":
		fmt.Fprintf(&b, "**%s** — keyword", header)
	case "import":
		fmt.Fprintf(&b, "**%s** — module", header)
	case "anchor":
		fmt.Fprintf(&b, "**%s** — anchor", header)
	case "js":
		fmt.Fprintf(&b, "**%s** — JavaScript", header)
	case "workspace":
		fmt.Fprintf(&b, "**%s** — workspace component", header)
	case "quickshell-snippet":
		fmt.Fprintf(&b, "**%s** — snippet", header)
	default:
		fmt.Fprintf(&b, "**%s**", header)
	}
	if s.Module != "" {
		fmt.Fprintf(&b, "  \n_%s_", s.Module)
	}
	b.WriteString("\n\n")

	if s.Signature != "" {
		b.WriteString("```qml\n")
		b.WriteString(s.Signature)
		b.WriteString("\n```\n\n")
	}

	if s.Description != "" {
		b.WriteString(s.Description)
	}

	return b.String()
}

// CompletionItem builds a fully populated LSP completion item including docs.
func (s QMLSymbol) CompletionItem() lsp.CompletionItem {
	kind := s.Kind
	item := lsp.CompletionItem{
		Label:  s.Label,
		Kind:   &kind,
		Detail: s.Detail,
		Documentation: &lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: s.Render(),
		},
	}
	if s.InsertText != "" {
		item.InsertText = s.InsertText
	}
	if s.InsertSnippet || strings.Contains(s.InsertText, "${") {
		fmtSnippet := lsp.InsertTextFormatSnippet
		item.InsertTextFormat = &fmtSnippet
	}
	return item
}

// symbolRegistry is looked up by Label. ResolveCompletionItem uses this as a
// fallback for clients that send us back items missing Documentation. The
// workspace scanner writes to it concurrently, so access is synchronized.
var (
	registryMu     sync.RWMutex
	symbolRegistry = map[string]QMLSymbol{}
)

func registerSymbols(entries ...QMLSymbol) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for _, e := range entries {
		symbolRegistry[e.Label] = e
	}
}

func lookupSymbol(label string) (QMLSymbol, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	s, ok := symbolRegistry[label]
	return s, ok
}

func symbolsByCategory(cats ...string) []QMLSymbol {
	want := make(map[string]struct{}, len(cats))
	for _, c := range cats {
		want[c] = struct{}{}
	}
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]QMLSymbol, 0, len(symbolRegistry))
	for _, s := range symbolRegistry {
		if _, ok := want[s.Category]; ok {
			out = append(out, s)
		}
	}
	return out
}

func completionItemsByCategory(cats ...string) []lsp.CompletionItem {
	syms := symbolsByCategory(cats...)
	items := make([]lsp.CompletionItem, 0, len(syms))
	for _, s := range syms {
		items = append(items, s.CompletionItem())
	}
	return items
}

func init() {
	registerBuiltinTypes()
	registerBuiltinProperties()
	registerAnchors()
	registerKeywords()
	registerImports()
	registerJSBuiltins()
	registerQuickshellBuiltins()
}

func registerBuiltinTypes() {
	t := lsp.CompletionItemKindClass
	types := []struct {
		name, module, detail, desc string
	}{
		{"Item", "QtQuick", "Basic visual QML type",
			"The base type of all visual items in Qt Quick. `Item` itself draws nothing, but carries all positioning, focus, key handling, transform and anchor behavior that its descendants inherit.\n\n**Common properties:** `x`, `y`, `width`, `height`, `anchors`, `visible`, `opacity`, `z`, `rotation`, `scale`, `transform`, `focus`, `children`."},
		{"Rectangle", "QtQuick", "Filled rectangle with optional border",
			"Paints a filled rectangle. Supports solid and gradient fills, border strokes, and rounded corners via `radius`.\n\n**Common properties:** `color`, `border.color`, `border.width`, `radius`, `gradient`."},
		{"Text", "QtQuick", "Text display",
			"Renders plain or rich text. Supports word wrap, eliding, and HTML via `textFormat: Text.RichText`.\n\n**Common properties:** `text`, `font`, `color`, `horizontalAlignment`, `verticalAlignment`, `wrapMode`, `elide`, `textFormat`."},
		{"Image", "QtQuick", "Image display",
			"Displays an image from a URL or resource path.\n\n**Common properties:** `source`, `fillMode`, `sourceSize`, `asynchronous`, `cache`, `mirror`, `status`."},
		{"MouseArea", "QtQuick", "Mouse/touch event handler",
			"An invisible rectangular region that emits signals for mouse and touch events. Typical usage: place inside or overlapping a visible item.\n\n**Common signals:** `onClicked`, `onPressed`, `onReleased`, `onDoubleClicked`, `onPositionChanged`, `onEntered`, `onExited`.\n\n```qml\nMouseArea {\n    anchors.fill: parent\n    onClicked: console.log(\"clicked\")\n}\n```"},
		{"Column", "QtQuick", "Stacks children vertically",
			"Positions children in a single vertical column. Each child's `y` is set automatically; do not set it yourself.\n\n**Common properties:** `spacing`, `padding`, `topPadding`, `bottomPadding`."},
		{"Row", "QtQuick", "Lays out children horizontally",
			"Positions children in a single horizontal row. Do not set child `x` manually.\n\n**Common properties:** `spacing`, `padding`, `leftPadding`, `rightPadding`, `layoutDirection`."},
		{"Grid", "QtQuick", "Positions children in a grid",
			"Arranges children in a grid defined by `rows`/`columns` and `spacing`.\n\n**Common properties:** `rows`, `columns`, `rowSpacing`, `columnSpacing`, `flow`, `layoutDirection`."},
		{"Flow", "QtQuick", "Flow layout (wrapping)",
			"Like `Row` but wraps children onto new lines when they run out of room.\n\n**Common properties:** `spacing`, `flow`, `layoutDirection`."},
		{"ListView", "QtQuick", "Vertically or horizontally scrolling list",
			"Displays items from a `model` using a `delegate` component. Supports flicking, highlight tracking, sections and headers.\n\n**Common properties:** `model`, `delegate`, `orientation`, `spacing`, `currentIndex`, `highlight`, `header`, `footer`, `section.property`."},
		{"GridView", "QtQuick", "Grid-based data view",
			"Like `ListView` but lays delegates out on a grid. Set `cellWidth` and `cellHeight`.\n\n**Common properties:** `model`, `delegate`, `cellWidth`, `cellHeight`, `flow`, `currentIndex`."},
		{"Repeater", "QtQuick", "Instantiates items from a model",
			"Creates multiple items from a model, each parented to the Repeater's parent. Useful for static lists inside a layout.\n\n**Common properties:** `model`, `delegate`, `count`."},
		{"ListModel", "QtQml.Models", "In-QML list data model",
			"Defines list data inline with `ListElement` children. Supports append/insert/remove/move from JavaScript.\n\n```qml\nListModel {\n    ListElement { name: \"Alice\"; age: 30 }\n    ListElement { name: \"Bob\";   age: 25 }\n}\n```"},
		{"ListElement", "QtQml.Models", "Row in a ListModel",
			"A single row inside a `ListModel`. Property values must be literals — no expressions or binding."},
		{"Component", "QtQml", "Reusable object prototype",
			"Defines a lazily instantiated object prototype. Often used inline as a delegate or loaded via `Loader`.\n\n```qml\nComponent {\n    id: myDelegate\n    Rectangle { color: \"red\" }\n}\n```"},
		{"QtObject", "QtQml", "Non-visual base object",
			"A minimal, non-visual QML object. Use it when you need an object with custom properties/signals but no visual representation."},
		{"Timer", "QtQml", "Interval timer",
			"Fires `triggered()` on an interval. Set `running: true` to start, `repeat: true` to keep firing.\n\n```qml\nTimer {\n    interval: 1000; running: true; repeat: true\n    onTriggered: tick++\n}\n```"},
		{"State", "QtQuick", "Named set of property overrides",
			"Declares a named state. When the containing item's `state` property matches this state's `name`, the listed `PropertyChanges` are applied.\n\n**Common properties:** `name`, `when`, `extend`, `changes`."},
		{"PropertyChanges", "QtQuick", "Property overrides in a State",
			"Applied while a `State` is active. Assigns new values to properties on a target object.\n\n```qml\nPropertyChanges { target: box; color: \"red\" }\n```"},
		{"Transition", "QtQuick", "Animated transition between states",
			"Animates property changes when the item moves between states. Match source/target states with `from`/`to` (or `\"*\"`).\n\n```qml\nTransition {\n    from: \"*\"; to: \"active\"\n    NumberAnimation { properties: \"opacity\"; duration: 200 }\n}\n```"},
		{"Behavior", "QtQuick", "Default animation for a property",
			"Attaches a default animation that runs whenever the named property changes.\n\n```qml\nBehavior on opacity { NumberAnimation { duration: 150 } }\n```"},
		{"NumberAnimation", "QtQuick", "Animates a numeric property",
			"Interpolates a number between `from` and `to` over `duration` ms. Accepts `easing.type`."},
		{"PropertyAnimation", "QtQuick", "Animates any property",
			"Generic property animation. Prefer `NumberAnimation`/`ColorAnimation` for type-specific easing."},
		{"ColorAnimation", "QtQuick", "Animates a color property",
			"Interpolates between two colors in RGB space."},
		{"SequentialAnimation", "QtQuick", "Runs child animations in order", "Plays its children one after another."},
		{"ParallelAnimation", "QtQuick", "Runs child animations in parallel", "Plays all children simultaneously."},
		{"PauseAnimation", "QtQuick", "Adds a delay inside an animation sequence", "Waits `duration` ms. Useful inside `SequentialAnimation`."},
		{"Loader", "QtQuick", "Lazily loads a Component or URL",
			"Instantiates an item from either a `Component`, QML `source` URL, or on demand when `active: true`.\n\n**Common properties:** `source`, `sourceComponent`, `active`, `asynchronous`, `item`."},
		{"FocusScope", "QtQuick", "Keyboard focus boundary",
			"Isolates focus-moving key events to its subtree."},
		{"Keys", "QtQuick", "Attached key event handler",
			"Attached signals for key presses on any `Item`.\n\n```qml\nItem { Keys.onReturnPressed: accept() }\n```"},
		{"ColumnLayout", "QtQuick.Layouts", "Vertical flex layout",
			"Unlike `Column`, children use `Layout.*` attached properties to declare flex behavior.\n\n**Children commonly set:** `Layout.fillWidth`, `Layout.fillHeight`, `Layout.preferredWidth`, `Layout.preferredHeight`, `Layout.alignment`."},
		{"RowLayout", "QtQuick.Layouts", "Horizontal flex layout",
			"Flex-style horizontal layout. See `ColumnLayout` notes for `Layout.*` attached properties."},
		{"GridLayout", "QtQuick.Layouts", "Grid flex layout",
			"Children are placed by `Layout.row`, `Layout.column`, `Layout.rowSpan`, `Layout.columnSpan`."},
		{"StackLayout", "QtQuick.Layouts", "Z-stacked layout (one child visible)",
			"Only the child at `currentIndex` is visible."},
		{"Button", "QtQuick.Controls", "Push button",
			"A standard clickable button.\n\n**Common properties:** `text`, `icon.source`, `checkable`, `checked`, `enabled`.\n\n**Common signals:** `onClicked`, `onPressed`, `onReleased`."},
		{"TextField", "QtQuick.Controls", "Single-line text input",
			"A single-line editable text field.\n\n**Common properties:** `text`, `placeholderText`, `echoMode`, `validator`, `readOnly`."},
		{"TextArea", "QtQuick.Controls", "Multi-line text editor",
			"Multi-line editable text.\n\n**Common properties:** `text`, `placeholderText`, `wrapMode`, `readOnly`."},
		{"Label", "QtQuick.Controls", "Styled Text",
			"A `Text` subclass that follows the control theme's font."},
		{"CheckBox", "QtQuick.Controls", "Check box control", "A toggleable checkbox. Use `checked` and `onToggled`."},
		{"RadioButton", "QtQuick.Controls", "Radio button", "Exclusive option in a `ButtonGroup`."},
		{"Switch", "QtQuick.Controls", "On/off switch", "A toggle switch. Use `checked` and `onToggled`."},
		{"Slider", "QtQuick.Controls", "Value slider", "Set `from`, `to`, and bind `value`."},
		{"ComboBox", "QtQuick.Controls", "Dropdown selector", "Set `model`; read `currentIndex`/`currentText`."},
		{"ProgressBar", "QtQuick.Controls", "Progress indicator", "Bind `value` (0.0–1.0) or set `indeterminate: true`."},
		{"ScrollView", "QtQuick.Controls", "Scrollable viewport", "Wraps a single child in a scrollable area with scrollbars."},
		{"StackView", "QtQuick.Controls", "Stack-based navigation",
			"Push/pop pages. Use `push`, `pop`, `replace`."},
		{"SwipeView", "QtQuick.Controls", "Swipeable pages", "Horizontally paged content. Pair with `PageIndicator`."},
		{"TabBar", "QtQuick.Controls", "Tab strip",
			"A strip of `TabButton` children. Bind `currentIndex` to a `StackLayout`."},
		{"TabButton", "QtQuick.Controls", "Tab in a TabBar", "A tab inside a `TabBar`."},
		{"Dialog", "QtQuick.Controls", "Modal or modeless dialog", "Use `open()`/`close()`. Contains a `ContentItem` and standard buttons."},
		{"Popup", "QtQuick.Controls", "Floating popup", "Base type for dialogs, menus and tooltips. Use `open()`/`close()`."},
		{"Menu", "QtQuick.Controls", "Popup menu", "Contains `MenuItem` children."},
		{"MenuItem", "QtQuick.Controls", "Entry in a Menu", "Selectable item inside a `Menu`. Has `text` and `triggered()`."},
		{"Window", "QtQuick", "Top-level window",
			"A standalone window. In QML apps you typically have a top-level `Window` or `ApplicationWindow`.\n\n**Common properties:** `title`, `width`, `height`, `visible`, `color`, `flags`."},
		{"ApplicationWindow", "QtQuick.Controls", "Themed top-level window",
			"A themed `Window` with `header`, `footer`, `menuBar`, and `contentItem` slots."},
	}
	for _, x := range types {
		registerSymbols(QMLSymbol{
			Label:       x.name,
			Kind:        t,
			Detail:      fmt.Sprintf("%s — %s", x.module, x.detail),
			Signature:   x.name + " { ... }",
			Description: x.desc,
			Module:      x.module,
			Category:    "type",
		})
	}
}

func registerBuiltinProperties() {
	k := lsp.CompletionItemKindProperty
	props := []struct {
		label, typ, detail, desc string
	}{
		{"id", "string", "Unique identifier", "Gives this object an id by which siblings and descendants can reference it. Must be unique within the file and start with a lowercase letter.\n\n```qml\nid: root\n```"},
		{"x", "real", "X position in parent coordinates", "The item's horizontal position relative to its parent."},
		{"y", "real", "Y position in parent coordinates", "The item's vertical position relative to its parent."},
		{"z", "real", "Stacking order", "Z-order among siblings. Higher values are drawn on top."},
		{"width", "real", "Item width in px", "The item's width in pixels."},
		{"height", "real", "Item height in px", "The item's height in pixels."},
		{"visible", "bool", "Whether the item is shown", "If `false`, the item and its children are not rendered and do not receive input."},
		{"enabled", "bool", "Whether the item accepts input", "If `false`, the item and its children receive no input events."},
		{"opacity", "real", "Opacity 0.0–1.0", "Multiplies through to children. Use 0 to hide without disabling."},
		{"rotation", "real", "Rotation (degrees) around `transformOrigin`", "Rotation applied to the item, in degrees."},
		{"scale", "real", "Scale factor around `transformOrigin`", "Uniform scale. 1.0 is natural size."},
		{"transformOrigin", "enumeration", "Pivot for rotation/scale", "One of `Item.TopLeft`, `Item.Top`, `Item.TopRight`, `Item.Left`, `Item.Center`, `Item.Right`, `Item.BottomLeft`, `Item.Bottom`, `Item.BottomRight`."},
		{"clip", "bool", "Whether children are clipped to bounds", "When `true`, descendants are clipped to this item's rectangle."},
		{"focus", "bool", "Whether this item wants active focus", "Combined with the FocusScope it lives under, determines whether key events arrive here."},
		{"activeFocus", "bool (read-only)", "Whether this item currently has focus", "Read-only: `true` when the item is receiving key events."},
		{"anchors", "group", "Anchor layout properties", "Group property — attach children via `anchors.fill`, `anchors.centerIn`, `anchors.top`, etc."},
		{"parent", "Item (read-only)", "Parent item reference", "Reference to this item's parent. Changing assignment re-parents the item."},
		{"children", "list<Item>", "Visual children of this item", "Visual children. Assigning re-parents."},
		{"color", "color", "Fill color", "Color value. Accepts `\"#RRGGBB\"`, `\"#AARRGGBB\"`, SVG color names, or `Qt.rgba(r,g,b,a)`."},
		{"text", "string", "Displayed text", "The text string shown by the item."},
		{"font", "font", "Font group property", "Group property: `font.family`, `font.pixelSize`, `font.pointSize`, `font.bold`, `font.italic`, `font.weight`."},
		{"radius", "real", "Corner radius in px", "Rounded-corner radius (for `Rectangle`)."},
		{"source", "url", "Source URL/path", "URL of the image, component or data to load."},
		{"model", "any", "Data model", "Model for views/repeaters. Can be an int, list, `ListModel`, or any C++ model."},
		{"delegate", "Component", "Delegate component", "The `Component` used to instantiate each item in the view."},
		{"currentIndex", "int", "Current item index", "Index of the currently selected/highlighted item."},
		{"count", "int (read-only)", "Number of items", "Count of items in the model."},
		{"spacing", "real", "Spacing between children", "Pixel spacing between laid-out children."},
		{"onClicked", "signal", "Click handler", "Handler called when the item is clicked. Available on `MouseArea`, `Button`, etc."},
		{"onPressed", "signal", "Press handler", "Handler called when the pointer is pressed."},
		{"onReleased", "signal", "Release handler", "Handler called when the pointer is released."},
		{"onEntered", "signal", "Pointer enter handler", "Fired when the pointer enters the item. Requires `hoverEnabled: true` on `MouseArea`."},
		{"onExited", "signal", "Pointer exit handler", "Fired when the pointer leaves the item. Requires `hoverEnabled: true`."},
		{"onTriggered", "signal", "Trigger handler", "Fired by `Timer.triggered`, `Action.triggered`, etc."},
		{"Layout.fillWidth", "bool", "Fill horizontal space in a Layout", "In a `RowLayout`/`ColumnLayout`/`GridLayout`, expands to take available horizontal space."},
		{"Layout.fillHeight", "bool", "Fill vertical space in a Layout", "In a `RowLayout`/`ColumnLayout`/`GridLayout`, expands to take available vertical space."},
		{"Layout.alignment", "enumeration", "Alignment within Layout cell", "Combination of `Qt.AlignLeft`, `Qt.AlignRight`, `Qt.AlignHCenter`, `Qt.AlignTop`, `Qt.AlignBottom`, `Qt.AlignVCenter`."},
		{"Layout.columnSpan", "int", "Columns spanned in a GridLayout", "How many columns this item spans inside a `GridLayout`."},
		{"Layout.rowSpan", "int", "Rows spanned in a GridLayout", "How many rows this item spans inside a `GridLayout`."},
		{"Layout.preferredWidth", "real", "Preferred width in a Layout", "Suggested width when the layout can honor it."},
		{"Layout.preferredHeight", "real", "Preferred height in a Layout", "Suggested height when the layout can honor it."},
		{"Layout.minimumWidth", "real", "Minimum width in a Layout", "Minimum width honored by the layout."},
		{"Layout.maximumWidth", "real", "Maximum width in a Layout", "Maximum width honored by the layout."},
	}
	for _, p := range props {
		registerSymbols(QMLSymbol{
			Label:       p.label,
			Kind:        k,
			Detail:      fmt.Sprintf("%s — %s", p.typ, p.detail),
			Signature:   fmt.Sprintf("%s: %s", p.label, p.typ),
			Description: p.desc,
			Module:      "",
			Category:    "property",
			InsertText:  p.label + ": ",
		})
	}
}

func registerAnchors() {
	k := lsp.CompletionItemKindProperty
	anchors := []struct {
		label, insert, detail, desc string
	}{
		{"fill", "fill: parent", "Fill the target item",
			"Stretches this item to match the target's geometry on all four sides."},
		{"centerIn", "centerIn: parent", "Center within the target",
			"Centers this item horizontally and vertically within the target."},
		{"top", "top: parent.top", "Anchor the top edge", "Binds this item's top edge to another item's edge."},
		{"bottom", "bottom: parent.bottom", "Anchor the bottom edge", "Binds this item's bottom edge."},
		{"left", "left: parent.left", "Anchor the left edge", "Binds this item's left edge."},
		{"right", "right: parent.right", "Anchor the right edge", "Binds this item's right edge."},
		{"horizontalCenter", "horizontalCenter: parent.horizontalCenter", "Center horizontally",
			"Aligns the item's horizontal center with another item's horizontal center."},
		{"verticalCenter", "verticalCenter: parent.verticalCenter", "Center vertically",
			"Aligns the item's vertical center with another item's vertical center."},
		{"margins", "margins: ${1:10}", "Margin on every anchored edge",
			"Applied to all anchored edges. Specific edges (`topMargin`, etc.) override this."},
		{"topMargin", "topMargin: ${1:10}", "Margin on the top edge", "Overrides `margins` on the top edge."},
		{"bottomMargin", "bottomMargin: ${1:10}", "Margin on the bottom edge", "Overrides `margins` on the bottom edge."},
		{"leftMargin", "leftMargin: ${1:10}", "Margin on the left edge", "Overrides `margins` on the left edge."},
		{"rightMargin", "rightMargin: ${1:10}", "Margin on the right edge", "Overrides `margins` on the right edge."},
	}
	for _, a := range anchors {
		registerSymbols(QMLSymbol{
			Label:         a.label,
			Kind:          k,
			Detail:        a.detail,
			Signature:     "anchors." + a.insert,
			Description:   a.desc,
			Category:      "anchor",
			InsertText:    a.insert,
			InsertSnippet: strings.Contains(a.insert, "${"),
		})
	}
}

func registerKeywords() {
	k := lsp.CompletionItemKindKeyword
	keywords := []QMLSymbol{
		{
			Label: "import", Kind: k, Detail: "Import a QML module",
			Signature:   "import <module> [<version>] [as <alias>]",
			Description: "Brings a module's types into scope. Must appear at the top of the file, before any object declarations.\n\n```qml\nimport QtQuick\nimport QtQuick.Controls\nimport QtQuick.Layouts as L\nimport \"./components\"\n```",
			Category:    "keyword",
			InsertText:  "import ",
		},
		{
			Label: "property", Kind: k, Detail: "Declare a property",
			Signature:   "property <type> <name>: <value>",
			Description: "Adds a new property to the surrounding object. Valid types include: `int`, `real`, `string`, `bool`, `color`, `url`, `var`, `list<T>`, any QML type, or `alias` (a pass-through reference).\n\n```qml\nproperty int count: 0\nproperty alias label: myText.text\nproperty list<Item> panels\n```",
			Category:    "keyword",
			InsertText:  "property ",
		},
		{
			Label: "readonly property", Kind: k, Detail: "Declare a read-only property",
			Signature:   "readonly property <type> <name>: <value>",
			Description: "A property that cannot be reassigned after its initial value is set.",
			Category:    "keyword",
			InsertText:  "readonly property ",
		},
		{
			Label: "required property", Kind: k, Detail: "Declare a required property",
			Signature:   "required property <type> <name>",
			Description: "A property that callers **must** supply. The QML engine throws if it's missing when the component is instantiated. Common in delegates that consume model roles.",
			Category:    "keyword",
			InsertText:  "required property ",
		},
		{
			Label: "default property", Kind: k, Detail: "Declare the default property",
			Signature:   "default property <type> <name>",
			Description: "The property that receives child objects when none is explicitly named. For `Item`, the default property is `data` (the list of children).",
			Category:    "keyword",
			InsertText:  "default property ",
		},
		{
			Label: "signal", Kind: k, Detail: "Declare a signal",
			Signature:   "signal <name>(<type> <param>, ...)",
			Description: "Declares a signal that can be emitted with `emitName()`. Listeners attach via `on<Name>` handlers.\n\n```qml\nsignal accepted(string value)\n```",
			Category:    "keyword",
			InsertText:  "signal ",
		},
		{
			Label: "function", Kind: k, Detail: "Declare a method",
			Signature:   "function name(param1, param2): returnType { ... }",
			Description: "Declares a JavaScript method on the surrounding object. Typed parameters and return type are optional but recommended.\n\n```qml\nfunction add(a: int, b: int): int {\n    return a + b\n}\n```",
			Category:    "keyword",
			InsertText:  "function ${1:name}(${2:params}) {\n\t$0\n}",
		},
		{
			Label: "component", Kind: k, Detail: "Declare an inline Component type",
			Signature:   "component <Name>: <BaseType> { ... }",
			Description: "Qt 6+ syntax for declaring a named inline component inside a QML file.\n\n```qml\ncomponent Badge: Rectangle {\n    color: \"red\"\n    width: 20; height: 20\n}\n```",
			Category:    "keyword",
			InsertText:  "component ",
		},
		{
			Label: "pragma", Kind: k, Detail: "File-level pragma",
			Signature:   "pragma <Name>",
			Description: "File-level directive. Common pragmas: `Singleton`, `NativeMethodBehavior: AcceptThisObject`, `ComponentBehavior: Bound`, `FunctionSignatureBehavior: Enforced`.",
			Category:    "keyword",
			InsertText:  "pragma ",
		},
		{
			Label: "as", Kind: k, Detail: "Import alias",
			Description: "Used inside an `import` statement to rename a module.",
			Category:    "keyword",
			InsertText:  "as ",
		},
		{
			Label: "on", Kind: k, Detail: "Behavior target specifier",
			Description: "Used in `Behavior on <property>` and `PropertyAnimation on <property>` to target a specific property.",
			Category:    "keyword",
		},
	}
	registerSymbols(keywords...)
}

func registerImports() {
	k := lsp.CompletionItemKindModule
	imports := []struct {
		name, detail, desc string
	}{
		{"QtQuick", "Core Qt Quick types",
			"Core visual types: `Item`, `Rectangle`, `Text`, `Image`, `MouseArea`, `Column`, `Row`, `ListView`, animations, transitions, and more."},
		{"QtQuick.Controls", "Qt Quick Controls",
			"Themed UI controls: `Button`, `Label`, `TextField`, `ComboBox`, `Slider`, `TabBar`, `StackView`, `ApplicationWindow`, etc."},
		{"QtQuick.Layouts", "Flex-style layouts",
			"`RowLayout`, `ColumnLayout`, `GridLayout`, `StackLayout` with the `Layout.*` attached properties."},
		{"QtQuick.Window", "Window types",
			"Top-level `Window` and `Screen`."},
		{"QtQuick.Dialogs", "Native-styled dialogs",
			"`FileDialog`, `ColorDialog`, `FontDialog`, `MessageDialog`."},
		{"QtQuick.Shapes", "Vector shapes",
			"`Shape`, `ShapePath`, `PathLine`, etc. for arbitrary 2D geometry."},
		{"QtQuick.Templates", "Controls without a style",
			"Non-styled base types for custom Controls styles."},
		{"QtQml", "Core QML engine types",
			"`QtObject`, `Component`, `Binding`, `Connections`, `Timer`."},
		{"QtQml.Models", "Model types",
			"`ListModel`, `ListElement`, `DelegateModel`, `ObjectModel`."},
	}
	for _, m := range imports {
		registerSymbols(QMLSymbol{
			Label:       m.name,
			Kind:        k,
			Detail:      m.detail,
			Signature:   "import " + m.name,
			Description: m.desc,
			Module:      m.name,
			Category:    "import",
		})
	}
}

func registerJSBuiltins() {
	fnKind := lsp.CompletionItemKindFunction
	objKind := lsp.CompletionItemKindModule
	entries := []QMLSymbol{
		{
			Label: "console", Kind: objKind, Detail: "Console logging object",
			Description: "Debug logging API.\n\n- `console.log(...)` — info-level output\n- `console.info(...)`\n- `console.warn(...)`\n- `console.error(...)`\n- `console.debug(...)`\n- `console.trace()` — current QML stack trace\n- `console.time(label)` / `console.timeEnd(label)`",
			Category:    "js",
		},
		{
			Label: "Math", Kind: objKind, Detail: "JavaScript Math object",
			Description: "Standard JS Math.\n\n**Constants:** `Math.PI`, `Math.E`, `Math.LN2`, `Math.LN10`, `Math.LOG2E`, `Math.LOG10E`, `Math.SQRT2`, `Math.SQRT1_2`.\n\n**Methods:** `abs`, `acos`, `asin`, `atan`, `atan2`, `ceil`, `cos`, `exp`, `floor`, `log`, `log2`, `log10`, `max`, `min`, `pow`, `random`, `round`, `sign`, `sin`, `sqrt`, `tan`, `trunc`, `hypot`, `cbrt`.",
			Category:    "js",
		},
		{
			Label: "JSON", Kind: objKind, Detail: "JSON parse/stringify",
			Description: "`JSON.parse(text)` and `JSON.stringify(value[, replacer, space])`.",
			Category:    "js",
		},
		{
			Label: "Date", Kind: objKind, Detail: "JavaScript Date",
			Description: "`new Date()`, `Date.now()`, `Date.parse(str)`. Instance methods: `getFullYear`, `getMonth`, `getDate`, `getHours`, `getTime`, `toISOString`, `toLocaleDateString`, `toLocaleTimeString`.",
			Category:    "js",
		},
		{
			Label: "Array", Kind: objKind, Detail: "JavaScript Array",
			Description: "`Array.isArray(v)`, `Array.from(iter)`, `Array.of(...items)`.\n\n**Instance methods:** `push`, `pop`, `shift`, `unshift`, `slice`, `splice`, `concat`, `join`, `map`, `filter`, `reduce`, `forEach`, `some`, `every`, `find`, `findIndex`, `includes`, `indexOf`, `sort`, `reverse`, `flat`, `flatMap`.",
			Category:    "js",
		},
		{
			Label: "Object", Kind: objKind, Detail: "JavaScript Object",
			Description: "`Object.keys(o)`, `Object.values(o)`, `Object.entries(o)`, `Object.assign(target, ...src)`, `Object.freeze(o)`, `Object.fromEntries(entries)`.",
			Category:    "js",
		},
		{
			Label: "String", Kind: fnKind, Detail: "String(value) — coerce to string",
			Signature:   "String(value: any): string",
			Description: "Converts any value to a string. Also used as a namespace for `String.fromCharCode(n)`.",
			Category:    "js",
			InsertText:  "String(${1:value})",
		},
		{
			Label: "Number", Kind: fnKind, Detail: "Number(value) — coerce to number",
			Signature:   "Number(value: any): number",
			Description: "Converts any value to a number. Also a namespace for `Number.isFinite`, `Number.isInteger`, `Number.parseFloat`, `Number.parseInt`, `Number.MAX_SAFE_INTEGER`, etc.",
			Category:    "js",
			InsertText:  "Number(${1:value})",
		},
		{
			Label: "Boolean", Kind: fnKind, Detail: "Boolean(value) — coerce to bool",
			Signature:   "Boolean(value: any): bool",
			Description: "Converts any value to a boolean.",
			Category:    "js",
			InsertText:  "Boolean(${1:value})",
		},
		{
			Label: "parseInt", Kind: fnKind, Detail: "Parse an integer from a string",
			Signature:   "parseInt(s: string, radix?: int): int",
			Description: "Parses the leading integer from `s` in the given `radix` (2–36, default 10).",
			Category:    "js",
			InsertText:  "parseInt(${1:s})",
		},
		{
			Label: "parseFloat", Kind: fnKind, Detail: "Parse a float from a string",
			Signature:   "parseFloat(s: string): number",
			Description: "Parses the leading floating-point value from `s`.",
			Category:    "js",
			InsertText:  "parseFloat(${1:s})",
		},
		{
			Label: "isNaN", Kind: fnKind, Detail: "Check for NaN",
			Signature:   "isNaN(value: any): bool",
			Description: "Returns `true` if `value` is `NaN` after numeric coercion. Prefer `Number.isNaN` to avoid coercion surprises.",
			Category:    "js",
			InsertText:  "isNaN(${1:value})",
		},
		{
			Label: "isFinite", Kind: fnKind, Detail: "Check for finite number",
			Signature:   "isFinite(value: any): bool",
			Description: "Returns `true` if `value` is a finite number.",
			Category:    "js",
			InsertText:  "isFinite(${1:value})",
		},
		{
			Label: "Qt", Kind: objKind, Detail: "Qt global object",
			Description: "QML engine globals.\n\n**Value factories:** `Qt.rect(x, y, w, h)`, `Qt.size(w, h)`, `Qt.point(x, y)`, `Qt.vector2d/3d/4d`, `Qt.quaternion(s, x, y, z)`, `Qt.matrix4x4(...)`.\n\n**Color/font:** `Qt.rgba(r,g,b,a)`, `Qt.hsla(h,s,l,a)`, `Qt.hsva(h,s,v,a)`, `Qt.tint(base, tint)`, `Qt.lighter(c[,factor])`, `Qt.darker(c[,factor])`, `Qt.font(obj)`.\n\n**Utilities:** `Qt.createComponent(url)`, `Qt.createQmlObject(qml, parent)`, `Qt.openUrlExternally(url)`, `Qt.resolvedUrl(url)`, `Qt.application`, `Qt.platform`, `Qt.locale()`, `Qt.formatDate/Time/DateTime`, `Qt.binding(fn)`, `Qt.callLater(fn)`.",
			Category:    "js",
		},
	}
	registerSymbols(entries...)
}

// Back-compat shims — keep the old getTypeInfo/getPropertyInfo signatures working
// for hover code that hasn't been migrated to the registry yet.
func registryTypeInfo(name string) (QMLTypeInfo, bool) {
	s, ok := lookupSymbol(name)
	if !ok || (s.Category != "type" && s.Category != "workspace") {
		return QMLTypeInfo{}, false
	}
	return QMLTypeInfo{
		Description: s.Description,
		Type:        "Object",
		Module:      s.Module,
	}, true
}

func registryPropertyInfo(name string) (PropertyInfo, bool) {
	s, ok := lookupSymbol(name)
	if !ok || (s.Category != "property" && s.Category != "anchor") {
		return PropertyInfo{}, false
	}
	typ := ""
	if idx := strings.Index(s.Detail, " — "); idx > 0 {
		typ = s.Detail[:idx]
	}
	return PropertyInfo{
		Description: s.Description,
		Type:        typ,
	}, true
}
