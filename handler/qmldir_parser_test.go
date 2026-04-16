package handler

import (
	"testing"
)

func TestParseQMLDir(t *testing.T) {
	input := `module QtQuick
linktarget Qt6::qtquick2plugin
optional plugin qtquick2plugin
classname QtQuick2Plugin
designersupported
typeinfo plugins.qmltypes
import QtQml auto
prefer :/qt-project.org/imports/QtQuick/
`
	m := ParseQMLDir(input)
	if m.Name != "QtQuick" {
		t.Errorf("Name = %q, want QtQuick", m.Name)
	}
	if m.TypeInfo != "plugins.qmltypes" {
		t.Errorf("TypeInfo = %q, want plugins.qmltypes", m.TypeInfo)
	}
	if len(m.Imports) != 1 || m.Imports[0] != "QtQml" {
		t.Errorf("Imports = %v, want [QtQml]", m.Imports)
	}
}

func TestParseQMLDirWithDependsAndComments(t *testing.T) {
	input := `# Example module
module QtQuick.Controls
depends QtQuick 2.15
depends QtQuick.Templates 2.15
typeinfo plugins.qmltypes
`
	m := ParseQMLDir(input)
	if m.Name != "QtQuick.Controls" {
		t.Errorf("Name = %q", m.Name)
	}
	if len(m.Depends) != 2 {
		t.Errorf("expected 2 depends, got %d", len(m.Depends))
	}
}
