package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Rename(_ context.Context, params *lsp.RenameParams) (*lsp.WorkspaceEdit, error) {
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

	var edits []lsp.TextEdit
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) == "identifier" && string(content[n.StartByte():n.EndByte()]) == target {
			edits = append(edits, lsp.TextEdit{
				Range:   nodeRange(content, n),
				NewText: params.NewName,
			})
		}
		return true
	})

	return &lsp.WorkspaceEdit{
		Changes: map[lsp.DocumentURI][]lsp.TextEdit{uri: edits},
	}, nil
}

func (h *Handler) PrepareRename(_ context.Context, params *lsp.PrepareRenameParams) (*lsp.PrepareRenameResult, error) {
	uri := params.TextDocument.URI
	doc, ok := h.getDocument(uri)
	if !ok || h.parser == nil {
		return nil, nil
	}

	content := []byte(doc)
	node := h.parser.GetNodeAt(uri, params.Position, content)
	if node == nil || node.Type(h.parser.Language()) != "identifier" {
		return nil, nil
	}

	return &lsp.PrepareRenameResult{
		Range:       nodeRange(content, node),
		Placeholder: string(content[node.StartByte():node.EndByte()]),
	}, nil
}
