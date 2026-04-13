package handler

import (
	"context"
	"encoding/json"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) InlayHint(_ context.Context, params *lsp.InlayHintParams) ([]lsp.InlayHint, error) {
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

	return collectInlayHints(root, lang, content), nil
}

func collectInlayHints(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.InlayHint {
	var hints []lsp.InlayHint

	if node == nil {
		return hints
	}

	nodeType := node.Type(lang)

	switch nodeType {
	case "ui_binding":
		propertyName := extractPropertyIdentifier(node, lang, content)
		if propertyName != "" {
			valueRange := getBindingValueRange(node, lang)
			if valueRange != nil {
				paddingLeft := true
				kind := lsp.InlayHintKind(lsp.InlayHintKindType)
				hints = append(hints, lsp.InlayHint{
					Position: lsp.Position{
						Line:      0,
						Character: int(node.StartByte()),
					},
					Label:        json.RawMessage("\"" + propertyName + ": \""),
					Kind:         &kind,
					PaddingLeft:  &paddingLeft,
					PaddingRight: nil,
				})
			}
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			hints = append(hints, collectInlayHints(child, lang, content)...)
		}
	}

	return hints
}

func extractPropertyIdentifier(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			childType := child.Type(lang)
			if childType == "identifier" || childType == "property_identifier" {
				return string(content[child.StartByte():child.EndByte()])
			}
			if childType == "nested_identifier" {
				return string(content[child.StartByte():child.EndByte()])
			}
		}
	}
	return ""
}

func getBindingValueRange(node *gotreesitter.Node, lang *gotreesitter.Language) *lsp.Range {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			if child.Type(lang) == ":" || child.Type(lang) == "::" {
				return &lsp.Range{
					Start: lsp.Position{Line: 0, Character: int(child.EndByte())},
					End:   lsp.Position{Line: 0, Character: int(node.EndByte())},
				}
			}
		}
	}
	return nil
}
