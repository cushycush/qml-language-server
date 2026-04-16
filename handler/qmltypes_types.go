package handler

// Parsed representation of Qt .qmltypes files. These files describe the types,
// properties, signals, methods, and enums that a QML module exports.

type QMLTypesModule struct {
	Components []QMLTypesComponent
}

type QMLTypesComponent struct {
	Name            string
	File            string
	Prototype       string   // base type (C++ class name)
	Exports         []string // e.g. "QtQuick/Rectangle 2.0"
	DefaultProperty string
	AccessSemantics string // "reference", "value", "sequence"
	AttachedType    string
	IsSingleton     bool
	IsCreatable     bool

	Properties []QMLTypesProperty
	Signals    []QMLTypesSignal
	Methods    []QMLTypesMethod
	Enums      []QMLTypesEnum
}

// ExportedName returns the QML-visible name from the first export entry
// (e.g. "Rectangle" from "QtQuick/Rectangle 2.0"), or falls back to the
// C++ class name stripped of its Q prefix.
func (c *QMLTypesComponent) ExportedName() string {
	for _, exp := range c.Exports {
		name, _ := parseExport(exp)
		if name != "" {
			return name
		}
	}
	return ""
}

// ExportedModule returns the module from the first export entry (e.g.
// "QtQuick" from "QtQuick/Rectangle 2.0").
func (c *QMLTypesComponent) ExportedModule() string {
	for _, exp := range c.Exports {
		_, mod := parseExport(exp)
		if mod != "" {
			return mod
		}
	}
	return ""
}

type QMLTypesProperty struct {
	Name               string
	Type               string
	IsReadonly         bool
	IsList             bool
	IsPropertyConstant bool
	Notify             string // change-signal name
}

type QMLTypesSignal struct {
	Name       string
	Parameters []QMLTypesParameter
}

type QMLTypesMethod struct {
	Name       string
	ReturnType string
	Parameters []QMLTypesParameter
}

type QMLTypesParameter struct {
	Name string
	Type string
}

type QMLTypesEnum struct {
	Name   string
	Values []string // ordered enum member names
	IsFlag bool
}

// QMLDirModule represents a parsed qmldir file.
type QMLDirModule struct {
	Name     string   // module name, e.g. "QtQuick"
	TypeInfo string   // relative path to .qmltypes file
	Depends  []string // dependent modules
	Imports  []string // auto-imported modules
	Dir      string   // directory containing the qmldir file
}

// parseExport splits "QtQuick/Rectangle 2.0" into ("Rectangle", "QtQuick").
func parseExport(exp string) (name, module string) {
	// Strip version suffix.
	sp := exp
	if i := lastIndexByte(sp, ' '); i >= 0 {
		sp = sp[:i]
	}
	if i := lastIndexByte(sp, '/'); i >= 0 {
		return sp[i+1:], sp[:i]
	}
	return sp, ""
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
