package handler

import (
	"github.com/odvcencio/gotreesitter"
)

// buildIDTypeIndex walks the tree and returns a map from `id` name to the
// enclosing object's type name. QML ids are file-scoped, so a single map per
// document is sufficient for property resolution on member expressions like
// `root.width`.
//
// Duplicate ids (which QML forbids) resolve first-writer-wins — whichever the
// walk sees first. This is consistent with how the existing id → location
// lookup in definition.go behaves.
func buildIDTypeIndex(root *gotreesitter.Node, lang *gotreesitter.Language, content []byte) map[string]string {
	index := map[string]string{}
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) != "ui_binding" {
			return true
		}
		if bindingName(n, lang, content) != "id" {
			return true
		}
		idName := bindingValueIdentifier(n, lang, content)
		if idName == "" {
			return true
		}
		typeName := resolveEnclosingType(n, lang, content)
		if typeName == "" {
			return true
		}
		if _, exists := index[idName]; !exists {
			index[idName] = typeName
		}
		return true
	})
	return index
}

// resolveEnclosingType finds the type name of the object that contains the
// given binding. Prefers the parse tree (ui_object_definition ancestor); on
// partial parses — which happen while the user is mid-edit and the file is
// temporarily invalid — falls back to a textual brace-balance scan.
func resolveEnclosingType(b *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for anc := b.Parent(); anc != nil; anc = anc.Parent() {
		if anc.Type(lang) != "ui_object_definition" {
			continue
		}
		if name := objectDefinitionTypeName(anc, lang, content); name != "" {
			return name
		}
	}
	return enclosingTypeFromText(content, b.StartByte())
}

// objectDefinitionTypeName returns the type name of a ui_object_definition
// node (e.g. "Rectangle" for `Rectangle { ... }`). Handles the dotted form
// `QtQuick.Window` by returning the last segment.
func objectDefinitionTypeName(obj *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < obj.ChildCount(); i++ {
		c := obj.Child(i)
		if c == nil {
			continue
		}
		t := c.Type(lang)
		if t == "identifier" || t == "nested_identifier" {
			return lastDottedSegment(string(content[c.StartByte():c.EndByte()]))
		}
	}
	return ""
}

func bindingName(b *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < b.ChildCount(); i++ {
		c := b.Child(i)
		if c == nil {
			continue
		}
		if c.Type(lang) == "identifier" {
			return string(content[c.StartByte():c.EndByte()])
		}
	}
	return ""
}

func bindingValueIdentifier(b *gotreesitter.Node, lang *gotreesitter.Language, content []byte) string {
	for i := 0; i < b.ChildCount(); i++ {
		c := b.Child(i)
		if c == nil || c.Type(lang) != "expression_statement" {
			continue
		}
		for j := 0; j < c.ChildCount(); j++ {
			cc := c.Child(j)
			if cc == nil {
				continue
			}
			if cc.Type(lang) == "identifier" {
				return string(content[cc.StartByte():cc.EndByte()])
			}
		}
	}
	return ""
}
