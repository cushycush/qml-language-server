package handler

import "testing"

func TestDetectCompletionContext(t *testing.T) {
	tests := []struct {
		name string
		text string
		pos  int
		want CompletionContext
	}{
		{"import with trailing dot", "import QtQuick.", 15, ContextImport},
		{"import with module and version", "import QtQuick.Controls 2.", 26, ContextImport},
		{"bare import keyword", "import ", 7, ContextImport},
		{"property dot chain", "anchors.", 8, ContextProperty},
		{"value side after colon", "color: ", 7, ContextAfterColon},
		{"empty line is default", "", 0, ContextDefault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectCompletionContext(tt.text, tt.pos); got != tt.want {
				t.Fatalf("detectCompletionContext(%q, %d) = %v, want %v", tt.text, tt.pos, got, tt.want)
			}
		})
	}
}
