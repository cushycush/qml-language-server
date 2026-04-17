# QML Language Server

A Go-based Language Server for QML (Qt Meta-Object Language) that provides intelligent code editing features.

## Features

### Core Language Features
- **Pure-Go Parser** - Powered by [gotreesitter](https://github.com/odvcencio/gotreesitter) with an embedded tree-sitter-qmljs grammar; no CGO, no external tree-sitter install
- **Incremental Parsing** - Reparses only the affected regions on each `didChange`
- **Workspace Indexing** - Scans the project on startup to resolve cross-file symbols, imports, and IDs
- **Qt Module Discovery** - Parses `.qmltypes` and `qmldir` files from your Qt installation to provide completions, hover, and signatures for the full Qt API — not just a hard-coded subset

### LSP Features
- **Hover** - Type documentation, property info, and signal/method details with fallbacks through the workspace index
- **Completions** - Context-aware completions for:
  - QML types — all types from installed Qt modules (QtQuick, QtQml, QtQuick.Controls, QtQuick.Layouts, QtMultimedia, Qt3D, and more)
  - Imports (`import QtQuick`)
  - Properties — generic (`width`, `height`, `color`, `anchors`) and type-specific (`Window.title`, `Text.wrapMode`, `Image.fillMode`); includes inheritance
  - Type-aware member completion on ids — `root.` where `id: root` is a `Rectangle` offers Rectangle + Item properties, not the generic list
  - Signal handlers (`onClicked`, `onPressed`, etc.)
  - Methods and enums from Qt type info
  - Values (`true`, `false`, colors, `parent`, `this`)
  - Anchor completions (`fill`, `centerIn`, `top`, `bottom`, etc.)
  - Quickshell types, imports, singletons, and boilerplate snippets
  - Workspace components (user-defined `.qml` files)
- **Go to Definition** - Jumps to ids in the current file, cross-file to workspace components (e.g. `MyButton` → `MyButton.qml`), and to the originating `import` line for built-in types
- **Document Links** - `import` statements are clickable — named modules jump to the `qmldir` discovered at startup; relative `import "./components"` jumps to the target directory's `qmldir`
- **Find References** - Find all uses of an identifier
- **Diagnostics** - Parse error highlighting from tree-sitter
- **Document Symbols** - File outline with hierarchical structure (properties, bindings, nested objects)
- **Workspace Symbol Search** - Find QML components across the workspace by name
- **Semantic Tokens** - Tree-sitter-driven semantic highlighting for imports, types, properties, signal handlers, keywords, strings, numbers, and comments
- **Document Formatting** - Re-indent based on brace depth, trim trailing whitespace, collapse blank lines, ensure final newline; honours `tabSize` and `insertSpaces`
- **Code Actions** - Quick fixes
- **Rename** - Rename identifiers across a document
- **Signature Help** - Function parameter hints for `Qt.rect()`, `Qt.rgba()`, `console.log()`, `String()`, `parseInt()`, and more
- **Inlay Hints** - Parameter name hints on function call arguments

## Installation

### Arch Linux (AUR)

```bash
# Prebuilt binary (recommended)
yay -S qml-language-server-bin

# Or build from latest main
yay -S qml-language-server-git
```

### Nix

```bash
# Run directly
nix run github:cushycush/qml-language-server

# Or add to your flake inputs
inputs.qml-language-server.url = "github:cushycush/qml-language-server";
```

### Prebuilt Binaries

Download the latest release for your platform from the [Releases page](https://github.com/cushycush/qml-language-server/releases).

### Build from Source

Requires Go 1.26.1+.

```bash
git clone https://github.com/cushycush/qml-language-server.git
cd qml-language-server
make build
```

`make install` will build the binary and copy it to `~/.local/bin`.

## Editor Configuration

### VS Code

1. Install the "Local LSP" extension or create a custom extension
2. Add to your settings.json:

```json
{
  "languageServers": {
    "qml": {
      "command": "qml-language-server",
      "filetypes": ["qml"]
    }
  }
}
```

### Neovim

For Neovim 0.11+, use the built-in LSP configuration:

```lua
vim.lsp.config.qml = {
  name = "qml-language-server",
  filetypes = { "qml" },
  root_dir = function(fname)
    local root_patterns = { '.git', '.qml', 'qmldir' }
    for _, pattern in ipairs(root_patterns) do
      local root = vim.fs.find(pattern, { path = fname, upward = true })[1]
      if root then
        return vim.fs.dirname(root)
      end
    end
    return vim.fs.dirname(fname)
  end,
  cmd = { "qml-language-server" },
}

vim.lsp.enable("qml")
```

For Neovim 0.10 and earlier, use `lspconfig`:

```lua
local lspconfig = require('lspconfig')

lspconfig.qmlls.setup {
  cmd = { "qml-language-server" },
  filetypes = { "qml" },
  root_dir = function(fname)
    return lspconfig.util.find_git_roots(fname) or lspconfig.util.find_root({ '*.qml' }, fname)
  end,
}
```

### Neovim with blink.cmp

For a modern completion experience with fuzzy matching and snippets, use [blink.cmp](https://github.com/saghen/blink.cmp):

```lua
{
  'saghen/blink.cmp',
  opts = {
    sources = {
      default = { 'lsp' },
    },
    completion = {
      documentation = {
        auto_show = true,
      },
    },
  },
}
```

## Development

```bash
make test       # run tests with race detector
make lint       # golangci-lint
make coverage   # generate coverage report
make build      # compile binary
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
