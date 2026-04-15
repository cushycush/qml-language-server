package handler

import (
	"context"
	"strings"

	"github.com/owenrumney/go-lsp/lsp"
)

// Formatting takes the current document text, re-indents it based on brace
// depth, trims trailing whitespace from every line, collapses runs of blank
// lines to at most one, and ensures the file ends with a single newline.
// The whole pass is whitespace-only — token content is never modified — so
// even a syntactically broken document gets back something valid for the
// non-broken parts.
func (h *Handler) Formatting(_ context.Context, params *lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	doc, ok := h.getDocument(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	formatted := formatQML(doc, params.Options)
	if formatted == doc {
		return []lsp.TextEdit{}, nil
	}
	content := []byte(doc)
	return []lsp.TextEdit{{
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   byteOffsetToPosition(content, uint32(len(content))),
		},
		NewText: formatted,
	}}, nil
}

// formatQML applies the indentation/whitespace rules. It walks the document
// character-by-character to track string/comment state so braces inside those
// don't move the indentation depth.
func formatQML(text string, opts lsp.FormattingOptions) string {
	indentUnit := indentUnitFrom(opts)
	lines := strings.Split(text, "\n")

	type lineInfo struct {
		text       string
		indent     int  // depth this line is rendered at
		blank      bool // true after trim-and-strip
	}
	infos := make([]lineInfo, 0, len(lines))

	depth := 0
	state := scanState{}
	for _, raw := range lines {
		stripped := strings.TrimSpace(raw)
		// Compute the leading-brace adjustment first so a line that *starts*
		// with `}` is rendered one level shallower than its content depth.
		leadingClose := countLeadingClose(stripped)
		renderDepth := max(depth-leadingClose, 0)
		infos = append(infos, lineInfo{text: stripped, indent: renderDepth, blank: stripped == ""})

		// Then advance the state for whatever this line contributes to depth.
		depth = max(scanLineForBraces(raw, &state, depth), 0)
	}

	var b strings.Builder
	prevBlank := false
	for i, info := range infos {
		if info.blank {
			// Collapse runs of blanks to a single blank line; never start the
			// file with a blank line.
			if prevBlank || i == 0 {
				continue
			}
			b.WriteByte('\n')
			prevBlank = true
			continue
		}
		for j := 0; j < info.indent; j++ {
			b.WriteString(indentUnit)
		}
		b.WriteString(info.text)
		b.WriteByte('\n')
		prevBlank = false
	}

	out := b.String()
	// Trim any trailing extra blank lines (we may have written a blank from a
	// mid-document blank that turned out to be the last meaningful line).
	out = strings.TrimRight(out, "\n") + "\n"
	if strings.TrimSpace(out) == "" {
		return ""
	}
	return out
}

func indentUnitFrom(opts lsp.FormattingOptions) string {
	if !opts.InsertSpaces {
		return "\t"
	}
	size := opts.TabSize
	if size <= 0 {
		size = 4
	}
	return strings.Repeat(" ", size)
}

// countLeadingClose returns how many `}` or `)` characters lead this trimmed
// line. Used so that closing braces render at the parent's indentation level.
func countLeadingClose(s string) int {
	count := 0
	for _, r := range s {
		switch r {
		case '}', ')':
			count++
		default:
			return count
		}
	}
	return count
}

// scanState tracks whether the cursor is inside a string or block comment so
// that braces in those contexts don't change indentation depth.
type scanState struct {
	inLineComment  bool
	inBlockComment bool
	inString       byte // 0 if not in string; otherwise the opening quote byte
	escape         bool
}

func scanLineForBraces(line string, st *scanState, depth int) int {
	st.inLineComment = false // line comments end at end of line
	for i := 0; i < len(line); i++ {
		c := line[i]
		if st.escape {
			st.escape = false
			continue
		}
		if st.inBlockComment {
			if c == '*' && i+1 < len(line) && line[i+1] == '/' {
				st.inBlockComment = false
				i++
			}
			continue
		}
		if st.inLineComment {
			continue
		}
		if st.inString != 0 {
			switch c {
			case '\\':
				st.escape = true
			case st.inString:
				st.inString = 0
			}
			continue
		}
		switch c {
		case '/':
			if i+1 < len(line) {
				switch line[i+1] {
				case '/':
					st.inLineComment = true
					i++
					continue
				case '*':
					st.inBlockComment = true
					i++
					continue
				}
			}
		case '"', '\'', '`':
			st.inString = c
		case '{', '(':
			depth++
		case '}', ')':
			depth--
		}
	}
	return depth
}
