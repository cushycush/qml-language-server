package handler

import (
	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
)

// Position/byte-offset conversions used by every feature handler.
//
// LSP defines `Position.Character` as a UTF-16 code-unit offset by default. We
// treat it as a byte offset on the assumption that QML source is ASCII — the
// go-lsp library version in use doesn't support negotiating a different
// encoding. For non-ASCII content this is wrong by one char per multi-byte
// sequence; upgrade to a UTF-16-aware implementation if that becomes real.

// positionToByte converts an LSP Position into a byte offset within content.
// Out-of-range lines clamp to the last line; out-of-range characters clamp to
// the end of the target line.
func positionToByte(content []byte, pos lsp.Position) uint32 {
	var offset uint32
	line := int(pos.Line)
	for i := 0; i < line; i++ {
		nl := indexOfByte(content[offset:], '\n')
		if nl < 0 {
			return uint32(len(content))
		}
		offset += uint32(nl) + 1
	}
	lineEnd := offset
	for lineEnd < uint32(len(content)) && content[lineEnd] != '\n' {
		lineEnd++
	}
	char := uint32(pos.Character)
	if char > lineEnd-offset {
		char = lineEnd - offset
	}
	return offset + char
}

// byteOffsetToPosition is the inverse of positionToByte.
func byteOffsetToPosition(content []byte, offset uint32) lsp.Position {
	line := 0
	char := 0
	end := int(offset)
	if end > len(content) {
		end = len(content)
	}
	for i := 0; i < end; i++ {
		if content[i] == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}
	return lsp.Position{Line: line, Character: char}
}

func indexOfByte(data []byte, ch byte) int {
	for i, b := range data {
		if b == ch {
			return i
		}
	}
	return -1
}

// nodeRange returns an LSP Range for a tree-sitter node in the given content.
func nodeRange(content []byte, node *gotreesitter.Node) lsp.Range {
	return lsp.Range{
		Start: byteOffsetToPosition(content, node.StartByte()),
		End:   byteOffsetToPosition(content, node.EndByte()),
	}
}

// nodeLocation returns an LSP Location for a node inside a document URI.
func nodeLocation(uri lsp.DocumentURI, content []byte, node *gotreesitter.Node) lsp.Location {
	return lsp.Location{URI: uri, Range: nodeRange(content, node)}
}

// findSmallestNodeAt descends into node's subtree to find the tightest node
// whose byte range contains offset. Returns nil when offset is outside node.
func findSmallestNodeAt(node *gotreesitter.Node, offset uint32, _ *gotreesitter.Language) *gotreesitter.Node {
	if node == nil || offset < node.StartByte() || offset > node.EndByte() {
		return nil
	}
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if offset >= child.StartByte() && offset <= child.EndByte() {
			if smaller := findSmallestNodeAt(child, offset, nil); smaller != nil {
				return smaller
			}
		}
	}
	return node
}
