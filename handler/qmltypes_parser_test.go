package handler

import (
	"os"
	"testing"
)

const testQMLTypes = `import QtQuick.tooling 1.2

// Auto-generated test file.

Module {
    Component {
        file: "qquickrectangle.h"
        name: "QQuickRectangle"
        accessSemantics: "reference"
        prototype: "QQuickItem"
        exports: ["QtQuick/Rectangle 2.0", "QtQuick/Rectangle 6.0"]
        Property { name: "color"; type: "QColor" }
        Property { name: "radius"; type: "double" }
        Property {
            name: "border"
            type: "QQuickPen"
            isReadonly: true
            isPointer: true
        }
        Signal {
            name: "colorChanged"
        }
        Signal {
            name: "radiusChanged"
            Parameter { name: "radius"; type: "double" }
        }
        Method {
            name: "mapToItem"
            type: "QPointF"
            Parameter { name: "item"; type: "QQuickItem" }
            Parameter { name: "point"; type: "QPointF" }
        }
        Enum {
            name: "VerticalAlignment"
            values: ["AlignTop", "AlignVCenter", "AlignBottom"]
        }
    }
    Component {
        name: "QQuickItem"
        accessSemantics: "reference"
        exports: ["QtQuick/Item 2.0"]
        Property { name: "width"; type: "double" }
        Property { name: "height"; type: "double" }
        Property { name: "visible"; type: "bool" }
    }
}
`

func TestParseQMLTypesBasic(t *testing.T) {
	mod, err := ParseQMLTypes(testQMLTypes)
	if err != nil {
		t.Fatalf("ParseQMLTypes: %v", err)
	}
	if len(mod.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(mod.Components))
	}
}

func TestParseQMLTypesComponentAttributes(t *testing.T) {
	mod, _ := ParseQMLTypes(testQMLTypes)
	rect := mod.Components[0]

	if rect.Name != "QQuickRectangle" {
		t.Errorf("name = %q, want QQuickRectangle", rect.Name)
	}
	if rect.Prototype != "QQuickItem" {
		t.Errorf("prototype = %q, want QQuickItem", rect.Prototype)
	}
	if rect.AccessSemantics != "reference" {
		t.Errorf("accessSemantics = %q, want reference", rect.AccessSemantics)
	}
	if len(rect.Exports) != 2 {
		t.Fatalf("expected 2 exports, got %d", len(rect.Exports))
	}
	if rect.ExportedName() != "Rectangle" {
		t.Errorf("ExportedName() = %q, want Rectangle", rect.ExportedName())
	}
	if rect.ExportedModule() != "QtQuick" {
		t.Errorf("ExportedModule() = %q, want QtQuick", rect.ExportedModule())
	}
}

func TestParseQMLTypesProperties(t *testing.T) {
	mod, _ := ParseQMLTypes(testQMLTypes)
	rect := mod.Components[0]

	if len(rect.Properties) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(rect.Properties))
	}

	color := rect.Properties[0]
	if color.Name != "color" || color.Type != "QColor" {
		t.Errorf("property[0] = {%q, %q}, want {color, QColor}", color.Name, color.Type)
	}

	border := rect.Properties[2]
	if !border.IsReadonly {
		t.Error("border should be readonly")
	}
}

func TestParseQMLTypesSignals(t *testing.T) {
	mod, _ := ParseQMLTypes(testQMLTypes)
	rect := mod.Components[0]

	if len(rect.Signals) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(rect.Signals))
	}
	if rect.Signals[0].Name != "colorChanged" {
		t.Errorf("signal[0].Name = %q", rect.Signals[0].Name)
	}
	if len(rect.Signals[1].Parameters) != 1 {
		t.Errorf("radiusChanged should have 1 param, got %d", len(rect.Signals[1].Parameters))
	}
}

func TestParseQMLTypesMethods(t *testing.T) {
	mod, _ := ParseQMLTypes(testQMLTypes)
	rect := mod.Components[0]

	if len(rect.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(rect.Methods))
	}
	m := rect.Methods[0]
	if m.Name != "mapToItem" || m.ReturnType != "QPointF" {
		t.Errorf("method = {%q, %q}", m.Name, m.ReturnType)
	}
	if len(m.Parameters) != 2 {
		t.Errorf("expected 2 params, got %d", len(m.Parameters))
	}
}

func TestParseQMLTypesEnums(t *testing.T) {
	mod, _ := ParseQMLTypes(testQMLTypes)
	rect := mod.Components[0]

	if len(rect.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(rect.Enums))
	}
	e := rect.Enums[0]
	if e.Name != "VerticalAlignment" {
		t.Errorf("enum name = %q", e.Name)
	}
	if len(e.Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(e.Values))
	}
}

func TestParseQMLTypesQt5EnumFormat(t *testing.T) {
	input := `import QtQuick.tooling 1.2
Module {
    Component {
        name: "QQuickText"
        exports: ["QtQuick/Text 2.0"]
        Enum {
            name: "WrapMode"
            values: {
                "NoWrap": 0,
                "WordWrap": 1,
                "WrapAnywhere": 3,
                "Wrap": 4
            }
        }
    }
}
`
	mod, err := ParseQMLTypes(input)
	if err != nil {
		t.Fatalf("ParseQMLTypes: %v", err)
	}
	if len(mod.Components) != 1 || len(mod.Components[0].Enums) != 1 {
		t.Fatal("expected 1 component with 1 enum")
	}
	e := mod.Components[0].Enums[0]
	if len(e.Values) != 4 {
		t.Errorf("expected 4 enum values, got %d: %v", len(e.Values), e.Values)
	}
}

func TestCppTypeToQML(t *testing.T) {
	cases := []struct{ in, want string }{
		{"double", "real"},
		{"QString", "string"},
		{"QColor", "color"},
		{"bool", "bool"},
		{"int", "int"},
		{"QUrl", "url"},
		{"QVariant", "var"},
		{"QQuickItem*", "Item"},
		{"", "void"},
	}
	for _, tc := range cases {
		if got := cppTypeToQML(tc.in); got != tc.want {
			t.Errorf("cppTypeToQML(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseExport(t *testing.T) {
	cases := []struct {
		in         string
		wantName   string
		wantModule string
	}{
		{"QtQuick/Rectangle 2.0", "Rectangle", "QtQuick"},
		{"QtQuick.Controls/Button 6.0", "Button", "QtQuick.Controls"},
		{"Rectangle", "Rectangle", ""},
	}
	for _, tc := range cases {
		name, mod := parseExport(tc.in)
		if name != tc.wantName || mod != tc.wantModule {
			t.Errorf("parseExport(%q) = (%q, %q), want (%q, %q)", tc.in, name, mod, tc.wantName, tc.wantModule)
		}
	}
}

func TestParseRealQtQuickQMLTypes(t *testing.T) {
	path := "/usr/lib/qt6/qml/QtQuick/plugins.qmltypes"
	if _, err := os.Stat(path); err != nil {
		t.Skip("Qt6 not installed at " + path)
	}

	mod, err := ParseQMLTypesFile(path)
	if err != nil {
		t.Fatalf("ParseQMLTypesFile: %v", err)
	}
	if len(mod.Components) < 50 {
		t.Errorf("expected 50+ components from QtQuick, got %d", len(mod.Components))
	}

	// Spot-check that Rectangle was parsed.
	found := false
	for _, c := range mod.Components {
		if c.ExportedName() == "Rectangle" {
			found = true
			if len(c.Properties) == 0 {
				t.Error("Rectangle should have properties")
			}
			break
		}
	}
	if !found {
		t.Error("Rectangle not found in QtQuick module")
	}
}
