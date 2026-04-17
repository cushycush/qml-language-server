# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Type inference for id member expressions** — typing `root.` where `id: root` resolves to a known type now completes with that type's properties (and everything it inherits via the prototype chain) rather than the generic property list. Works even while the file is mid-edit and error-recovering, via a textual fallback that finds the enclosing type when tree-sitter can't wrap it in a `ui_object_definition`.

### Added
- **Document links** (`textDocument/documentLink`) — `import QtQuick`, `import QtQuick.Controls`, and relative `import "./components"` statements are now clickable. Named modules jump to the `qmldir` discovered at startup; relative imports jump to the target directory's `qmldir` when present, otherwise to the directory itself. Dotted module names fall back to parent modules when the exact name isn't registered.
- **Cross-file go-to-definition for workspace components** — `gd` on a user-defined component like `MyButton` now jumps to `MyButton.qml` in the workspace rather than staying in the current file. Built-in Qt types still navigate to the originating `import` line.

## [1.6.0] - 2026-04-16

### Added
- **Signature help from qmltypes** — method signatures parsed from `.qmltypes` files are now registered for signature help. Extends coverage from 13 hand-rolled functions to every method Qt exposes. Both bare (`mapToItem`) and dotted (`Item.mapToItem`) call forms are supported.
- **Range formatting** (`textDocument/rangeFormatting`) — format a selected range instead of the whole document.
- **Folding ranges** (`textDocument/foldingRange`) — collapsible regions for object blocks (`{ }`) and multi-line comments.

## [1.5.0] - 2026-04-16

### Added
- **`.qmlls.ini` support** — reads the same config file format Qt's qmlls uses. Projects that generate one (e.g. Quickshell) now get full type discovery without manual qmldir setup. `buildDir` and `importPaths` from the config are merged into the module search.
- **Prototype chain walking** — completions now include inherited properties from the full type hierarchy. `Rectangle` shows `Item`'s properties (`x`, `y`, `width`, `height`, `visible`, `anchors`, etc.), not just its own 4. Chains are built automatically from qmltypes `prototype` fields.

### Fixed
- Dropped `ResolveProvider` — docs are shipped eagerly on every completion item, so the resolve round-trip was unnecessary overhead.
- Grammar load failures are now logged so users can diagnose why the LSP appears to do nothing.

## [1.4.0] - 2026-04-15

### Added
- **qmltypes/qmldir module discovery** — at startup the server walks Qt installation directories (`/usr/lib/qt6/qml/`, `/usr/lib/qt/qml/`, `QML_IMPORT_PATH`, `QML2_IMPORT_PATH`) and parses every `qmldir` + `.qmltypes` pair it finds. Types, properties, signals, methods, and enums from all installed Qt modules are registered into the symbol registry, providing completions and hover for the full Qt API rather than only the ~50 hard-coded types.
- Recursive descent parser for the `.qmltypes` DSL format (supports both Qt5 object-map and Qt6 string-array enum value formats).
- `qmldir` line-based parser extracting module name, typeinfo path, depends, and imports.
- C++ → QML type name mapping (`QColor` → `color`, `QString` → `string`, `double` → `real`, etc.).
- Hard-coded types, keywords, anchors, JS builtins, and Quickshell entries are preserved as fallbacks — qmltypes data augments but never overwrites them.

## [1.3.1] - 2026-04-15

### Fixed
- Inlay hints no longer redundantly echo the property name on every binding. Hints now only appear on function call arguments whose callee has a known signature (`Qt.rect`, `console.log`, etc.), showing the parameter name being filled in.

## [1.3.0] - 2026-04-15

### Added
- `textDocument/formatting` — whitespace-only formatter that re-indents based on brace depth, trims trailing whitespace, collapses runs of blank lines, and ensures exactly one trailing newline. Honours the client's `tabSize` / `insertSpaces` options. String and comment contents are tracked so braces inside them don't move the indent depth; token content is never modified.

## [1.2.0] - 2026-04-15

### Added
- `workspace/symbol` — search for QML components across the workspace by name. Backed by the existing workspace index (refreshed at startup and on `workspace/didChangeWatchedFiles`); query matches case-insensitive substrings of the top-level component name.

## [1.1.0] - 2026-04-15

### Added
- `textDocument/semanticTokens/full` — tree-sitter-driven semantic highlighting that classifies imports (namespace + keyword), object types, property/binding names, signal handlers (`event`), required properties, comments, strings, numbers, regexes, and keyword literals. Multi-line strings/comments are split per line; tokens are delta-encoded per the LSP spec.
- Unit tests covering hover, completion, definition, references, document symbols, diagnostics, rename, signature help, and semantic tokens.

### Fixed
- `DocumentSymbol` now surfaces properties, bindings, and nested objects as children of their enclosing object — previously it walked one layer too shallow and returned classes with empty children.

### Changed
- Internal docs: `CLAUDE.md` rewritten to reflect the current handler layout (workspace index, registry, no `cache.go`).

## [1.0.2] - 2026

### Added
- Type-aware property completions inside object bodies — the completion list now reflects the enclosing QML type (`Window`, `Text`, `Rectangle`, `ApplicationWindow`, etc.) and includes inherited properties.

### Changed
- CI: bumped GitHub Actions to Node.js 24-compatible versions; pinned `golangci-lint-action` to a Node 24 runtime.

## [1.0.1] - 2026

### Fixed
- Release workflow: granted `contents: write` so tag releases can publish artifacts; dropped the docker job; unified the zip step.
- Lint findings flagged by golangci-lint v2.
- CI: bumped the `go` directive to 1.25, upgraded the golangci-lint action, and passed `VERSION` through to the release build.

## [1.0.0] - 2025

First tagged release.

### Added
- LSP server speaking over stdio with capabilities for hover, completion, go-to-definition, find references, document symbols, document highlight, signature help, code actions, rename (with prepare), diagnostics, and inlay hints.
- Tree-sitter QML grammar bundled and loaded in pure Go via `gotreesitter` + `grammargen`, with a hand-ported external scanner (ASI, template literals, regex).
- Built-in type registry covering QtQuick, QtQuick.Controls, QtQuick.Layouts, and QtQml; Quickshell types, imports, singletons, and snippets.
- Workspace index that scans `*.qml` files at startup so user-defined components show up in completions.
- Diagnostics derived from tree-sitter ERROR/MISSING nodes; empty results normalized to `[]` for client compatibility.
- Generated grammar blob cached on disk for fast startup.
- Distribution: GitHub Actions release workflow, Dockerfile, README with installation and Neovim/blink.cmp setup.

[Unreleased]: https://github.com/cushycush/qml-language-server/compare/v1.6.0...HEAD
[1.6.0]: https://github.com/cushycush/qml-language-server/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/cushycush/qml-language-server/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/cushycush/qml-language-server/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/cushycush/qml-language-server/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/cushycush/qml-language-server/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/cushycush/qml-language-server/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/cushycush/qml-language-server/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/cushycush/qml-language-server/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/cushycush/qml-language-server/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/cushycush/qml-language-server/releases/tag/v1.0.0
