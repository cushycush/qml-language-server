package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) References(_ context.Context, params *lsp.ReferenceParams) ([]lsp.Location, error) {
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

	node := h.parser.GetNodeAt(uri, params.Position, content)
	if node == nil || node.Type(lang) != "identifier" {
		return nil, nil
	}

	target := string(content[node.StartByte():node.EndByte()])
	var locations []lsp.Location
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) == "identifier" && string(content[n.StartByte():n.EndByte()]) == target {
			locations = append(locations, nodeLocation(uri, content, n))
		}
		return true
	})
	return locations, nil
}
