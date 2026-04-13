package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) References(_ context.Context, params *lsp.ReferenceParams) ([]lsp.Location, error) {
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

	node := h.parser.GetNodeAt(params.TextDocument.URI, params.Position, content)
	if node == nil {
		return nil, nil
	}

	nodeType := node.Type(lang)
	if nodeType != "identifier" {
		return nil, nil
	}

	identifierText := string(content[node.StartByte():node.EndByte()])

	var locations []lsp.Location
	findAllReferences(root, lang, content, identifierText, &locations)

	return locations, nil
}

func findAllReferences(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, target string, locations *[]lsp.Location) {
	if node == nil {
		return
	}

	if node.Type(lang) == "identifier" {
		nodeText := string(content[node.StartByte():node.EndByte()])
		if nodeText == target {
			*locations = append(*locations, lsp.Location{
				URI: "",
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(node.StartByte())},
					End:   lsp.Position{Line: 0, Character: int(node.EndByte())},
				},
			})
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			findAllReferences(child, lang, content, target, locations)
		}
	}
}
