package handler

import (
	"context"
	"sort"
	"strings"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

// Semantic token types we publish in the legend. Index in this slice == the
// token-type integer the LSP client expects.
var semanticTokenTypes = []string{
	"namespace",
	"type",
	"class",
	"enum",
	"interface",
	"struct",
	"typeParameter",
	"parameter",
	"variable",
	"property",
	"enumMember",
	"event",
	"function",
	"method",
	"macro",
	"keyword",
	"modifier",
	"comment",
	"string",
	"number",
	"regexp",
	"operator",
}

var semanticTokenModifiers = []string{
	"declaration",
	"definition",
	"readonly",
	"static",
	"deprecated",
	"abstract",
	"async",
	"modification",
	"documentation",
	"defaultLibrary",
}

const (
	tokTypeNamespace = iota
	tokTypeType
	tokTypeClass
	tokTypeEnum
	tokTypeInterface
	tokTypeStruct
	tokTypeTypeParameter
	tokTypeParameter
	tokTypeVariable
	tokTypeProperty
	tokTypeEnumMember
	tokTypeEvent
	tokTypeFunction
	tokTypeMethod
	tokTypeMacro
	tokTypeKeyword
	tokTypeModifier
	tokTypeComment
	tokTypeString
	tokTypeNumber
	tokTypeRegexp
	tokTypeOperator
)

const (
	tokModDeclaration = 1 << iota
	tokModDefinition
	tokModReadonly
	tokModStatic
	tokModDeprecated
	tokModAbstract
	tokModAsync
	tokModModification
	tokModDocumentation
	tokModDefaultLibrary
)

// SemanticTokensLegend is published in the server capabilities so clients know
// how to interpret token-type and modifier indices we emit.
func SemanticTokensLegend() lsp.SemanticTokensLegend {
	return lsp.SemanticTokensLegend{
		TokenTypes:     append([]string(nil), semanticTokenTypes...),
		TokenModifiers: append([]string(nil), semanticTokenModifiers...),
	}
}

// rawToken is one un-encoded semantic token. The wire format LSP expects is
// delta-encoded; we collect tokens absolute, sort them, then encode.
type rawToken struct {
	Line, Char int
	Length     int
	TokenType  int
	Modifiers  int
}

func (h *Handler) SemanticTokensFull(_ context.Context, params *lsp.SemanticTokensParams) (*lsp.SemanticTokens, error) {
	doc, ok := h.getDocument(params.TextDocument.URI)
	if !ok || h.parser == nil {
		return &lsp.SemanticTokens{Data: []int{}}, nil
	}
	tree := h.parser.GetTree(params.TextDocument.URI)
	if tree == nil {
		return &lsp.SemanticTokens{Data: []int{}}, nil
	}
	root := tree.RootNode()
	if root == nil {
		return &lsp.SemanticTokens{Data: []int{}}, nil
	}

	tokens := collectSemanticTokens(root, h.parser.Language(), []byte(doc))
	return &lsp.SemanticTokens{Data: encodeSemanticTokens(tokens)}, nil
}

// collectSemanticTokens walks the tree once and emits a token for each node
// whose type maps to a known semantic-token category. For composite nodes
// (ui_object_definition, ui_import, ui_property, ui_binding) we emit tokens
// for the meaningful child rather than the whole span.
func collectSemanticTokens(root *gotreesitter.Node, lang *gotreesitter.Language, content []byte) []rawToken {
	var tokens []rawToken
	walkTree(root, func(n *gotreesitter.Node) bool {
		switch n.Type(lang) {
		case "ui_import":
			emitImport(n, lang, content, &tokens)
		case "ui_object_definition":
			emitObjectType(n, lang, content, &tokens)
		case "ui_object_binding":
			emitObjectType(n, lang, content, &tokens)
		case "ui_property":
			emitPropertyDecl(n, lang, content, &tokens)
		case "ui_required":
			emitRequired(n, lang, content, &tokens)
		case "ui_binding":
			emitBindingName(n, lang, content, &tokens)
		case "ui_signal":
			emitSignal(n, lang, content, &tokens)
		case "comment":
			pushTokenForNode(n, content, &tokens, tokTypeComment, 0)
			return false
		case "string", "template_string":
			pushTokenForNode(n, content, &tokens, tokTypeString, 0)
			return false
		case "number":
			pushTokenForNode(n, content, &tokens, tokTypeNumber, 0)
		case "regex":
			pushTokenForNode(n, content, &tokens, tokTypeRegexp, 0)
		case "true", "false", "null", "undefined":
			pushTokenForNode(n, content, &tokens, tokTypeKeyword, 0)
		}
		return true
	})
	return tokens
}

func emitImport(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, tokens *[]rawToken) {
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}
		ctype := child.Type(lang)
		switch ctype {
		case "import":
			pushTokenForNode(child, content, tokens, tokTypeKeyword, 0)
		case "identifier", "nested_identifier":
			pushTokenForNode(child, content, tokens, tokTypeNamespace, 0)
		case "string":
			pushTokenForNode(child, content, tokens, tokTypeString, 0)
		}
	}
}

func emitObjectType(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, tokens *[]rawToken) {
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}
		ctype := child.Type(lang)
		if ctype == "identifier" || ctype == "nested_identifier" {
			pushTokenForNode(child, content, tokens, tokTypeType, tokModDefaultLibrary)
			return
		}
	}
}

func emitPropertyDecl(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, tokens *[]rawToken) {
	// Pattern: [property|readonly property|default property] <type> <name> [: value]
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "property", "readonly", "default", "required":
			pushTokenForNode(child, content, tokens, tokTypeKeyword, 0)
		case "ui_property_type", "type_identifier":
			pushTokenForNode(child, content, tokens, tokTypeType, 0)
		case "identifier":
			pushTokenForNode(child, content, tokens, tokTypeProperty, tokModDeclaration)
		}
	}
}

func emitRequired(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, tokens *[]rawToken) {
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "required":
			pushTokenForNode(child, content, tokens, tokTypeKeyword, 0)
		case "identifier":
			pushTokenForNode(child, content, tokens, tokTypeProperty, tokModDeclaration)
		}
	}
}

func emitBindingName(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, tokens *[]rawToken) {
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}
		ctype := child.Type(lang)
		if ctype != "identifier" && ctype != "nested_identifier" {
			continue
		}
		name := string(content[child.StartByte():child.EndByte()])
		tokenType := tokTypeProperty
		if isSignalHandler(strings.TrimSpace(name)) {
			tokenType = tokTypeEvent
		}
		pushTokenForNode(child, content, tokens, tokenType, 0)
		return
	}
}

func emitSignal(n *gotreesitter.Node, lang *gotreesitter.Language, content []byte, tokens *[]rawToken) {
	for i := 0; i < n.ChildCount(); i++ {
		child := n.Child(i)
		if child == nil {
			continue
		}
		switch child.Type(lang) {
		case "signal":
			pushTokenForNode(child, content, tokens, tokTypeKeyword, 0)
		case "identifier":
			pushTokenForNode(child, content, tokens, tokTypeEvent, tokModDeclaration)
			return
		}
	}
}

// pushTokenForNode appends a token if the node spans a single line and has
// non-zero length. Multi-line tokens have to be split per LSP, and we just
// drop them for now (only applies to multi-line strings/comments).
func pushTokenForNode(n *gotreesitter.Node, content []byte, tokens *[]rawToken, tokenType, modifiers int) {
	start := byteOffsetToPosition(content, n.StartByte())
	end := byteOffsetToPosition(content, n.EndByte())
	if start.Line != end.Line {
		// Split a multi-line token into one per line so clients render every
		// line of a multi-line string/comment.
		startByte := n.StartByte()
		endByte := n.EndByte()
		lineStart := startByte
		for off := startByte; off < endByte; off++ {
			if content[off] == '\n' {
				if off > lineStart {
					p := byteOffsetToPosition(content, lineStart)
					*tokens = append(*tokens, rawToken{
						Line:      p.Line,
						Char:      p.Character,
						Length:    int(off - lineStart),
						TokenType: tokenType,
						Modifiers: modifiers,
					})
				}
				lineStart = off + 1
			}
		}
		if endByte > lineStart {
			p := byteOffsetToPosition(content, lineStart)
			*tokens = append(*tokens, rawToken{
				Line:      p.Line,
				Char:      p.Character,
				Length:    int(endByte - lineStart),
				TokenType: tokenType,
				Modifiers: modifiers,
			})
		}
		return
	}
	length := end.Character - start.Character
	if length <= 0 {
		return
	}
	*tokens = append(*tokens, rawToken{
		Line:      start.Line,
		Char:      start.Character,
		Length:    length,
		TokenType: tokenType,
		Modifiers: modifiers,
	})
}

// encodeSemanticTokens turns absolute tokens into the LSP delta-encoded uint
// stream: [deltaLine, deltaStart, length, type, modifiers] per token.
func encodeSemanticTokens(tokens []rawToken) []int {
	if len(tokens) == 0 {
		return []int{}
	}
	sort.SliceStable(tokens, func(i, j int) bool {
		if tokens[i].Line != tokens[j].Line {
			return tokens[i].Line < tokens[j].Line
		}
		return tokens[i].Char < tokens[j].Char
	})

	data := make([]int, 0, len(tokens)*5)
	prevLine, prevChar := 0, 0
	for _, t := range tokens {
		deltaLine := t.Line - prevLine
		deltaChar := t.Char
		if deltaLine == 0 {
			deltaChar = t.Char - prevChar
		}
		data = append(data, deltaLine, deltaChar, t.Length, t.TokenType, t.Modifiers)
		prevLine = t.Line
		prevChar = t.Char
	}
	return data
}
