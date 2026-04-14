# QML Language Server

A Go-based Language Server for QML (Qt Meta-Object Language) that provides intelligent code editing features.

## Features

### Core Language Features
- **Pure-Go Parser** - Powered by [gotreesitter](https://github.com/odvcencio/gotreesitter) with an embedded tree-sitter-qmljs grammar; no CGO, no external tree-sitter install
- **Incremental Parsing** - Reparses only the affected regions on each `didChange`
- **Workspace Indexing** - Scans the project on startup to resolve cross-file symbols, imports, and IDs

### LSP Features
- **Hover** - Type documentation, property info, and signal/method details with fallbacks through the workspace index
- **Completions** - Context-aware completions for:
  - QML types (QtQuick, QtQml, QtQuick.Controls)
  - Imports (`import QtQuick 2.0`)
  - Properties (`width`, `height`, `color`, `anchors`, etc.)
  - Signal handlers (`onClicked`, `onPressed`, etc.)
  - Values (`true`, `false`, colors, `parent`, `this`)
  - Anchor completions (`fill`, `centerIn`, `top`, `bottom`, etc.)
  - Quickshell types, imports, singletons, and boilerplate snippets
- **Go to Definition** - Jump to identifier definitions across the workspace
- **Find References** - Find all uses of an identifier
- **Diagnostics** - Parse error highlighting from tree-sitter
- **Document Symbols** - File outline with hierarchical structure
- **Code Actions** - Quick fixes
- **Rename** - Rename identifiers across a document
- **Signature Help** - Function parameter hints for:
  - `Qt.rect()`, `Qt.size()`, `Qt.point()`
  - `console.log()`, `console.warn()`, `console.error()`
  - `String()`, `Number()`, `Boolean()`
- **Inlay Hints** - Property name annotations

### Documentation Registry
Built-in documentation for QtQuick, QtQml, QtQuick.Controls, QtQuick.Layouts, and Quickshell:
- Core items: Item, Rectangle, Text, Image, MouseArea, Flickable, Loader, Repeater
- Layouts: Column, Row, Grid, ColumnLayout, RowLayout, GridLayout
- Models & views: ListView, GridView, ListModel, ListElement
- Animations: PropertyAnimation, NumberAnimation, ColorAnimation, Behavior, Transition
- States: State, PropertyChanges, StateGroup
- Quickshell: PanelWindow, FloatingWindow, PopupWindow, Scope, Variants, Process, FileView, WlrLayershell, Hyprland singletons, and more

## Installation

### Prerequisites
- Go 1.26.1+
- A code editor with LSP support (VS Code, Neovim, etc.)

### Build from Source

```bash
git clone https://github.com/cushycush/qml-language-server.git
cd qml-language-server
make build
```

`make install` will build the binary and copy it to `~/.local/bin`.

### Prebuilt Binaries

Download the latest release for your platform from the [Releases page](https://github.com/cushycush/qml-language-server/releases).

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

For Neovim 0.11+, use the built-in LSP configuration:

```lua
-- In your init.lua or lua/config/lsp.lua
vim.lsp.config.qml = {
  name = "qml-language-server",
  filetypes = { "qml" },
  root_dir = function(fname)
    -- Find project root (git, .qml files, or parent directory)
    local root_patterns = { '.git', '.qml', 'qmldir' }
    for _, pattern in ipairs(root_patterns) do
      local root = vim.fs.find(pattern, { path = fname, upward = true })[1]
      if root then
        return vim.fs.dirname(root)
      end
    end
    return vim.fs.dirname(fname)
  end,
  cmd = { "/path/to/qml-language-server" },
  settings = {},
}

-- Enable the language server
vim.lsp.enable("qml")
```

For Neovim 0.10 and earlier, use the `lspconfig` plugin:

```lua
-- Using lspconfig (Neovim 0.10 and earlier)
local lspconfig = require('lspconfig')

lspconfig.qmlls.setup {
  cmd = { "/path/to/qml-language-server" },
  filetypes = { "qml" },
  root_dir = function(fname)
    return lspconfig.util.find_git_roots(fname) or lspconfig.util.find_root({ '*.qml' }, fname)
  end,
}
```

### Neovim with blink.cmp

For a modern completion experience with fuzzy matching and snippets, use [blink.cmp](https://github.com/saghen/blink.cmp):

```lua
-- In your init.lua or lua/config/blink.lua
{
  'saghen/blink.cmp',
  dependencies = {
    'rafamadriz/friendly-snippets', -- Optional: QML snippets
  },
  opts = {
    sources = {
      default = { 'lsp' },
      providers = {
        lsp = {
          name = 'qml',
          module = 'blink.cmp.sources.lsp',
          score_offset = 100, -- Boost LSP completions
        },
      },
    },
    completion = {
      menu = {
        border = 'rounded',
        draw = {
          components = {
            kind_icon = {
              width = 1,
              text = {
                blink.cmp.Icon('kind'),
              },
            },
          },
        },
      },
      documentation = {
        auto_show = true,
        window = {
          border = 'rounded',
        },
      },
    },
    snippets = {
      preset = 'friendly-snippets',
    },
  },
}
```

For the LSP integration with blink.cmp:

```lua
-- Ensure LSP is configured before blink.cmp loads
vim.lsp.config.qml = {
  name = "qml-language-server",
  filetypes = { "qml" },
  cmd = { "/path/to/qml-language-server" },
}

vim.lsp.enable("qml")
```

## Project Structure

```
qml-language-server/
├── main.go                 # Entry point
├── handler/
│   ├── handler.go          # LSP handler + capability registration
│   ├── hover.go            # Hover provider
│   ├── completion.go       # Completion provider
│   ├── definition.go       # Go to definition
│   ├── references.go       # Find references
│   ├── diagnostics.go      # Parse error diagnostics
│   ├── symbols.go          # Document symbols
│   ├── codeactions.go      # Quick fixes
│   ├── rename.go           # Rename refactoring
│   ├── signature.go        # Signature help
│   ├── inlayhints.go       # Inlay annotations
│   ├── parser.go           # Tree-sitter integration
│   ├── positions.go        # LSP <-> byte-offset helpers
│   ├── workspace.go        # Workspace symbol index
│   ├── registry.go         # Documentation registry
│   ├── quickshell.go       # Quickshell types, imports, snippets
│   ├── types.go            # QML type info
│   ├── errors.go           # Error helpers
│   └── handler_test.go     # Unit tests
├── grammars/
│   ├── loader.go           # Grammar loader + scanner wiring
│   ├── qmljs.grammar.json  # Tree-sitter grammar
│   ├── qmljs_scanner.go    # Go port of the external scanner
│   ├── grammar_blobs/      # Cached generated language blob
│   └── queries/            # Highlights + locals queries
├── .github/workflows/      # CI/CD
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
