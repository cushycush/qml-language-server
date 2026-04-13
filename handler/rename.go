package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) Rename(_ context.Context, params *lsp.RenameParams) (*lsp.WorkspaceEdit, error) {
	doc, ok := h.documents[params.TextDocument.URI]
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

	lang := h.parser.Language()
	content := []byte(doc)

	node := h.parser.GetNodeAt(params.TextDocument.URI, params.Position)
	if node == nil || node.Type(lang) != "identifier" {
		return nil, nil
	}

	identifierText := string(content[node.StartByte():node.EndByte()])

	var textEdits []lsp.TextEdit
	collectRenameEdits(root, lang, content, identifierText, &textEdits)

	return &lsp.WorkspaceEdit{
		Changes: map[lsp.DocumentURI][]lsp.TextEdit{
			params.TextDocument.URI: textEdits,
		},
	}, nil
}

func collectRenameEdits(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, target string, edits *[]lsp.TextEdit) {
	if node == nil {
		return
	}

	if node.Type(lang) == "identifier" {
		nodeText := string(content[node.StartByte():node.EndByte()])
		if nodeText == target {
			*edits = append(*edits, lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(node.StartByte())},
					End:   lsp.Position{Line: 0, Character: int(node.EndByte())},
				},
				NewText: target + "_renamed",
			})
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			collectRenameEdits(child, lang, content, target, edits)
		}
	}
}

func (h *Handler) PrepareRename(_ context.Context, params *lsp.PrepareRenameParams) (*lsp.PrepareRenameResult, error) {
	if h.parser == nil {
		return nil, nil
	}

	node := h.parser.GetNodeAt(params.TextDocument.URI, params.Position)
	if node == nil || node.Type(h.parser.Language()) != "identifier" {
		return nil, nil
	}

	content := []byte(h.documents[params.TextDocument.URI])
	placeholder := string(content[node.StartByte():node.EndByte()])

	return &lsp.PrepareRenameResult{
		Range: lsp.Range{
			Start: lsp.Position{Line: int(params.Position.Line), Character: int(node.StartByte())},
			End:   lsp.Position{Line: int(params.Position.Line), Character: int(node.EndByte())},
		},
		Placeholder: placeholder,
	}, nil
}
