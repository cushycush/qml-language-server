# QML Language Server

A Go-based Language Server for QML (Qt Meta-Object Language) that provides intelligent code editing features.

## Features

### Core Language Features
- **QML Parsing** - Powered by [gotreesitter](https://github.com/odvcencio/gotreesitter) with tree-sitter-qmljs grammar
- **Incremental Parsing** - Efficient document changes

### LSP Features
- **Hover** - Type documentation and property info
- **Completions** - Context-aware completions for:
  - QML types (QtQuick, QtQml, QtQuick.Controls)
  - Imports (`import QtQuick 2.0`)
  - Properties (`width`, `height`, `color`, `anchors`, etc.)
  - Signal handlers (`onClicked`, `onPressed`, etc.)
  - Values (`true`, `false`, colors, `parent`, `this`)
  - Anchor completions (`fill`, `centerIn`, `top`, `bottom`, etc.)
- **Go to Definition** - Find identifier definitions
- **Find References** - Find all uses of an identifier
- **Diagnostics** - Parse error highlighting
- **Document Symbols** - File outline with hierarchical structure
- **Code Actions** - Quick fixes
- **Rename** - Rename identifiers across document
- **Signature Help** - Function parameter hints for:
  - `Qt.rect()`, `Qt.size()`, `Qt.point()`
  - `console.log()`, `console.warn()`, `console.error()`
  - `String()`, `Number()`, `Boolean()`
- **Inlay Hints** - Property name annotations

### Documentation
Built-in documentation for 30+ QtQuick types including:
- Item, Rectangle, Text, Image, MouseArea
- Layout types (Column, Row, Grid, ColumnLayout, RowLayout, GridLayout)
- List types (ListView, ListModel, ListElement)
- Animation types (PropertyAnimation, NumberAnimation, ColorAnimation)
- State types (State, PropertyChanges, Transition)
- And more...

## Installation

### Prerequisites
- Go 1.21+
- A code editor with LSP support (VS Code, Neovim, etc.)

### Build from Source

```bash
git clone https://github.com/yourusername/qml-language-server.git
cd qml-language-server
go build -o qml-language-server .
```

### Using Go Install

```bash
go install github.com/yourusername/qml-language-server@latest
```

## Editor Configuration

### VS Code

1. Install the "Local LSP" extension or create a custom extension
2. Add to your settings.json:

```json
{
  "languageServers": {
    "qml": {
      "command": "path/to/qml-language-server",
      "filetypes": ["qml"]
    }
  }
}
```

Or use the built-in Language Server Protocol client with:

```json
{
  "qmlls.path": "/path/to/qml-language-server"
}
```

### Neovim

Add to your `init.lua`:

```lua
local lspconfig = require('lspconfig')

lspconfig.qmlls.setup {
  cmd = { "/path/to/qml-language-server" },
  filetypes = { "qml" },
  root_dir = function(fname)
    return lspconfig.util.find_git_roots(fname) or lspconfig.util.find_root({ '*.qml' }, fname)
  end,
}
```

## Project Structure

```
qml-language-server/
├── main.go              # Entry point
├── handler/
│   ├── handler.go       # Main LSP handler
│   ├── hover.go         # Hover provider
│   ├── completion.go    # Completion provider
│   ├── definition.go     # Go to definition
│   ├── references.go     # Find references
│   ├── diagnostics.go    # Parse error diagnostics
│   ├── symbols.go       # Document symbols
│   ├── codeactions.go   # Quick fixes
│   ├── rename.go        # Rename refactoring
│   ├── signature.go     # Function signatures
│   ├── inlayhints.go    # Inlay annotations
│   ├── parser.go         # Tree-sitter integration
│   ├── types.go         # QML type info
│   ├── errors.go        # Error handling
│   └── handler_test.go  # Unit tests
├── .github/
│   └── workflows/       # CI/CD workflows
└── README.md
```

## Development

### Running Tests

```bash
go test ./... -v
```

### Building

```bash
go build -o qml-language-server .
```

### Project Dependencies

- [go-lsp](https://github.com/owenrumney/go-lsp) - LSP protocol implementation
- [gotreesitter](https://github.com/odvcencio/gotreesitter) - Pure-Go tree-sitter runtime

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

## Acknowledgments

- [gotreesitter](https://github.com/odvcencio/gotreesitter) - Pure-Go tree-sitter runtime
- [tree-sitter-qmljs](https://github.com/yuja/tree-sitter-qmljs) - QML grammar for tree-sitter
- [go-lsp](https://github.com/owenrumney/go-lsp) - Go LSP library
