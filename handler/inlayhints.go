package handler

import (
	"context"
	"encoding/json"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) InlayHint(_ context.Context, params *lsp.InlayHintParams) ([]lsp.InlayHint, error) {
	doc, ok := h.getDocument(params.TextDocument.URI)
	if !ok || h.parser == nil {
		return nil, nil
	}

	tree := h.parser.GetTree(params.TextDocument.URI)
	if tree == nil {
		return nil, nil
	}
	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

	content := []byte(doc)
	lang := h.parser.Language()

	var hints []lsp.InlayHint
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) != "ui_binding" {
			return true
		}
		name := extractPropertyName(n, lang, content)
		if name == "" {
			return true
		}
		paddingLeft := true
		kind := lsp.InlayHintKind(lsp.InlayHintKindType)
		hints = append(hints, lsp.InlayHint{
			Position:    byteOffsetToPosition(content, n.StartByte()),
			Label:       json.RawMessage(`"` + name + `: "`),
			Kind:        &kind,
			PaddingLeft: &paddingLeft,
		})
		return true
	})
	return hints, nil
}
