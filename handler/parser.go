package handler

import (
	"sync"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
	qmlgrammars "qml-language-server/grammars"
)

type QMLParser struct {
	parser *gotreesitter.Parser
	lang   *gotreesitter.Language
	mu     sync.RWMutex
	trees  map[lsp.DocumentURI]*gotreesitter.Tree
}

func NewQMLParser() *QMLParser {
	lang, err := qmlgrammars.QmljsLanguage()
	if err != nil || lang == nil {
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

func (p *QMLParser) GetNodeAt(uri lsp.DocumentURI, pos lsp.Position, content []byte) *gotreesitter.Node {
	p.mu.RLock()
	tree, ok := p.trees[uri]
	p.mu.RUnlock()
	if !ok {
		return nil
	}

	root := tree.RootNode()
	if root == nil {
		return nil
	}

	offset := positionToByte(content, pos)
	// When the cursor sits immediately after the last character of an
	// identifier (e.g. "foo|") the byte offset equals the node's EndByte.
	// Bias it one byte left so we still hit the identifier and features like
	// hover and go-to-definition work at the end of a word.
	if offset > 0 && offset == uint32(len(content)) {
		offset--
	} else if offset > 0 && isWordByte(contentByteAt(content, offset-1)) && !isWordByte(contentByteAt(content, offset)) {
		offset--
	}
	return findSmallestNodeAt(root, offset, nil)
}

func contentByteAt(content []byte, i uint32) byte {
	if int(i) >= len(content) {
		return 0
	}
	return content[i]
}

func isWordByte(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' || b == '$'
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
