package qmlgrammars

import (
	"unicode"

	"github.com/odvcencio/gotreesitter"
)

const (
	qmlTokAutoSemicolon   = 0
	qmlTokTemplateChars   = 1
	qmlTokTernaryQmark    = 2
	qmlTokHtmlComment     = 3
	qmlTokLogicalOr       = 4
	qmlTokEscapeSequence  = 5
	qmlTokRegexPattern    = 6
	qmlTokJsxText         = 7
	qmlTokFuncSigAutoSemi = 8
	qmlTokErrorRecovery   = 9
)

const (
	qmlSymAutoSemicolon   gotreesitter.Symbol = 169
	qmlSymTemplateChars   gotreesitter.Symbol = 170
	qmlSymTernaryQmark    gotreesitter.Symbol = 171
	qmlSymHtmlComment     gotreesitter.Symbol = 172
	qmlSymJsxText         gotreesitter.Symbol = 173
	qmlSymFuncSigAutoSemi gotreesitter.Symbol = 174
)

type QmljsExternalScanner struct{}

func newQmljsScanner() *QmljsExternalScanner {
	return &QmljsExternalScanner{}
}

func (QmljsExternalScanner) Create() any                           { return nil }
func (QmljsExternalScanner) Destroy(payload any)                   {}
func (QmljsExternalScanner) Serialize(payload any, buf []byte) int { return 0 }
func (QmljsExternalScanner) Deserialize(payload any, buf []byte)   {}
func (QmljsExternalScanner) SupportsIncrementalReuse() bool        { return true }

func (QmljsExternalScanner) Scan(payload any, lexer *gotreesitter.ExternalLexer, validSymbols []bool) bool {
	if qmlValid(validSymbols, qmlTokTemplateChars) {
		if qmlValid(validSymbols, qmlTokAutoSemicolon) {
			return false
		}
		return qmlScanTemplateChars(lexer)
	}

	preferAutoSemicolon := qmlPreferAutoSemicolonOverJsxText(lexer, validSymbols)

	if qmlValid(validSymbols, qmlTokJsxText) && !preferAutoSemicolon {
		if qmlScanJsxText(lexer) {
			return true
		}
	}

	if qmlValid(validSymbols, qmlTokAutoSemicolon) || qmlValid(validSymbols, qmlTokFuncSigAutoSemi) {
		scannedComment := false
		ret := qmlScanAutoSemicolon(lexer, validSymbols, &scannedComment)
		if !ret && !scannedComment && qmlValid(validSymbols, qmlTokTernaryQmark) && lexer.Lookahead() == '?' {
			return qmlScanTernaryQmark(lexer)
		}
		if !ret && !scannedComment && preferAutoSemicolon && qmlValid(validSymbols, qmlTokJsxText) {
			return qmlScanJsxText(lexer)
		}
		return ret
	}

	if qmlValid(validSymbols, qmlTokJsxText) && preferAutoSemicolon {
		return qmlScanJsxText(lexer)
	}

	if qmlValid(validSymbols, qmlTokTernaryQmark) {
		return qmlScanTernaryQmark(lexer)
	}

	if qmlValid(validSymbols, qmlTokHtmlComment) &&
		!qmlValid(validSymbols, qmlTokLogicalOr) &&
		!qmlValid(validSymbols, qmlTokEscapeSequence) &&
		!qmlValid(validSymbols, qmlTokRegexPattern) {
		return qmlScanClosingComment(lexer)
	}

	return false
}

func qmlScanTemplateChars(lexer *gotreesitter.ExternalLexer) bool {
	lexer.SetResultSymbol(qmlSymTemplateChars)
	hasContent := false
	for {
		lexer.MarkEnd()
		switch lexer.Lookahead() {
		case '`':
			return hasContent
		case 0:
			return false
		case '$':
			lexer.Advance(false)
			if lexer.Lookahead() == '{' {
				return hasContent
			}
		case '\\':
			return hasContent
		default:
			lexer.Advance(false)
			hasContent = true
		}
	}
}

func qmlScanAutoSemicolon(lexer *gotreesitter.ExternalLexer, validSymbols []bool, scannedComment *bool) bool {
	lexer.SetResultSymbol(qmlSymAutoSemicolon)
	lexer.MarkEnd()

	for {
		ch := lexer.Lookahead()
		if ch == 0 {
			return true
		}
		if ch == '}' {
			lexer.Advance(true)
			for unicode.IsSpace(lexer.Lookahead()) {
				lexer.Advance(true)
			}
			switch lexer.Lookahead() {
			case ':':
				return qmlValid(validSymbols, qmlTokLogicalOr)
			default:
				if qmlValid(validSymbols, qmlTokJsxText) {
					return false
				}
				if qmlLooksLikeJSXAttributeContinuation(lexer) {
					return false
				}
			}
			switch lexer.Lookahead() {
			case '>':
				return false
			case '/':
				lexer.Advance(true)
				return lexer.Lookahead() != '>'
			case '<':
				lexer.Advance(true)
				return lexer.Lookahead() != '/'
			default:
				return true
			}
		}
		if !unicode.IsSpace(ch) {
			return false
		}
		if ch == '\n' {
			break
		}
		lexer.Advance(true)
	}

	lexer.Advance(true)

	if !qmlScanWSAndComments(lexer, scannedComment) {
		return false
	}

	switch lexer.Lookahead() {
	case '`', ',', '.', ';', '*', '%', '>', '<', '=', '?', '^', '|', '&', '/', ':':
		return false
	case '{':
		if qmlValid(validSymbols, qmlTokFuncSigAutoSemi) {
			return false
		}
	case '(', '[':
		if qmlValid(validSymbols, qmlTokLogicalOr) {
			return false
		}
	case '+':
		lexer.Advance(true)
		return lexer.Lookahead() == '+'
	case '-':
		lexer.Advance(true)
		return lexer.Lookahead() == '-'
	case '!':
		lexer.Advance(true)
		return lexer.Lookahead() != '='
	case 'i':
		lexer.Advance(true)
		if lexer.Lookahead() != 'n' {
			return true
		}
		lexer.Advance(true)
		if !isQmljsIdentifierChar(lexer.Lookahead()) {
			return false
		}
		instanceof := "instanceof"
		for i := 0; i < len(instanceof); i++ {
			if lexer.Lookahead() != rune(instanceof[i]) {
				return true
			}
			lexer.Advance(true)
		}
		if !isQmljsIdentifierChar(lexer.Lookahead()) {
			return false
		}
	}

	return true
}

func qmlScanWSAndComments(lexer *gotreesitter.ExternalLexer, scannedComment *bool) bool {
	for {
		for unicode.IsSpace(lexer.Lookahead()) {
			lexer.Advance(true)
		}
		if lexer.Lookahead() == '/' {
			lexer.Advance(true)
			if lexer.Lookahead() == '/' {
				lexer.Advance(true)
				for lexer.Lookahead() != 0 && lexer.Lookahead() != '\n' {
					lexer.Advance(true)
				}
				*scannedComment = true
			} else if lexer.Lookahead() == '*' {
				lexer.Advance(true)
				for lexer.Lookahead() != 0 {
					if lexer.Lookahead() == '*' {
						lexer.Advance(true)
						if lexer.Lookahead() == '/' {
							lexer.Advance(true)
							break
						}
					} else {
						lexer.Advance(true)
					}
				}
			} else {
				return false
			}
		} else {
			return true
		}
	}
}

func qmlScanTernaryQmark(lexer *gotreesitter.ExternalLexer) bool {
	for unicode.IsSpace(lexer.Lookahead()) {
		lexer.Advance(true)
	}

	if lexer.Lookahead() != '?' {
		return false
	}
	lexer.Advance(false)

	if lexer.Lookahead() == '?' || lexer.Lookahead() == '.' {
		return false
	}

	lexer.MarkEnd()
	lexer.SetResultSymbol(qmlSymTernaryQmark)

	for unicode.IsSpace(lexer.Lookahead()) {
		lexer.Advance(false)
	}

	if lexer.Lookahead() == ':' || lexer.Lookahead() == ')' || lexer.Lookahead() == ',' {
		return false
	}

	if lexer.Lookahead() == '.' {
		lexer.Advance(false)
		return unicode.IsDigit(lexer.Lookahead())
	}
	return true
}

func qmlScanClosingComment(lexer *gotreesitter.ExternalLexer) bool {
	for unicode.IsSpace(lexer.Lookahead()) || lexer.Lookahead() == 0x2028 || lexer.Lookahead() == 0x2029 {
		lexer.Advance(true)
	}

	commentStart := "<!--"
	commentEnd := "-->"

	if lexer.Lookahead() == '<' {
		for i := 0; i < len(commentStart); i++ {
			if lexer.Lookahead() != rune(commentStart[i]) {
				return false
			}
			lexer.Advance(false)
		}
	} else if lexer.Lookahead() == '-' {
		for i := 0; i < len(commentEnd); i++ {
			if lexer.Lookahead() != rune(commentEnd[i]) {
				return false
			}
			lexer.Advance(false)
		}
	} else {
		return false
	}

	for lexer.Lookahead() != 0 && lexer.Lookahead() != '\n' &&
		lexer.Lookahead() != 0x2028 && lexer.Lookahead() != 0x2029 {
		lexer.Advance(false)
	}

	lexer.SetResultSymbol(qmlSymHtmlComment)
	lexer.MarkEnd()
	return true
}

func qmlScanJsxText(lexer *gotreesitter.ExternalLexer) bool {
	sawText := false
	atNewline := false
	onlyWhitespace := true

	for lexer.Lookahead() != 0 && lexer.Lookahead() != '<' && lexer.Lookahead() != '>' &&
		lexer.Lookahead() != '{' && lexer.Lookahead() != '}' && lexer.Lookahead() != '&' {
		if lexer.Lookahead() == '/' && onlyWhitespace {
			lexer.Advance(false)
			if lexer.Lookahead() == '>' {
				return false
			}
			sawText = true
			onlyWhitespace = false
			continue
		}
		if onlyWhitespace && (lexer.Lookahead() == '_' || unicode.IsLetter(lexer.Lookahead())) {
			for {
				lexer.Advance(false)
				ch := lexer.Lookahead()
				if ch == '_' || ch == '-' || ch == ':' || ch == '.' ||
					unicode.IsLetter(ch) || unicode.IsDigit(ch) {
					continue
				}
				break
			}
			for unicode.IsSpace(lexer.Lookahead()) {
				lexer.Advance(false)
			}
			if lexer.Lookahead() == '=' {
				return false
			}
			sawText = true
			onlyWhitespace = false
			continue
		}
		isWS := unicode.IsSpace(lexer.Lookahead())
		if lexer.Lookahead() == '\n' {
			atNewline = true
		} else {
			atNewline = atNewline && isWS
			if !atNewline {
				sawText = true
			}
		}
		if !isWS {
			onlyWhitespace = false
		}
		lexer.Advance(false)
	}

	lexer.MarkEnd()
	lexer.SetResultSymbol(qmlSymJsxText)
	return sawText
}

func qmlValid(vs []bool, i int) bool { return i < len(vs) && vs[i] }

func qmlLooksLikeJSXAttributeContinuation(lexer *gotreesitter.ExternalLexer) bool {
	ch := lexer.Lookahead()
	if ch != '_' && !unicode.IsLetter(ch) {
		return false
	}
	for {
		lexer.Advance(true)
		ch = lexer.Lookahead()
		if ch == '_' || ch == '-' || ch == ':' || ch == '.' ||
			unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			continue
		}
		break
	}
	for unicode.IsSpace(ch) {
		lexer.Advance(true)
		ch = lexer.Lookahead()
	}
	return ch == '=' || ch == '/' || ch == '>'
}

func qmlPreferAutoSemicolonOverJsxText(lexer *gotreesitter.ExternalLexer, validSymbols []bool) bool {
	if !qmlValid(validSymbols, qmlTokAutoSemicolon) || !qmlValid(validSymbols, qmlTokJsxText) {
		return false
	}
	switch lexer.Lookahead() {
	case 0, '\n', '\r', 0x2028, 0x2029:
		return true
	default:
		return false
	}
}

func isQmljsIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$'
}
