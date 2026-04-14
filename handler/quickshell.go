package handler

import "github.com/owenrumney/go-lsp/lsp"

// registerQuickshellBuiltins adds Quickshell-specific types, imports, global
// singleton members, and snippet completions to the shared symbol registry.
// Types go into the normal "type" category so they show up alongside QtQuick
// types; imports into "import" so they appear after the `import` keyword;
// singleton members into "js" so they surface as global-scope names;
// snippets into a dedicated "quickshell-snippet" category consumed by
// Completion's default context.
//
// Data sourced from the user's prior quickshell-completions.nvim plugin,
// plus a few common boilerplate snippets.
func registerQuickshellBuiltins() {
	registerQuickshellTypes()
	registerQuickshellImports()
	registerQuickshellSingletons()
	registerQuickshellSnippets()
}

func registerQuickshellTypes() {
	kind := lsp.CompletionItemKindClass
	types := []struct {
		name, module, detail, desc string
	}{
		{"PanelWindow", "Quickshell", "Anchored panel window",
			"Window anchored to a screen edge. The go-to type for bars, docks, and panel overlays on Wayland. Use `anchors.top`/`bottom`/`left`/`right` (bool) to pick edges and `implicitHeight`/`implicitWidth` to size it.\n\n```qml\nPanelWindow {\n    anchors { top: true; left: true; right: true }\n    implicitHeight: 30\n}\n```"},
		{"FloatingWindow", "Quickshell", "Floating desktop window",
			"Top-level movable window. Use `width`/`height`/`visible` like a standard `Window`."},
		{"PopupWindow", "Quickshell", "Popup dialog window",
			"Short-lived popup (menus, tooltips). Positioned relative to an anchor window."},
		{"Scope", "Quickshell", "Root scope component",
			"Non-visual container used as the root of a shell file. Holds windows, services, and singletons without producing a visible item itself."},
		{"Variants", "Quickshell", "Instantiate per model item",
			"Creates one instance of its child delegate per entry in `model`. The delegate must declare `required property var modelData`. Commonly used to spawn one window per screen.\n\n```qml\nVariants {\n    model: Quickshell.screens\n    PanelWindow {\n        required property var modelData\n        screen: modelData\n    }\n}\n```"},
		{"Process", "Quickshell.Io", "Run external command",
			"Asynchronous process runner. Pair with `StdioCollector` children on `stdout`/`stderr` to read output.\n\n```qml\nProcess {\n    command: [\"/bin/sh\", \"-c\", \"date\"]\n    running: true\n    stdout: StdioCollector { onStreamFinished: console.log(text) }\n}\n```"},
		{"FileView", "Quickshell.Io", "Reactive file reader",
			"Watches and reads a file. `text` updates when the file changes on disk; `onTextChanged` fires on each update."},
		{"StdioCollector", "Quickshell.Io", "Collect process output",
			"Child of `Process` that accumulates a stream's bytes. Read `text` in `onStreamFinished`."},
		{"ScreencopyView", "Quickshell.Wayland", "Screen capture texture",
			"Captures a region of a screen and exposes it as a live texture. Useful for wallpaper blur / screen mirroring."},
		{"WlrLayershell", "Quickshell.Wayland", "wlr-layer-shell attached props",
			"Attached properties for configuring a `PanelWindow` as a wlr-layer-shell surface.\n\n```qml\nWlrLayershell.layer: WlrLayer.Top\nWlrLayershell.keyboardFocus: WlrKeyboardFocus.None\nWlrLayershell.exclusionMode: ExclusionMode.Normal\n```"},
		{"WlrLayer", "Quickshell.Wayland", "wlr-layer-shell layer enum",
			"Values: `WlrLayer.Background`, `WlrLayer.Bottom`, `WlrLayer.Top`, `WlrLayer.Overlay`."},
		{"WlrKeyboardFocus", "Quickshell.Wayland", "Layer-shell keyboard focus enum",
			"Values: `WlrKeyboardFocus.None`, `WlrKeyboardFocus.Exclusive`, `WlrKeyboardFocus.OnDemand`."},
		{"ExclusionMode", "Quickshell.Wayland", "Exclusive-zone mode enum",
			"Values: `ExclusionMode.Normal`, `ExclusionMode.Ignore`, `ExclusionMode.Auto`."},
	}
	for _, t := range types {
		registerSymbols(QMLSymbol{
			Label:       t.name,
			Kind:        kind,
			Detail:      t.module + " — " + t.detail,
			Signature:   t.name + " { ... }",
			Description: t.desc,
			Module:      t.module,
			Category:    "type",
		})
	}
}

func registerQuickshellImports() {
	kind := lsp.CompletionItemKindModule
	modules := []struct {
		name, detail, desc string
	}{
		{"Quickshell", "Core Quickshell types",
			"Core types and the `Quickshell` singleton (`screens`, `env`, `execDetached`, …)."},
		{"Quickshell.Hyprland", "Hyprland integration",
			"IPC and state bindings for the Hyprland compositor. Exposes the `Hyprland` singleton."},
		{"Quickshell.Io", "File I/O and processes",
			"`Process`, `StdioCollector`, `FileView`, `Socket`, and related I/O types."},
		{"Quickshell.Wayland", "Wayland protocol bindings",
			"wlr-layer-shell, screencopy, and other Wayland-specific surfaces."},
		{"Quickshell.Widgets", "Additional widgets",
			"Higher-level widget types built on top of Quickshell primitives."},
	}
	for _, m := range modules {
		registerSymbols(QMLSymbol{
			Label:       m.name,
			Kind:        kind,
			Detail:      m.detail,
			Signature:   "import " + m.name,
			Description: m.desc,
			Module:      m.name,
			Category:    "import",
		})
	}
}

func registerQuickshellSingletons() {
	propKind := lsp.CompletionItemKindProperty
	fnKind := lsp.CompletionItemKindFunction
	modKind := lsp.CompletionItemKindModule
	fmtSnippet := lsp.InsertTextFormatSnippet
	_ = fmtSnippet

	entries := []QMLSymbol{
		{
			Label: "Quickshell", Kind: modKind, Detail: "Quickshell global singleton",
			Description: "Global singleton exposing shell-wide state and helpers.\n\n**Properties:** `screens`, `shellDir`.\n**Methods:** `env(name)`, `execDetached(cmd)`, `cachePath(sub)`.",
			Category:    "js",
		},
		{
			Label: "Hyprland", Kind: modKind, Detail: "Hyprland global singleton",
			Description: "IPC bridge to the Hyprland compositor. Requires `import Quickshell.Hyprland`.\n\n**Properties:** `focusedMonitor`, `monitors`, `workspaces`, `activeWorkspace`.",
			Category:    "js",
		},
		{
			Label: "Quickshell.screens", Kind: propKind, Detail: "list<ShellScreen>",
			Signature:   "Quickshell.screens",
			Description: "All screens the shell is aware of. Pair with `Variants` to spawn one window per screen.",
			Category:    "js",
		},
		{
			Label: "Quickshell.shellDir", Kind: propKind, Detail: "string",
			Signature:   "Quickshell.shellDir",
			Description: "Absolute path to the directory containing `shell.qml`. Useful for loading sibling files.",
			Category:    "js",
		},
		{
			Label: "Quickshell.env", Kind: fnKind, Detail: "(name: string) -> string",
			Signature:     "Quickshell.env(name: string): string",
			Description:   "Reads an environment variable. Returns `\"\"` when unset.",
			Category:      "js",
			InsertText:    "Quickshell.env(\"${1:NAME}\")",
			InsertSnippet: true,
		},
		{
			Label: "Quickshell.execDetached", Kind: fnKind, Detail: "(cmd: list<string>) -> void",
			Signature:     "Quickshell.execDetached(cmd: list<string>)",
			Description:   "Runs a command with no lifetime tied to the shell. Fire-and-forget.",
			Category:      "js",
			InsertText:    "Quickshell.execDetached([${1:\"cmd\"}])",
			InsertSnippet: true,
		},
		{
			Label: "Quickshell.cachePath", Kind: fnKind, Detail: "(sub: string) -> string",
			Signature:     "Quickshell.cachePath(sub: string): string",
			Description:   "Returns an absolute path inside the Quickshell cache directory, creating it if needed.",
			Category:      "js",
			InsertText:    "Quickshell.cachePath(\"${1:file}\")",
			InsertSnippet: true,
		},
		{
			Label: "Hyprland.focusedMonitor", Kind: propKind, Detail: "HyprlandMonitor",
			Signature:   "Hyprland.focusedMonitor",
			Description: "The monitor Hyprland currently considers focused.",
			Category:    "js",
		},
		{
			Label: "Hyprland.monitors", Kind: propKind, Detail: "list<HyprlandMonitor>",
			Signature:   "Hyprland.monitors",
			Description: "All monitors Hyprland knows about.",
			Category:    "js",
		},
	}
	registerSymbols(entries...)
}

func registerQuickshellSnippets() {
	kind := lsp.CompletionItemKindSnippet
	snippets := []struct {
		label, body, detail, desc string
	}{
		{
			"qs-scope",
			"import Quickshell\n\nScope {\n\t${0}\n}",
			"Quickshell shell scope",
			"Top-level `Scope` with the Quickshell import.",
		},
		{
			"qs-panel",
			"PanelWindow {\n\tanchors {\n\t\ttop: true\n\t\tleft: true\n\t\tright: true\n\t}\n\timplicitHeight: ${1:30}\n\n\t${0}\n}",
			"PanelWindow (bar / dock)",
			"Top-anchored `PanelWindow` with default height.",
		},
		{
			"qs-float",
			"FloatingWindow {\n\tvisible: ${1:true}\n\twidth: ${2:400}\n\theight: ${3:300}\n\n\t${0}\n}",
			"FloatingWindow",
			"Basic `FloatingWindow` with size.",
		},
		{
			"qs-variants",
			"Variants {\n\tmodel: ${1:Quickshell.screens}\n\n\t${2:ComponentName} {\n\t\trequired property var modelData\n\t\t${0}\n\t}\n}",
			"Variants (one instance per screen)",
			"`Variants` with model and a delegate declaring `required property var modelData`.",
		},
		{
			"qs-process",
			"Process {\n\tid: ${1:proc}\n\tcommand: [${2:\"command\"}]\n\trunning: ${3:false}\n\n\tstdout: StdioCollector {\n\t\tonStreamFinished: {\n\t\t\t${0}\n\t\t}\n\t}\n\n\tonExited: (code) => {\n\t\tif (code !== 0) return\n\t}\n}",
			"Process with stdout collector",
			"`Process` pattern with a `StdioCollector` on stdout and an `onExited` guard.",
		},
		{
			"qs-fileview",
			"FileView {\n\tid: ${1:fileView}\n\tpath: ${2:\"~/.config/example\"}\n\n\tonTextChanged: {\n\t\t${0}\n\t}\n}",
			"FileView watcher",
			"`FileView` watching a path with an `onTextChanged` handler.",
		},
		{
			"qs-layershell",
			"WlrLayershell.layer: WlrLayer.${1:Top}\nWlrLayershell.keyboardFocus: WlrKeyboardFocus.${2:None}\nWlrLayershell.exclusionMode: ExclusionMode.${3:Normal}",
			"WlrLayershell attached properties",
			"Standard `WlrLayershell` configuration block for a `PanelWindow`.",
		},
		{
			"qs-timer",
			"Timer {\n\tid: ${1:timer}\n\tinterval: ${2:1000}\n\trunning: ${3:false}\n\trepeat: ${4:false}\n\n\tonTriggered: {\n\t\t${0}\n\t}\n}",
			"Timer",
			"`Timer` with interval and handler.",
		},
		{
			"qs-prop",
			"property ${1:var} ${2:name}: ${3:null}",
			"Property declaration",
			"QML property declaration with type, name, and default value.",
		},
		{
			"qs-signal",
			"signal ${1:signalName}(${2:type} ${3:param})",
			"Signal declaration",
			"QML signal declaration with typed parameters.",
		},
	}
	for _, s := range snippets {
		registerSymbols(QMLSymbol{
			Label:         s.label,
			Kind:          kind,
			Detail:        s.detail,
			Description:   s.desc,
			Category:      "quickshell-snippet",
			InsertText:    s.body,
			InsertSnippet: true,
		})
	}
}
