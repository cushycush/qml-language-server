package handler

import (
	"fmt"

	"github.com/owenrumney/go-lsp/lsp"
)

// typeProperties stores the *additional* properties that belong to a
// specific QML type, beyond the generic Item-style list registered in
// registerBuiltinProperties. Lookup is by enclosing type name; values
// inherit through baseTypes (e.g. ApplicationWindow -> Window -> Item).
var typeProperties = map[string][]QMLSymbol{}

// baseTypes maps a QML type to the chain of bases whose properties also
// apply. Order is significant only for documentation; duplicates are
// collapsed by Label in typePropertyCompletions.
var baseTypes = map[string][]string{
	"ApplicationWindow": {"Window"},
	"Label":             {"Text"},
	"TextField":         {"TextInput"},
	"TextArea":          {"TextEdit"},
	"Button":            {"AbstractButton"},
	"CheckBox":          {"AbstractButton"},
	"RadioButton":       {"AbstractButton"},
	"Switch":            {"AbstractButton"},
	"TabButton":         {"AbstractButton"},
	"GridView":          {"Flickable"},
	"ListView":          {"Flickable"},
	"ScrollView":        {"Flickable"},
	"SwipeView":         {"Flickable"},
}

func init() {
	registerTypeProperties()
}

func registerTypeProperties() {
	addTypeProps("Window", []propSpec{
		{"title", "string", "Window title", "Text shown in the window's title bar."},
		{"flags", "Qt.WindowFlags", "Window flags", "OR-able window hints, e.g. `Qt.FramelessWindowHint | Qt.WindowStaysOnTopHint`."},
		{"modality", "Qt.WindowModality", "Modality", "One of `Qt.NonModal`, `Qt.WindowModal`, `Qt.ApplicationModal`."},
		{"transientParent", "Window", "Transient parent", "The window this one is transient for; influences stacking and centering."},
		{"screen", "Screen", "Target screen", "The screen the window should appear on."},
		{"minimumWidth", "int", "Minimum width", "Lower bound enforced by the window manager."},
		{"minimumHeight", "int", "Minimum height", "Lower bound enforced by the window manager."},
		{"maximumWidth", "int", "Maximum width", "Upper bound enforced by the window manager."},
		{"maximumHeight", "int", "Maximum height", "Upper bound enforced by the window manager."},
		{"visibility", "Window.Visibility", "Window visibility state", "One of `Window.Windowed`, `Window.Maximized`, `Window.FullScreen`, `Window.Hidden`, `Window.Minimized`."},
		{"contentItem", "Item (read-only)", "Root content item", "The implicit `Item` that holds child content."},
		{"active", "bool (read-only)", "Whether the window has focus", "True when this is the active (focused) window."},
	})

	addTypeProps("Text", []propSpec{
		{"elide", "Text.TextElideMode", "How to elide overflow", "One of `Text.ElideNone`, `Text.ElideLeft`, `Text.ElideMiddle`, `Text.ElideRight`."},
		{"wrapMode", "Text.WrapMode", "Word wrap mode", "One of `Text.NoWrap`, `Text.WordWrap`, `Text.WrapAnywhere`, `Text.Wrap`."},
		{"textFormat", "Text.TextFormat", "Plain/rich text", "One of `Text.AutoText`, `Text.PlainText`, `Text.RichText`, `Text.StyledText`, `Text.MarkdownText`."},
		{"horizontalAlignment", "enumeration", "Horizontal alignment", "One of `Text.AlignLeft`, `Text.AlignRight`, `Text.AlignHCenter`, `Text.AlignJustify`."},
		{"verticalAlignment", "enumeration", "Vertical alignment", "One of `Text.AlignTop`, `Text.AlignBottom`, `Text.AlignVCenter`."},
		{"lineHeight", "real", "Line height multiplier or pixel value", "Interpreted by `lineHeightMode`."},
		{"lineHeightMode", "Text.LineHeightMode", "How lineHeight is interpreted", "`Text.ProportionalHeight` or `Text.FixedHeight`."},
		{"lineCount", "int (read-only)", "Number of laid-out lines", "Read-only count after layout."},
		{"maximumLineCount", "int", "Cap on visible lines", "Limits the number of laid-out lines."},
		{"minimumPixelSize", "int", "Min font size when fontSizeMode shrinks", "Used with `fontSizeMode`."},
		{"minimumPointSize", "real", "Min point size when fontSizeMode shrinks", "Used with `fontSizeMode`."},
		{"fontSizeMode", "Text.FontSizeMode", "Auto-size behavior", "`Text.FixedSize`, `Text.HorizontalFit`, `Text.VerticalFit`, `Text.Fit`."},
		{"style", "Text.TextStyle", "Decoration style", "`Text.Normal`, `Text.Outline`, `Text.Raised`, `Text.Sunken`."},
		{"styleColor", "color", "Outline/raised/sunken color", "Color used for the chosen `style`."},
		{"linkColor", "color", "Color of inline links", "Used for clickable links in `RichText`."},
		{"textHovered", "bool (read-only)", "Whether a link is hovered", "True while a `<a>` is hovered."},
		{"contentWidth", "real (read-only)", "Laid-out content width", "Width occupied after wrapping."},
		{"contentHeight", "real (read-only)", "Laid-out content height", "Height occupied after wrapping."},
		{"renderType", "Text.RenderType", "Glyph renderer", "`Text.QtRendering`, `Text.NativeRendering`, `Text.CurveRendering`."},
		{"baseUrl", "url", "Base URL for relative <img> tags", "Used to resolve relative URLs in `RichText`."},
	})

	addTypeProps("Rectangle", []propSpec{
		{"border", "group", "Border properties (color, width, pixelAligned)", "Group: `border.color`, `border.width`, `border.pixelAligned`."},
		{"gradient", "Gradient", "Fill gradient", "Set to a `Gradient { GradientStop {...} ... }` object or a built-in like `Gradient.NightFade`."},
		{"antialiasing", "bool", "Enable per-edge antialiasing", "Costs more to render; on by default for non-zero `radius`."},
	})

	addTypeProps("Image", []propSpec{
		{"fillMode", "Image.FillMode", "How the image fills its rect", "`Image.Stretch`, `Image.PreserveAspectFit`, `Image.PreserveAspectCrop`, `Image.Tile`, `Image.TileVertically`, `Image.TileHorizontally`, `Image.Pad`."},
		{"sourceSize", "size", "Decoded size", "Set to load the image at a specific resolution."},
		{"asynchronous", "bool", "Load on a worker thread", "Avoids blocking the GUI thread on large images."},
		{"cache", "bool", "Cache decoded image", "Disable for large or one-shot images."},
		{"mirror", "bool", "Mirror horizontally", "Horizontally flips the rendered image."},
		{"mirrorVertically", "bool", "Mirror vertically", "Vertically flips the rendered image."},
		{"smooth", "bool", "Smooth filtering", "Linear filtering when scaled."},
		{"mipmap", "bool", "Use mipmaps when downscaling", "Costs memory; improves downscale quality."},
		{"status", "Image.Status (read-only)", "Load status", "`Image.Null`, `Image.Ready`, `Image.Loading`, `Image.Error`."},
		{"progress", "real (read-only)", "Load progress 0.0–1.0", "Updated while `status === Image.Loading`."},
		{"paintedWidth", "real (read-only)", "Actual painted width", "May differ from `width` based on `fillMode`."},
		{"paintedHeight", "real (read-only)", "Actual painted height", "May differ from `height` based on `fillMode`."},
		{"autoTransform", "bool", "Honor EXIF orientation", "Rotates per the source's orientation tag."},
		{"sourceClipRect", "rect", "Crop region of the source", "Subset of the source image to load."},
		{"currentFrame", "int", "Current frame for animated images", "Index into a multi-frame image."},
		{"frameCount", "int (read-only)", "Total frames", "1 for non-animated images."},
	})

	addTypeProps("MouseArea", []propSpec{
		{"hoverEnabled", "bool", "Track pointer position when not pressed", "Required for `onEntered`/`onExited`/`containsMouse`."},
		{"acceptedButtons", "Qt.MouseButtons", "Buttons that fire signals", "OR-able, e.g. `Qt.LeftButton | Qt.RightButton`."},
		{"propagateComposedEvents", "bool", "Forward unhandled events", "Lets composed events fall through to siblings beneath."},
		{"preventStealing", "bool", "Prevent ancestor flickables from stealing", "Useful inside scrollable parents."},
		{"drag", "group", "Drag behavior", "Group: `drag.target`, `drag.axis`, `drag.minimumX`, `drag.maximumX`, ..."},
		{"cursorShape", "Qt.CursorShape", "Cursor over the area", "e.g. `Qt.PointingHandCursor`, `Qt.IBeamCursor`."},
		{"scrollGestureEnabled", "bool", "Enable two-finger scroll gesture", "Generates wheel events from touch."},
		{"pressed", "bool (read-only)", "Currently pressed", "True between press and release."},
		{"containsMouse", "bool (read-only)", "Pointer is over the area", "Requires `hoverEnabled: true`."},
		{"mouseX", "real (read-only)", "Pointer X within the area", "Coordinates local to the MouseArea."},
		{"mouseY", "real (read-only)", "Pointer Y within the area", "Coordinates local to the MouseArea."},
		{"pressedButtons", "Qt.MouseButtons (read-only)", "Currently pressed buttons", "Bitmask of mouse buttons."},
		{"pressAndHoldInterval", "int", "ms before pressAndHold fires", "Defaults to the platform's long-press timeout."},
	})

	addTypeProps("AbstractButton", []propSpec{
		{"checkable", "bool", "Whether the button toggles", "When true, click toggles `checked`."},
		{"checked", "bool", "Toggled state", "Honored when `checkable: true`."},
		{"autoRepeat", "bool", "Re-fire onClicked while held", "Useful for spinner/scroll buttons."},
		{"autoExclusive", "bool", "Behave like a radio in a group", "Only one button in the parent is checked at a time."},
		{"flat", "bool", "Render without a frame", "A subtle visual style."},
		{"highlighted", "bool", "Render as primary/highlighted", "Theme-defined accent style."},
		{"icon", "group", "Icon properties (source, name, color, width, height)", "Group: `icon.source`, `icon.name`, `icon.color`, `icon.width`, `icon.height`."},
		{"display", "AbstractButton.Display", "Icon/text layout", "`AbstractButton.IconOnly`, `TextOnly`, `TextBesideIcon`, `TextUnderIcon`."},
	})

	addTypeProps("Flickable", []propSpec{
		{"contentWidth", "real", "Scrollable content width", "The width of the inner content area."},
		{"contentHeight", "real", "Scrollable content height", "The height of the inner content area."},
		{"contentX", "real", "Horizontal scroll offset", "Top-left of the visible area in content coordinates."},
		{"contentY", "real", "Vertical scroll offset", "Top-left of the visible area in content coordinates."},
		{"flickableDirection", "Flickable.FlickableDirection", "Allowed scroll axes", "`Flickable.AutoFlickDirection`, `Flickable.HorizontalFlick`, `Flickable.VerticalFlick`, `Flickable.HorizontalAndVerticalFlick`."},
		{"boundsBehavior", "Flickable.BoundsBehavior", "Edge behavior", "`Flickable.StopAtBounds`, `Flickable.DragOverBounds`, `Flickable.OvershootBounds`, `Flickable.DragAndOvershootBounds`."},
		{"interactive", "bool", "Whether the user can flick", "When false, scroll only programmatically."},
		{"maximumFlickVelocity", "real", "Cap on flick speed", "Pixels per second."},
		{"flickDeceleration", "real", "Deceleration after a flick", "Pixels per second squared."},
		{"pressDelay", "int", "ms before press is delivered to children", "Helps distinguish flicks from clicks."},
		{"synchronousDrag", "bool", "Begin scrolling immediately on drag", "Skips the press delay."},
		{"moving", "bool (read-only)", "Currently moving", "True while content is moving."},
		{"flicking", "bool (read-only)", "Currently flicking", "True while a flick decay is animating."},
	})

	addTypeProps("ListView", []propSpec{
		{"orientation", "ListView.Orientation", "Vertical or horizontal", "`ListView.Vertical` or `ListView.Horizontal`."},
		{"highlight", "Component", "Highlight delegate", "Component used to render the highlight overlay."},
		{"highlightFollowsCurrentItem", "bool", "Auto-track current item", "Animate the highlight to follow `currentIndex`."},
		{"highlightMoveDuration", "int", "Highlight move animation ms", "Duration of the highlight tween."},
		{"highlightMoveVelocity", "real", "Highlight move speed (px/s)", "Alternative to duration."},
		{"highlightRangeMode", "ListView.HighlightRangeMode", "Snap behavior", "`ListView.NoHighlightRange`, `ApplyRange`, `StrictlyEnforceRange`."},
		{"snapMode", "ListView.SnapMode", "Snapping during flick", "`ListView.NoSnap`, `SnapToItem`, `SnapOneItem`."},
		{"header", "Component", "Header delegate", "Rendered above the first item."},
		{"footer", "Component", "Footer delegate", "Rendered below the last item."},
		{"section.property", "string", "Model role used to define sections", "Group with `section.criteria` and `section.delegate`."},
		{"section.criteria", "ListView.SectionCriteria", "How rows are bucketed into sections", "`ListView.FullString` or `ListView.FirstCharacter`."},
		{"section.delegate", "Component", "Section header delegate", "Renders each section's header."},
		{"cacheBuffer", "int", "Pixels of off-screen items to keep", "Higher = smoother scroll, more memory."},
		{"keyNavigationEnabled", "bool", "Arrow-key navigation", "Defaults to true."},
		{"keyNavigationWraps", "bool", "Wrap at ends", "When true, key nav wraps from last to first."},
	})

	addTypeProps("GridView", []propSpec{
		{"cellWidth", "real", "Per-cell width", "Required."},
		{"cellHeight", "real", "Per-cell height", "Required."},
		{"flow", "GridView.Flow", "Layout flow", "`GridView.FlowLeftToRight` or `FlowTopToBottom`."},
		{"snapMode", "GridView.SnapMode", "Snapping during flick", "`GridView.NoSnap`, `SnapToRow`, `SnapOneRow`."},
		{"highlight", "Component", "Highlight delegate", "Component for the selection highlight."},
		{"cacheBuffer", "int", "Pixels of off-screen items to keep", "Higher = smoother scroll, more memory."},
		{"keyNavigationEnabled", "bool", "Arrow-key navigation", "Defaults to true."},
		{"keyNavigationWraps", "bool", "Wrap at ends", "When true, key nav wraps from last to first."},
	})

	addTypeProps("Loader", []propSpec{
		{"sourceComponent", "Component", "Inline component to load", "Mutually exclusive with `source`."},
		{"active", "bool", "Whether to instantiate", "When false, deletes the loaded item."},
		{"asynchronous", "bool", "Load off the GUI thread", "Avoids blocking on large QML."},
		{"item", "QtObject (read-only)", "The loaded object", "Null until loading completes."},
		{"status", "Loader.Status (read-only)", "Load status", "`Loader.Null`, `Loader.Ready`, `Loader.Loading`, `Loader.Error`."},
		{"progress", "real (read-only)", "Load progress 0.0–1.0", "Updated while `Loader.Loading`."},
	})

	addTypeProps("Timer", []propSpec{
		{"interval", "int", "Interval in ms", "Defaults to 1000."},
		{"running", "bool", "Whether the timer ticks", "Set to true to start."},
		{"repeat", "bool", "Continue after first tick", "When false, fires once and stops."},
		{"triggeredOnStart", "bool", "Fire immediately on start", "Avoids the leading interval delay."},
	})

	addTypeProps("Column", []propSpec{
		{"padding", "real", "Inset on every side", "Shorthand for setting all four padding properties."},
		{"topPadding", "real", "Inset on the top edge", "Overrides `padding` on top."},
		{"bottomPadding", "real", "Inset on the bottom edge", "Overrides `padding` on bottom."},
		{"leftPadding", "real", "Inset on the left edge", "Overrides `padding` on left."},
		{"rightPadding", "real", "Inset on the right edge", "Overrides `padding` on right."},
		{"populate", "Transition", "Transition for initial population", "Played when children are first laid out."},
		{"add", "Transition", "Transition for added children", "Played when a child is added later."},
		{"move", "Transition", "Transition for re-positioned children", "Played when a child's slot changes."},
	})

	addTypeProps("Row", []propSpec{
		{"padding", "real", "Inset on every side", "Shorthand for setting all four padding properties."},
		{"topPadding", "real", "Inset on the top edge", "Overrides `padding` on top."},
		{"bottomPadding", "real", "Inset on the bottom edge", "Overrides `padding` on bottom."},
		{"leftPadding", "real", "Inset on the left edge", "Overrides `padding` on left."},
		{"rightPadding", "real", "Inset on the right edge", "Overrides `padding` on right."},
		{"layoutDirection", "Qt.LayoutDirection", "Left-to-right or right-to-left", "`Qt.LeftToRight` or `Qt.RightToLeft`."},
		{"effectiveLayoutDirection", "Qt.LayoutDirection (read-only)", "Resolved direction", "Honors mirroring inheritance."},
	})

	addTypeProps("Grid", []propSpec{
		{"rows", "int", "Row count", "Set together with `columns` to constrain the grid."},
		{"columns", "int", "Column count", "Set together with `rows` to constrain the grid."},
		{"rowSpacing", "real", "Vertical spacing between rows", "Defaults to `spacing`."},
		{"columnSpacing", "real", "Horizontal spacing between columns", "Defaults to `spacing`."},
		{"flow", "Grid.Flow", "Fill direction", "`Grid.LeftToRight` or `Grid.TopToBottom`."},
		{"layoutDirection", "Qt.LayoutDirection", "Left-to-right or right-to-left", "`Qt.LeftToRight` or `Qt.RightToLeft`."},
		{"horizontalItemAlignment", "enumeration", "Per-cell horizontal alignment", "`Grid.AlignLeft`, `Grid.AlignHCenter`, `Grid.AlignRight`."},
		{"verticalItemAlignment", "enumeration", "Per-cell vertical alignment", "`Grid.AlignTop`, `Grid.AlignVCenter`, `Grid.AlignBottom`."},
	})
}

type propSpec struct {
	label, typ, detail, desc string
}

func addTypeProps(typeName string, specs []propSpec) {
	syms := make([]QMLSymbol, 0, len(specs))
	for _, p := range specs {
		syms = append(syms, QMLSymbol{
			Label:       p.label,
			Kind:        lsp.CompletionItemKindProperty,
			Detail:      fmt.Sprintf("%s — %s (%s)", p.typ, p.detail, typeName),
			Signature:   fmt.Sprintf("%s.%s: %s", typeName, p.label, p.typ),
			Description: p.desc,
			Module:      "",
			Category:    "property",
			InsertText:  p.label + ": ",
		})
	}
	typeProperties[typeName] = append(typeProperties[typeName], syms...)
	// Make labels resolvable for hover/ResolveCompletionItem on clients
	// that strip Documentation, but never overwrite an existing entry — the
	// generic registration in registerBuiltinProperties wins on collisions.
	for _, s := range syms {
		if _, exists := lookupSymbol(s.Label); !exists {
			registerSymbols(s)
		}
	}
}

// typePropertyCompletions returns type-specific properties for `typeName`,
// walking the base-type chain so e.g. ApplicationWindow inherits Window's
// properties. Duplicates by Label are dropped (closest wins).
func typePropertyCompletions(typeName string) []lsp.CompletionItem {
	seen := map[string]bool{}
	var items []lsp.CompletionItem
	for _, t := range append([]string{typeName}, baseTypes[typeName]...) {
		for _, s := range typeProperties[t] {
			if seen[s.Label] {
				continue
			}
			seen[s.Label] = true
			items = append(items, s.CompletionItem())
		}
	}
	return items
}
