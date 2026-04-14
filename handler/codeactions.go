package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) CodeAction(_ context.Context, params *lsp.CodeActionParams) ([]lsp.CodeAction, error) {
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

	content := []byte(doc)
	lang := h.parser.Language()

	var actions []lsp.CodeAction
	walkTree(root, func(n *gotreesitter.Node) bool {
		switch n.Type(lang) {
		case "ui_object_definition":
			actions = append(actions, insertionFix(uri, content, n, "Add id property", "id: "))
		case "string", "string_fragment":
			if parent := n.Parent(); parent != nil && parent.Type(lang) == "ui_binding" {
				text := string(content[n.StartByte():n.EndByte()])
				if isColorKeyword(text) && !isQuotedString(text) {
					actions = append(actions, replaceFix(uri, content, n, "Quote as color string", `"`+text+`"`))
				}
			}
		}
		return true
	})
	return actions, nil
}

func insertionFix(uri lsp.DocumentURI, content []byte, node *gotreesitter.Node, title, insert string) lsp.CodeAction {
	kind := lsp.CodeActionQuickFix
	start := byteOffsetToPosition(content, node.StartByte())
	return lsp.CodeAction{
		Title: title,
		Kind:  &kind,
		Edit: &lsp.WorkspaceEdit{
			Changes: map[lsp.DocumentURI][]lsp.TextEdit{
				uri: {{Range: lsp.Range{Start: start, End: start}, NewText: insert}},
			},
		},
	}
}

func replaceFix(uri lsp.DocumentURI, content []byte, node *gotreesitter.Node, title, replacement string) lsp.CodeAction {
	kind := lsp.CodeActionQuickFix
	return lsp.CodeAction{
		Title: title,
		Kind:  &kind,
		Edit: &lsp.WorkspaceEdit{
			Changes: map[lsp.DocumentURI][]lsp.TextEdit{
				uri: {{Range: nodeRange(content, node), NewText: replacement}},
			},
		},
	}
}

var colorKeywords = map[string]struct{}{
	"red": {}, "green": {}, "blue": {}, "white": {}, "black": {},
	"yellow": {}, "cyan": {}, "magenta": {}, "gray": {}, "grey": {},
	"transparent": {},
}

func isColorKeyword(s string) bool {
	_, ok := colorKeywords[s]
	return ok
}

// isColorValue is kept as an alias for isColorKeyword to preserve the public
// surface tested by handler_test.go.
func isColorValue(s string) bool { return isColorKeyword(s) }

func isQuotedString(s string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"'
}
