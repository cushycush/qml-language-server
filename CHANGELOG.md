# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/cushycush/qml-language-server/compare/v1.3.1...HEAD
[1.3.1]: https://github.com/cushycush/qml-language-server/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/cushycush/qml-language-server/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/cushycush/qml-language-server/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/cushycush/qml-language-server/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/cushycush/qml-language-server/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/cushycush/qml-language-server/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/cushycush/qml-language-server/releases/tag/v1.0.0
