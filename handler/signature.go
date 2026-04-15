package handler

import (
	"context"

	"github.com/owenrumney/go-lsp/lsp"
)

// functionSignatures are resolved from the label the user just typed. These
// are a tiny hand-rolled set; fuller signature help requires reading the
// qmltypes files Qt ships.
var functionSignatures = map[string]lsp.SignatureInformation{
	"Qt.rect": {
		Label: "Qt.rect(x: real, y: real, width: real, height: real): rect",
		Parameters: []lsp.ParameterInformation{
			{Label: "x: real", Documentation: plainText("X coordinate")},
			{Label: "y: real", Documentation: plainText("Y coordinate")},
			{Label: "width: real", Documentation: plainText("Width")},
			{Label: "height: real", Documentation: plainText("Height")},
		},
	},
	"Qt.size": {
		Label: "Qt.size(width: real, height: real): size",
		Parameters: []lsp.ParameterInformation{
			{Label: "width: real", Documentation: plainText("Width")},
			{Label: "height: real", Documentation: plainText("Height")},
		},
	},
	"Qt.point": {
		Label: "Qt.point(x: real, y: real): point",
		Parameters: []lsp.ParameterInformation{
			{Label: "x: real", Documentation: plainText("X coordinate")},
			{Label: "y: real", Documentation: plainText("Y coordinate")},
		},
	},
	"Qt.rgba": {
		Label: "Qt.rgba(r: real, g: real, b: real, a?: real): color",
		Parameters: []lsp.ParameterInformation{
			{Label: "r: real", Documentation: plainText("Red 0.0–1.0")},
			{Label: "g: real", Documentation: plainText("Green 0.0–1.0")},
			{Label: "b: real", Documentation: plainText("Blue 0.0–1.0")},
			{Label: "a?: real", Documentation: plainText("Alpha 0.0–1.0 (default 1.0)")},
		},
	},
	"console.log":   {Label: "console.log(...args: any): void", Parameters: []lsp.ParameterInformation{{Label: "...args: any", Documentation: plainText("Values to print")}}},
	"console.warn":  {Label: "console.warn(...args: any): void", Parameters: []lsp.ParameterInformation{{Label: "...args: any", Documentation: plainText("Values to print at warn level")}}},
	"console.error": {Label: "console.error(...args: any): void", Parameters: []lsp.ParameterInformation{{Label: "...args: any", Documentation: plainText("Values to print at error level")}}},
	"console.debug": {Label: "console.debug(...args: any): void", Parameters: []lsp.ParameterInformation{{Label: "...args: any", Documentation: plainText("Values to print at debug level")}}},
	"String":        {Label: "String(value: any): string", Parameters: []lsp.ParameterInformation{{Label: "value: any", Documentation: plainText("Value to coerce")}}},
	"Number":        {Label: "Number(value: any): number", Parameters: []lsp.ParameterInformation{{Label: "value: any", Documentation: plainText("Value to coerce")}}},
	"Boolean":       {Label: "Boolean(value: any): boolean", Parameters: []lsp.ParameterInformation{{Label: "value: any", Documentation: plainText("Value to coerce")}}},
	"parseInt":      {Label: "parseInt(s: string, radix?: int): int", Parameters: []lsp.ParameterInformation{{Label: "s: string", Documentation: plainText("String to parse")}, {Label: "radix?: int", Documentation: plainText("Base 2–36")}}},
	"parseFloat":    {Label: "parseFloat(s: string): number", Parameters: []lsp.ParameterInformation{{Label: "s: string", Documentation: plainText("String to parse")}}},
}

func (h *Handler) SignatureHelp(_ context.Context, params *lsp.SignatureHelpParams) (*lsp.SignatureHelp, error) {
	doc, ok := h.getDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	lines := getLines(doc)
	if int(params.Position.Line) >= len(lines) {
		return nil, nil
	}
	line := lines[int(params.Position.Line)]
	cursor := int(params.Position.Character)
	if cursor > len(line) {
		cursor = len(line)
	}

	callName, activeArg := findActiveCall(line, cursor)
	if callName == "" {
		return nil, nil
	}
	sig, ok := functionSignatures[callName]
	if !ok {
		return nil, nil
	}

	active := activeArg
	if active >= len(sig.Parameters) {
		active = len(sig.Parameters) - 1
	}
	if active < 0 {
		active = 0
	}
	activeSig := 0
	return &lsp.SignatureHelp{
		Signatures:      []lsp.SignatureInformation{sig},
		ActiveSignature: &activeSig,
		ActiveParameter: &active,
	}, nil
}

// findActiveCall walks back from cursor to find the open paren of the
// enclosing call, returning the callee name and the zero-based index of the
// parameter the cursor is currently on.
func findActiveCall(line string, cursor int) (string, int) {
	depth := 0
	commas := 0
	inString := byte(0)
	escape := false

	for i := cursor - 1; i >= 0; i-- {
		c := line[i]
		if escape {
			escape = false
			continue
		}
		if inString != 0 {
			switch c {
			case '\\':
				escape = true
			case inString:
				inString = 0
			}
			continue
		}
		switch c {
		case '"', '\'', '`':
			inString = c
		case ')', ']', '}':
			depth++
		case '(':
			if depth == 0 {
				return extractCallee(line, i), commas
			}
			depth--
		case '[', '{':
			if depth == 0 {
				return "", 0
			}
			depth--
		case ',':
			if depth == 0 {
				commas++
			}
		}
	}
	return "", 0
}

func extractCallee(line string, parenIdx int) string {
	end := parenIdx
	start := end
	for start > 0 {
		c := line[start-1]
		if isIdentChar(c) || c == '.' {
			start--
			continue
		}
		break
	}
	if start == end {
		return ""
	}
	return line[start:end]
}

func plainText(s string) *lsp.MarkupContent {
	return &lsp.MarkupContent{Kind: lsp.PlainText, Value: s}
}

func countParams(text string) int {
	count := 0
	inString := byte(0)
	escape := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		if escape {
			escape = false
			continue
		}
		if inString != 0 {
			switch c {
			case '\\':
				escape = true
			case inString:
				inString = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			inString = c
		case ',':
			count++
		}
	}
	return count
}

