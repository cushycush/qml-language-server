package handler

import (
	"context"
	"encoding/json"
	"strings"

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
		if n.Type(lang) != "call_expression" {
			return true
		}
		hints = append(hints, callArgumentHints(n, lang, content)...)
		return false
	})
	return hints, nil
}

// callArgumentHints emits parameter-name hints for arguments in a call
// expression, using the known function signatures from the signature-help
// registry.
func callArgumentHints(call *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []lsp.InlayHint {
	callee := calleeText(call, lang, content)
	if callee == "" {
		return nil
	}
	sig, ok := functionSignatures[callee]
	if !ok || len(sig.Parameters) == 0 {
		return nil
	}

	args := findArguments(call, lang)
	if len(args) == 0 {
		return nil
	}

	var hints []lsp.InlayHint
	for i, arg := range args {
		if i >= len(sig.Parameters) {
			break
		}
		labelStr, ok := sig.Parameters[i].Label.(string)
		if !ok {
			continue
		}
		label := paramName(labelStr)
		if label == "" {
			continue
		}
		paddingRight := true
		kind := lsp.InlayHintKind(lsp.InlayHintKindParameter)
		hints = append(hints, lsp.InlayHint{
			Position:     byteOffsetToPosition(content, arg.StartByte()),
			Label:        json.RawMessage(`"` + label + `:"`),
			Kind:         &kind,
			PaddingRight: &paddingRight,
		})
	}
	return hints
}

func calleeText(call *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < call.ChildCount(); i++ {
		child := call.Child(i)
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "identifier", "member_expression", "nested_identifier":
			return string(content[child.StartByte():child.EndByte()])
		}
	}
	return ""
}

func findArguments(call *gotreesitter.Node, lang *gotreesitter.Language) []*gotreesitter.Node {
	for i := 0; i < call.ChildCount(); i++ {
		child := call.Child(i)
		if child == nil || child.Type(lang) != "arguments" {
			continue
		}
		var args []*gotreesitter.Node
		for j := 0; j < child.ChildCount(); j++ {
			arg := child.Child(j)
			if arg == nil {
				continue
			}
			t := arg.Type(lang)
			if t == "(" || t == ")" || t == "," {
				continue
			}
			args = append(args, arg)
		}
		return args
	}
	return nil
}

// paramName extracts the bare name from a parameter label like "x: real" → "x".
func paramName(label string) string {
	name, _, _ := strings.Cut(label, ":")
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "...")
	return name
}
