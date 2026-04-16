# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go-based Language Server Protocol implementation for QML (Qt Meta-Object Language). The binary speaks LSP over stdio and is intended to be launched by editors (VS Code, Neovim, etc.).

## Commands

- `make build` — compile to `./qml-language-server` with version ldflags
- `make test` / `go test -v -race ./...` — full test suite
- `go test -v -run TestName ./handler` — run a single test
- `make lint` — `golangci-lint run ./...`
- `make coverage` — writes `coverage.out` / `coverage.html`
- `make install` — builds and installs to `~/.local/bin`

Requires Go 1.26.1+ (see `go.mod`).

## Architecture

### Entry point & server lifecycle
`main.go` constructs a `handler.Handler` and calls `Serve`, which wires the `owenrumney/go-lsp` server to stdio. `Handler.Initialize` (handler/handler.go) is the source of truth for advertised LSP capabilities — add a new capability here when implementing a new feature.

### Handler package (`handler/`)
One file per LSP feature (`hover.go`, `completion.go`, `definition.go`, `references.go`, `diagnostics.go`, `symbols.go`, `codeactions.go`, `rename.go`, `signature.go`, `inlayhints.go`). All share:

- `Handler.documents` — `map[DocumentURI]string` kept in sync by `DidOpen`/`DidChange`/`DidSave`/`DidClose`.
- `Handler.parser` (`*QMLParser`, `parser.go`) — wraps gotreesitter. Maintains a per-URI `*gotreesitter.Tree`, reparses incrementally via `ParseIncremental` when a previous tree exists.
- `Handler.workspace` (`*workspaceIndex`, `workspace.go`) — scans workspace roots for `*.qml` files at startup so user-defined components show up in completions and the symbol registry.
- `positions.go` — `positionToByte` / `byteOffsetToPosition` convert between LSP `Position` (line/char) and tree-sitter byte offsets, and `findSmallestNodeAt` resolves a cursor position to an AST node. Use these helpers from feature code.
- `types.go` + `typeproperties.go` — hard-coded `qmlTypes` map and per-type property catalog for built-in QtQuick/QtQml/Controls/Layouts. `quickshell.go` adds Quickshell-specific entries. These serve as fallbacks when qmltypes files aren't available.
- `registry.go` — process-wide `QMLSymbol` registry; `workspace.go`, the static type tables, and qmltypes discovery all publish into it so completion/hover share one lookup path.
- `qmltypes_parser.go` + `qmltypes_types.go` — recursive descent parser for Qt's `.qmltypes` DSL format, producing structured Go types for components, properties, signals, methods, and enums.
- `qmldir_parser.go` — parses `qmldir` files to discover module name and path to `.qmltypes`.
- `qmltypes_discovery.go` — at startup, walks Qt install dirs and `QML_IMPORT_PATH` to find and parse every module's type info, then registers the results into the symbol registry without overwriting existing entries.

### Grammar package (`grammars/`)
Loads the QML tree-sitter grammar into `gotreesitter`. Two embedded assets are combined at startup by `QmljsLanguage()` (grammars/loader.go):

1. `qmljs.grammar.json` — run through `grammargen.ImportGrammarJSON` + `GenerateLanguage` to build the parse tables in pure Go.
2. `grammar_blobs/qmljs.bin` — a gzipped gob-encoded reference `Language`; used only to copy the external-scanner ordering via `AdaptExternalScannerByExternalOrder`.
3. `qmljs_scanner.go` — hand-written Go port of the tree-sitter-qmljs external scanner (automatic semicolon insertion, template literals, regex). Registered via `grammars.RegisterExternalScanner` before the language is detected.

Query files in `grammars/queries/` (`highlights.scm`, `locals.scm`) are embedded and used for highlighting; highlights inherit from the `javascript` query set.

### Diagnostics
Driven purely by tree-sitter: `collectDiagnostics` walks the tree for ERROR/MISSING nodes. `publishDiagnostics` always normalizes `nil` to `[]lsp.Diagnostic{}` before sending — do not remove that, empty-vs-nil matters for some LSP clients.

## Notes for editing

- When adding a new LSP method, update both `Handler.Initialize`'s capabilities and add the method on `Handler`.
- The parser can return `nil` if grammar loading fails; all handler features guard against `h.parser == nil` and `tree == nil`.
- Incremental parse relies on calling `Parse` with the full new document text on every `DidChange`; the server does not apply individual content change ranges.
