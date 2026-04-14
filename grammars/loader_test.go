package qmlgrammars

import (
	"testing"

	"github.com/odvcencio/gotreesitter"
	"github.com/stretchr/testify/require"
)

func TestQmljsLanguage(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)
	require.NotNil(t, lang)
}

func TestQmljsExternalSymbols(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)
	require.NotNil(t, lang)

	t.Logf("External symbols count: %d", len(lang.ExternalSymbols))
	for i, sym := range lang.ExternalSymbols {
		t.Logf("  External symbol %d: %v", i, sym)
	}
}

func TestQmljsParse(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	tree, err := parser.Parse([]byte("import QtQuick 2.0\n\nRectangle { width: 100 }"))
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	rootType := root.Type(lang)
	t.Logf("Root node type: %s", rootType)
	require.NotEqual(t, "ERROR", rootType, "Parse should not have errors")
}
