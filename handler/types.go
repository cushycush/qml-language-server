package handler

import "github.com/owenrumney/go-lsp/lsp"

// QMLTypeInfo and PropertyInfo are thin views over the symbol registry kept
// for readability at hover/completion call sites. The registry in registry.go
// is the canonical store; these structs only describe what callers look at.

type QMLTypeInfo struct {
	Description string
	Type        string
	Module      string
}

type PropertyInfo struct {
	Description string
	Type        string
}

func getTypeInfo(name string) (QMLTypeInfo, bool)      { return registryTypeInfo(name) }
func getPropertyInfo(name string) (PropertyInfo, bool) { return registryPropertyInfo(name) }
func getCompletionTypes() []lsp.CompletionItem         { return completionItemsByCategory("type") }

// getLines splits text on '\n'. An empty trailing line is preserved so the
// slice length tracks the 1-based line count.
func getLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(text); i++ {
		if i == len(text) || text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	return lines
}

func extractWordAt(text string, pos int) string {
	if pos < 0 || pos > len(text) {
		return ""
	}
	start, end := pos, pos
	for start > 0 && isIdentChar(text[start-1]) {
		start--
	}
	for end < len(text) && isIdentChar(text[end]) {
		end++
	}
	return text[start:end]
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
