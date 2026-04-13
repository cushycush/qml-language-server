package handler

import (
	"context"

	"github.com/owenrumney/go-lsp/lsp"
)

func (h *Handler) SignatureHelp(_ context.Context, params *lsp.SignatureHelpParams) (*lsp.SignatureHelp, error) {
	doc, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return nil, nil
	}

	pos := params.Position
	lines := getLines(doc)
	if int(pos.Line) >= len(lines) {
		return nil, nil
	}

	lineText := lines[int(pos.Line)]
	beforeCursor := lineText[:int(pos.Character)]

	sig := detectFunctionCall(beforeCursor)
	if sig == nil {
		return nil, nil
	}

	return sig, nil
}

func detectFunctionCall(text string) *lsp.SignatureHelp {
	functionSignatures := map[string]lsp.SignatureInformation{
		"Qt.rect": {
			Label: "Qt.rect(x: real, y: real, width: real, height: real): rect",
			Parameters: []lsp.ParameterInformation{
				{Label: "x: real", Documentation: stringPtr("X coordinate")},
				{Label: "y: real", Documentation: stringPtr("Y coordinate")},
				{Label: "width: real", Documentation: stringPtr("Width of rectangle")},
				{Label: "height: real", Documentation: stringPtr("Height of rectangle")},
			},
		},
		"Qt.size": {
			Label: "Qt.size(width: real, height: real): size",
			Parameters: []lsp.ParameterInformation{
				{Label: "width: real", Documentation: stringPtr("Width")},
				{Label: "height: real", Documentation: stringPtr("Height")},
			},
		},
		"Qt.point": {
			Label: "Qt.point(x: real, y: real): point",
			Parameters: []lsp.ParameterInformation{
				{Label: "x: real", Documentation: stringPtr("X coordinate")},
				{Label: "y: real", Documentation: stringPtr("Y coordinate")},
			},
		},
		"console.log": {
			Label: "console.log(message: string): void",
			Parameters: []lsp.ParameterInformation{
				{Label: "message: string", Documentation: stringPtr("The message to log")},
			},
		},
		"console.warn": {
			Label: "console.warn(message: string): void",
			Parameters: []lsp.ParameterInformation{
				{Label: "message: string", Documentation: stringPtr("The warning message")},
			},
		},
		"console.error": {
			Label: "console.error(message: string): void",
			Parameters: []lsp.ParameterInformation{
				{Label: "message: string", Documentation: stringPtr("The error message")},
			},
		},
		"String": {
			Label: "String(value: any): string",
			Parameters: []lsp.ParameterInformation{
				{Label: "value: any", Documentation: stringPtr("Value to convert to string")},
			},
		},
		"Number": {
			Label: "Number(value: any): number",
			Parameters: []lsp.ParameterInformation{
				{Label: "value: any", Documentation: stringPtr("Value to convert to number")},
			},
		},
		"Boolean": {
			Label: "Boolean(value: any): boolean",
			Parameters: []lsp.ParameterInformation{
				{Label: "value: any", Documentation: stringPtr("Value to convert to boolean")},
			},
		},
	}

	for funcName, sig := range functionSignatures {
		if len(text) > len(funcName) && text[len(text)-len(funcName):] == funcName {
			paramCount := countParams(text)
			activeParam := paramCount
			if activeParam >= len(sig.Parameters) {
				activeParam = len(sig.Parameters) - 1
			}
			activeSig := 0
			activePar := activeParam
			return &lsp.SignatureHelp{
				Signatures:      []lsp.SignatureInformation{sig},
				ActiveSignature: &activeSig,
				ActiveParameter: &activePar,
			}
		}
	}

	return nil
}

func countParams(text string) int {
	count := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(text); i++ {
		c := text[i]

		if !inString && (c == '"' || c == '\'') {
			inString = true
			stringChar = c
		} else if inString && c == stringChar && (i == 0 || text[i-1] != '\\') {
			inString = false
		} else if !inString && c == ',' {
			count++
		}
	}

	return count
}

func stringPtr(s string) *lsp.MarkupContent {
	return &lsp.MarkupContent{
		Kind:  lsp.PlainText,
		Value: s,
	}
}
