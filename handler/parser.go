package handler

import (
	"bytes"
	"sync"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
	"github.com/owenrumney/go-lsp/lsp"
)

type QMLParser struct {
	parser *gotreesitter.Parser
	lang   *gotreesitter.Language
	mu     sync.RWMutex
	trees  map[lsp.DocumentURI]*gotreesitter.Tree
}

func NewQMLParser() *QMLParser {
	lang := grammars.QmljsLanguage()
	if lang == nil {
		return nil
	}
	return &QMLParser{
		parser: gotreesitter.NewParser(lang),
		lang:   lang,
		trees:  make(map[lsp.DocumentURI]*gotreesitter.Tree),
	}
}

func (p *QMLParser) Language() *gotreesitter.Language {
	return p.lang
}

func (p *QMLParser) Parse(uri lsp.DocumentURI, content string) *gotreesitter.Tree {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldTree := p.trees[uri]
	newContent := []byte(content)

	var tree *gotreesitter.Tree
	if oldTree != nil {
		var err error
		tree, err = p.parser.ParseIncremental(newContent, oldTree)
		if err != nil {
			return nil
		}
	} else {
		var err error
		tree, err = p.parser.Parse(newContent)
		if err != nil {
			return nil
		}
	}

	p.trees[uri] = tree
	return tree
}

func (p *QMLParser) Invalidate(uri lsp.DocumentURI) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.trees, uri)
}

func (p *QMLParser) GetTree(uri lsp.DocumentURI) *gotreesitter.Tree {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.trees[uri]
}

func (p *QMLParser) GetNodeAt(uri lsp.DocumentURI, pos lsp.Position) *gotreesitter.Node {
	tree, ok := p.trees[uri]
	if !ok {
		return nil
	}

	root := tree.RootNode()
	if root == nil {
		return nil
	}

	byteOffset := positionToByteOffset([]byte(getDocContent(uri)), pos)
	return getNodeAtByte(root, byteOffset)
}

func getDocContent(uri lsp.DocumentURI) string {
	return ""
}

func positionToByteOffset(content []byte, pos lsp.Position) uint32 {
	line := int(pos.Line)
	char := int(pos.Character)

	offset := uint32(0)
	for i := 0; i < line && i < bytes.Count(content, []byte{'\n'}); i++ {
		idx := bytes.Index(content[offset:], []byte{'\n'})
		if idx == -1 {
			return offset
		}
		offset += uint32(idx) + 1
	}

	for i := 0; i < char; i++ {
		if offset+uint32(i) >= uint32(len(content)) {
			break
		}
		if content[offset+uint32(i)] == '\n' {
			break
		}
	}

	return offset + uint32(char)
}

func getNodeAtByte(node *gotreesitter.Node, offset uint32) *gotreesitter.Node {
	if node == nil {
		return nil
	}

	if offset < node.StartByte() || offset >= node.EndByte() {
		return nil
	}

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && offset >= child.StartByte() && offset < child.EndByte() {
			return getNodeAtByte(child, offset)
		}
	}

	return node
}

type ParseResult struct {
	Root    *gotreesitter.Node
	Tree    *gotreesitter.Tree
	Content []byte
}

func ParseQML(content string) (*ParseResult, error) {
	parser := NewQMLParser()
	if parser == nil {
		return nil, nil
	}

	tree, err := parser.parser.Parse([]byte(content))
	if err != nil {
		return nil, err
	}

	return &ParseResult{
		Root:    tree.RootNode(),
		Tree:    tree,
		Content: []byte(content),
	}, nil
}
