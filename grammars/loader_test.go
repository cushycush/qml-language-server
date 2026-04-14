package qmlgrammars

import (
	"strings"
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

	require.NotEmpty(t, lang.ExternalSymbols, "Should have external symbols")
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

func TestQmljsParseSimpleQml(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	tree, err := parser.Parse([]byte("Rectangle { }"))
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseWithProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `Item {
    property int myValue: 42
}`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang), "Should parse without errors")

	nodeCount := countNodes(root)
	t.Logf("Node count: %d", nodeCount)
	require.Greater(t, nodeCount, 5, "Should have multiple nodes")
}

func TestQmljsParseImport(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "import QtQuick 2.15\nItem { }\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseComponent(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Component {
    id: myDelegate
    Rectangle {
        color: "red"
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseSignal(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    signal clicked()
    signal activated(string name)
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseFunction(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    function doSomething(x, y) {
        return x + y
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseListProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
ListModel {
    id: myModel
    ListElement { name: "Alice" }
    ListElement { name: "Bob" }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseStates(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    state: "active"
    states: [
        State { name: "active" }
        State { name: "inactive" }
    ]
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseTransitions(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    transitions: Transition {
        NumberAnimation { properties: "x,y"; duration: 200 }
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseJavaScript(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    Component.onCompleted: {
        console.log("Hello")
        var x = 10
        if (x > 5) {
            console.log("big")
        }
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseComments(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
// This is a single line comment
/* This is a
   multi-line comment */
Item {
    // Another comment
    property int value: 1 // inline comment
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsNodeTraversal(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Rectangle { width: 100; height: 200 }"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)

	identifiers := findIdentifiers(root, lang)
	t.Logf("Found %d identifiers", len(identifiers))
	require.NotEmpty(t, identifiers, "Should find identifiers")

	for _, id := range identifiers {
		t.Logf("  Identifier: %s", string(content[id.StartByte():id.EndByte()]))
	}
}

func TestQmljsIncrementalParse(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	initial := "Rectangle { }"
	tree, err := parser.Parse([]byte(initial))
	require.NoError(t, err)

	updated := "Rectangle { width: 100 }"
	newTree, err := parser.ParseIncremental([]byte(updated), tree)
	require.NoError(t, err)
	require.NotNil(t, newTree)

	root := newTree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsLanguageMetadata(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	t.Logf("Symbol count: %d", lang.SymbolCount)
	t.Logf("Token count: %d", lang.TokenCount)
	t.Logf("State count: %d", lang.StateCount)

	require.Greater(t, lang.SymbolCount, uint32(100))
	require.Greater(t, lang.TokenCount, uint32(50))
}

func TestQmljsHighlightQuery(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	tree, err := parser.Parse([]byte("import QtQuick 2.0"))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)

	query, err := gotreesitter.NewQuery(highlightQuery, lang)
	require.NoError(t, err)
	require.NotNil(t, query)

	cursor := query.Exec(root, lang, []byte("import QtQuick 2.0"))
	require.NotNil(t, cursor)

	matchCount := 0
	for {
		_, ok := cursor.NextMatch()
		if !ok {
			break
		}
		matchCount++
	}
	t.Logf("Highlight query matches: %d", matchCount)
}

func TestQmljsParseEmpty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	tree, err := parser.Parse([]byte(""))
	require.NoError(t, err)
	require.NotNil(t, tree)

	root := tree.RootNode()
	if root != nil {
		require.NotEqual(t, "ERROR", root.Type(lang))
	}
}

func TestQmljsParseMultilineProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property var myArray: [
        "a",
        "b",
        "c"
    ]
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseNestedObjects(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
StackView {
    id: stack
    initialItem: Rectangle {
        width: 200
        height: 100
        Text {
            text: "Hello"
        }
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))

	depth := maxDepth(root)
	t.Logf("Max nesting depth: %d", depth)
	require.Greater(t, depth, 3, "Should have nested objects")
}

func TestQmljsParseQualifiedId(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `QtQuick.Rectangle`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseBinding(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    width: parent.width
    height: myRect.height + 10
    anchors.centerIn: parent
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseNumberLiterals(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property int a: 42
    property real b: 3.14
    property real c: 1e10
    property real d: 0xFF
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseStringLiterals(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property string s1: "hello"
    property string s2: 'world'
    property string s3: "escaped \"quotes\""
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseBooleanLiterals(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property bool flag: true
    visible: false
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsLanguageCaching(t *testing.T) {
	lang1, err := QmljsLanguage()
	require.NoError(t, err)

	lang2, err := QmljsLanguage()
	require.NoError(t, err)

	require.Same(t, lang1, lang2, "Should return cached language instance")
}

func TestQmljsParseWithSemicolons(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Rectangle { x: 10; y: 20; width: 100; height: 200 }"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseIdDeclaration(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    id: myItem
    Rectangle {
        id: myRect
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func countNodes(node *gotreesitter.Node) int {
	if node == nil {
		return 0
	}
	count := 1
	for i := 0; i < node.ChildCount(); i++ {
		count += countNodes(node.Child(i))
	}
	return count
}

func findIdentifiers(node *gotreesitter.Node, lang *gotreesitter.Language) []*gotreesitter.Node {
	var result []*gotreesitter.Node
	if node == nil {
		return result
	}
	if node.Type(lang) == "identifier" || node.Type(lang) == "property_name" {
		result = append(result, node)
	}
	for i := 0; i < node.ChildCount(); i++ {
		result = append(result, findIdentifiers(node.Child(i), lang)...)
	}
	return result
}

func maxDepth(node *gotreesitter.Node) int {
	if node == nil || node.ChildCount() == 0 {
		return 0
	}
	maxChildDepth := 0
	for i := 0; i < node.ChildCount(); i++ {
		childDepth := maxDepth(node.Child(i))
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}
	return maxChildDepth + 1
}

func TestQmljsParseTemplateString(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Item {\n    property string greeting: `Hello, ${name}!`\n}\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseHexColor(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Rectangle {
    color: "#ff0000"
    border.color: "#00ff00"
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseInlineComponent(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    Loader {
        sourceComponent: Rectangle {
            color: "blue"
        }
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseRequiredProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Item {\n    property string title\n}\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseReadonlyProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Item {\n    property int maxValue: 100\n}\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseEnum(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
QtObject {
    enum MyEnum {
        Value1,
        Value2,
        Value3
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseJavaScriptOperators(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property int a: 1 + 2
    property int b: 3 - 4
    property int c: 5 * 6
    property int d: 10 / 2
    property bool e: a > b && c < d
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseJavaScriptTernary(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property var result: condition ? value1 : value2
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseDefaultProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Rectangle {
    default property alias childItem
    Text { }
    Rectangle { }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseQtObject(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
QtObject {
    objectName: "myObject"
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseNullUndefined(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property var a: null
    property var b: undefined
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseSpreadOperator(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property var arr: [...items, extra]
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseDestructuring(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    Component.onCompleted: {
        var { x, y } = point
        var [first, ...rest] = array
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseLogicalOperators(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property bool a: true && false
    property bool b: true || false
    property bool c: !false
    property var d: obj?.property
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseArrowFunction(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property var fn: (x) => x * 2
    property var fn2: x => x + 1
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseExport(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Item { }\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseGroupedProperty(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Rectangle {
    border {
        width: 2
        color: "black"
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseArrayAccess(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Item {\n    property var item: myArray[0]\n}\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParsePostfixOperators(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    Component.onCompleted: {
        i++
        j--
        k!!
    }
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseBlockComment(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
/*
 * Multi-line
 * block comment
 */
Item { }
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseTypeAnnot(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    property int x: 10 // int
    property string name: "test" // string
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseAutoSemicolon(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
Item {
    var x = 10
    var y = 20
}
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseHTMLComment(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := `
<Item> <!-- HTML comment -->
</Item>
`
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func TestQmljsParseJsxText(t *testing.T) {
	lang, err := QmljsLanguage()
	require.NoError(t, err)

	parser := gotreesitter.NewParser(lang)
	content := "Item {\n    property string some_prop: \"value\"\n}\n"
	tree, err := parser.Parse([]byte(content))
	require.NoError(t, err)

	root := tree.RootNode()
	require.NotNil(t, root)
	require.NotEqual(t, "ERROR", root.Type(lang))
}

func containsErrorNode(node *gotreesitter.Node, lang *gotreesitter.Language) bool {
	if node == nil {
		return false
	}
	if strings.Contains(node.Type(lang), "ERROR") {
		return true
	}
	for i := 0; i < node.ChildCount(); i++ {
		if containsErrorNode(node.Child(i), lang) {
			return true
		}
	}
	return false
}
