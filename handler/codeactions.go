package handler

import (
	"context"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) CodeAction(_ context.Context, params *lsp.CodeActionParams) ([]lsp.CodeAction, error) {
	doc, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return nil, nil
	}

	var actions []lsp.CodeAction

	if h.parser != nil {
		tree := h.parser.GetTree(params.TextDocument.URI)
		if tree != nil {
			root := tree.RootNode()
			if root != nil {
				actions = collectCodeActions(root, h.parser.Language(), []byte(doc), params.Range)
			}
		}
	}

	return actions, nil
}

func collectCodeActions(node *gotreesitter.Node, lang *gotreesitter.Language, content []byte, selectionRange lsp.Range) []lsp.CodeAction {
	var actions []lsp.CodeAction

	if node == nil {
		return actions
	}

	nodeType := node.Type(lang)

	switch nodeType {
	case "ui_object_definition":
		actions = append(actions, createQuickFixAction(
			"Add id property",
			"Adds a unique id property to this component",
			node,
			"id: ",
		))

	case "string", "string_fragment":
		if parent := node.Parent(); parent != nil && parent.Type(lang) == "ui_binding" {
			valueText := string(content[node.StartByte():node.EndByte()])
			if isColorValue(valueText) && !isQuotedString(valueText) {
				actions = append(actions, createQuickFixAction(
					"Quote as color string",
					"Wrap color value in quotes",
					node,
					"\""+valueText+"\"",
				))
			}
		}
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			actions = append(actions, collectCodeActions(child, lang, content, selectionRange)...)
		}
	}

	return actions
}

func createQuickFixAction(title, description string, node *gotreesitter.Node, insertText string) lsp.CodeAction {
	kind := lsp.CodeActionQuickFix
	return lsp.CodeAction{
		Title: title,
		Kind:  &kind,
		Edit: &lsp.WorkspaceEdit{
			Changes: map[lsp.DocumentURI][]lsp.TextEdit{
				"": {
					{
						Range: lsp.Range{
							Start: lsp.Position{Line: 0, Character: int(node.StartByte())},
							End:   lsp.Position{Line: 0, Character: int(node.StartByte())},
						},
						NewText: insertText,
					},
				},
			},
		},
		Command:     nil,
		Diagnostics: []lsp.Diagnostic{},
	}
}

func isColorValue(s string) bool {
	colorKeywords := map[string]bool{
		"red": true, "green": true, "blue": true,
		"white": true, "black": true, "yellow": true,
		"cyan": true, "magenta": true, "gray": true,
		"grey": true, "transparent": true,
	}
	return colorKeywords[s]
}

func isQuotedString(s string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"'
}
