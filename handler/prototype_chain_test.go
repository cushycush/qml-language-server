package handler

import (
	"testing"
)

func TestResolvePrototypeChain(t *testing.T) {
	// Set up a mini prototype index: Rectangle -> Item -> QObject
	protoMu.Lock()
	saved := make(map[string]*QMLTypesComponent, len(protoIndex))
	for k, v := range protoIndex {
		saved[k] = v
	}

	protoIndex["QQuickRectangle"] = &QMLTypesComponent{
		Name:      "QQuickRectangle",
		Prototype: "QQuickItem",
		Exports:   []string{"QtQuick/Rectangle 6.0"},
	}
	protoIndex["QQuickItem"] = &QMLTypesComponent{
		Name:      "QQuickItem",
		Prototype: "QObject",
		Exports:   []string{"QtQuick/Item 6.0"},
	}
	protoIndex["QObject"] = &QMLTypesComponent{
		Name:    "QObject",
		Exports: []string{"QtQml/QtObject 6.0"},
	}
	protoMu.Unlock()

	defer func() {
		protoMu.Lock()
		protoIndex = saved
		protoMu.Unlock()
	}()

	chain := resolvePrototypeChain("QQuickItem")
	if len(chain) < 1 {
		t.Fatalf("expected at least 1 in chain, got %d", len(chain))
	}
	if chain[0] != "Item" {
		t.Errorf("chain[0] = %q, want Item", chain[0])
	}
}

func TestBuildInheritanceChains(t *testing.T) {
	// Save and restore global state.
	protoMu.Lock()
	savedProto := make(map[string]*QMLTypesComponent, len(protoIndex))
	for k, v := range protoIndex {
		savedProto[k] = v
	}
	savedBase := make(map[string][]string, len(baseTypes))
	for k, v := range baseTypes {
		savedBase[k] = v
	}
	protoMu.Unlock()

	defer func() {
		protoMu.Lock()
		protoIndex = savedProto
		protoMu.Unlock()
		// Restore baseTypes (no mutex needed — single-threaded test).
		for k := range baseTypes {
			if _, was := savedBase[k]; !was {
				delete(baseTypes, k)
			}
		}
		for k, v := range savedBase {
			baseTypes[k] = v
		}
	}()

	// Set up index.
	protoMu.Lock()
	protoIndex["QQuickText"] = &QMLTypesComponent{
		Name:      "QQuickText",
		Prototype: "QQuickItem",
		Exports:   []string{"QtQuick/Text 6.0"},
	}
	protoIndex["QQuickItem"] = &QMLTypesComponent{
		Name:      "QQuickItem",
		Prototype: "QObject",
		Exports:   []string{"QtQuick/Item 6.0"},
	}
	protoIndex["QObject"] = &QMLTypesComponent{
		Name:    "QObject",
		Exports: []string{"QtQml/QtObject 6.0"},
	}
	protoMu.Unlock()

	// Remove any existing chain for Text so we can verify it gets built.
	delete(baseTypes, "Text")

	buildInheritanceChains()

	chain, ok := baseTypes["Text"]
	if !ok {
		t.Fatal("expected baseTypes[Text] to be populated")
	}
	if len(chain) < 2 {
		t.Fatalf("expected chain length >= 2, got %d: %v", len(chain), chain)
	}
	if chain[0] != "Item" {
		t.Errorf("chain[0] = %q, want Item", chain[0])
	}
	if chain[1] != "QtObject" {
		t.Errorf("chain[1] = %q, want QtObject", chain[1])
	}
}

func TestBuildInheritanceChainsPreservesHandCoded(t *testing.T) {
	// ApplicationWindow has a hand-coded chain — verify it's not overwritten.
	original := baseTypes["ApplicationWindow"]
	if original == nil {
		t.Skip("ApplicationWindow not in baseTypes")
	}

	buildInheritanceChains()

	after := baseTypes["ApplicationWindow"]
	if len(after) != len(original) {
		t.Errorf("hand-coded chain was overwritten: before %v, after %v", original, after)
	}
}
