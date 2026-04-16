package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) FoldingRange(_ context.Context, params *lsp.FoldingRangeParams) ([]lsp.FoldingRange, error) {
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

	var ranges []lsp.FoldingRange
	walkTree(root, func(n *gotreesitter.Node) bool {
		kind := foldingKind(n, lang)
		if kind == nil {
			return true
		}
		start := byteOffsetToPosition(content, n.StartByte())
		end := byteOffsetToPosition(content, n.EndByte())
		if end.Line <= start.Line {
			return true
		}
		ranges = append(ranges, lsp.FoldingRange{
			StartLine: start.Line,
			EndLine:   end.Line,
			Kind:      kind,
		})
		return true
	})
	return ranges, nil
}

func foldingKind(n *gotreesitter.Node, lang *gotreesitter.Language) *lsp.FoldingRangeKind {
	switch n.Type(lang) {
	case "ui_object_initializer":
		kind := lsp.FoldingRangeKindRegion
		return &kind
	case "statement_block":
		kind := lsp.FoldingRangeKindRegion
		return &kind
	case "comment":
		kind := lsp.FoldingRangeKindComment
		return &kind
	case "ui_import":
		// Group consecutive imports — handled by the node itself if multi-line,
		// but single imports are one line and don't fold. We rely on the parent
		// program node containing a run of imports; the editor collapses them
		// via the "imports" kind. Return nil here; we handle import groups below.
		return nil
	}
	return nil
}
