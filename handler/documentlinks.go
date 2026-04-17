package handler

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

// DocumentLink returns one clickable link per `import` statement in the
// document. Named modules (e.g. `import QtQuick`) resolve to the qmldir file
// discovered at startup. String-literal imports (e.g. `import "./components"`)
// resolve relative to the current document, preferring a qmldir inside the
// target directory when one exists.
func (h *Handler) DocumentLink(_ context.Context, params *lsp.DocumentLinkParams) ([]lsp.DocumentLink, error) {
	uri := params.TextDocument.URI
	doc, ok := h.getDocument(uri)
	if !ok || h.parser == nil {
		return nil, nil
	}
	tree := h.parser.GetTree(uri)
	if tree == nil {
		return nil, nil
	}
	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

	lang := h.parser.Language()
	content := []byte(doc)
	docDir := filepath.Dir(uriToPath(uri))

	var links []lsp.DocumentLink
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) != "ui_import" {
			return true
		}
		if link, ok := importLink(n, lang, content, docDir); ok {
			links = append(links, link)
		}
		return false
	})
	return links, nil
}

// importLink builds a DocumentLink for a single ui_import node. Returns
// ok=false when the import has no resolvable target.
func importLink(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, docDir string) (lsp.DocumentLink, bool) {
	source := n.ChildByFieldName("source", lang)
	if source == nil {
		return lsp.DocumentLink{}, false
	}
	rng := nodeRange(content, source)
	text := string(content[source.StartByte():source.EndByte()])

	target, tooltip, ok := resolveImportTarget(source.Type(lang), text, docDir)
	if !ok {
		return lsp.DocumentLink{}, false
	}
	targetURI := lsp.DocumentURI(pathToURI(target))
	return lsp.DocumentLink{
		Range:   rng,
		Target:  &targetURI,
		Tooltip: tooltip,
	}, true
}

// resolveImportTarget returns the absolute filesystem path the import should
// navigate to, along with a human-readable tooltip. sourceType is the
// tree-sitter node type of the `source` field: "string" for quoted paths,
// anything else is treated as a qualified module id.
func resolveImportTarget(sourceType, text, docDir string) (target, tooltip string, ok bool) {
	if sourceType == "string" {
		// Strip the surrounding quotes — tree-sitter gives us the whole literal
		// including delimiters.
		rel := strings.Trim(text, "\"'`")
		if rel == "" {
			return "", "", false
		}
		abs := rel
		if !filepath.IsAbs(abs) && docDir != "" {
			abs = filepath.Join(docDir, rel)
		}
		abs = filepath.Clean(abs)
		info, err := os.Stat(abs)
		if err != nil {
			return "", "", false
		}
		if info.IsDir() {
			// Prefer the qmldir inside the directory so the editor opens it
			// rather than trying to render a folder.
			if qd := filepath.Join(abs, "qmldir"); fileExists(qd) {
				return qd, "Open qmldir for " + rel, true
			}
			return abs, "Open directory " + rel, true
		}
		return abs, "Open " + rel, true
	}
	// Qualified module id — may be dotted ("QtQuick.Controls"). Try the full
	// name first, then successively strip trailing segments.
	name := text
	for name != "" {
		if dir := LookupModuleQMLDir(name); dir != "" {
			return dir, "Open qmldir for " + name, true
		}
		idx := strings.LastIndex(name, ".")
		if idx < 0 {
			break
		}
		name = name[:idx]
	}
	return "", "", false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
