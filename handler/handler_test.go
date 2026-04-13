package handler

import (
	"testing"
)

func TestGetTypeInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOK   bool
		wantType string
	}{
		{"Rectangle type", "Rectangle", true, "Object"},
		{"Text type", "Text", true, "Object"},
		{"Item type", "Item", true, "Object"},
		{"Unknown type", "FooBar", false, ""},
		{"Empty string", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := getTypeInfo(tt.input)
			if ok != tt.wantOK {
				t.Errorf("getTypeInfo(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && info.Type != tt.wantType {
				t.Errorf("getTypeInfo(%q).Type = %v, want %v", tt.input, info.Type, tt.wantType)
			}
		})
	}
}

func TestGetPropertyInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOK   bool
		wantType string
	}{
		{"width property", "width", true, "real"},
		{"color property", "color", true, "color"},
		{"text property", "text", true, "string"},
		{"Unknown property", "fooBar", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := getPropertyInfo(tt.input)
			if ok != tt.wantOK {
				t.Errorf("getPropertyInfo(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && info.Type != tt.wantType {
				t.Errorf("getPropertyInfo(%q).Type = %v, want %v", tt.input, info.Type, tt.wantType)
			}
		})
	}
}

func TestGetLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single line", "hello", []string{"hello"}},
		{"two lines", "hello\nworld", []string{"hello", "world"}},
		{"three lines", "line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"empty string", "", []string{""}},
		{"newline only", "\n", []string{"", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("getLines(%q) returned %d lines, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("getLines(%q)[%d] = %q, want %q", tt.input, i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractWordAt(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		wantWord string
	}{
		{"word at start", "hello world", 0, "hello"},
		{"word in middle", "hello world", 6, "world"},
		{"word at end", "hello world", 10, "world"},
		{"single char", "a", 0, "a"},
		{"empty string", "", 0, ""},
		{"position out of bounds", "hello", 10, ""},
		{"camelCase word", "helloWorld", 5, "helloWorld"},
		{"with underscore", "my_variable", 3, "my_variable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWordAt(tt.text, tt.pos)
			if result != tt.wantWord {
				t.Errorf("extractWordAt(%q, %d) = %q, want %q", tt.text, tt.pos, result, tt.wantWord)
			}
		})
	}
}

func TestIsIdentChar(t *testing.T) {
	tests := []struct {
		name     string
		char     byte
		expected bool
	}{
		{"lowercase letter", 'a', true},
		{"uppercase letter", 'Z', true},
		{"digit", '5', true},
		{"underscore", '_', true},
		{"space", ' ', false},
		{"period", '.', false},
		{"colon", ':', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIdentChar(tt.char)
			if result != tt.expected {
				t.Errorf("isIdentChar(%q) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "hello world", 8, "hello..."},
		{"maxLen less than 3", "hello", 2, "he"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestIsSignalHandler(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"onClicked", "onClicked", true},
		{"onPressed", "onPressed", true},
		{"onReleased", "onReleased", true},
		{"onEntered", "onEntered", true},
		{"onExited", "onExited", true},
		{"on", "on", false},
		{"clicked", "clicked", false},
		{"property", "property", false},
		{"x", "x", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSignalHandler(tt.input)
			if result != tt.expected {
				t.Errorf("isSignalHandler(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsUpperCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"all uppercase", "HELLO", true},
		{"all lowercase", "hello", false},
		{"mixed", "Hello", false},
		{"single uppercase", "A", true},
		{"single lowercase", "a", false},
		{"with numbers", "HELLO123", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUpperCase(tt.input)
			if result != tt.expected {
				t.Errorf("isUpperCase(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrimLeadingWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no whitespace", "hello", "hello"},
		{"leading spaces", "  hello", "hello"},
		{"leading tab", "\thello", "hello"},
		{"leading spaces and tab", "  \thello", "hello"},
		{"only whitespace", "   \t", ""},
		{"empty", "", ""},
		{"whitespace only", "  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimLeadingWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("trimLeadingWhitespace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectCompletionContext(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		expected CompletionContext
	}{
		{"import statement", "import Qt", 7, ContextImport},
		{"type name", "  Rectangle", 3, ContextTypeName},
		{"property", "width: ", 6, ContextAfterColon},
		{"default", "someText", 4, ContextDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectCompletionContext(tt.text, tt.pos)
			if result != tt.expected {
				t.Errorf("detectCompletionContext(%q, %d) = %v, want %v", tt.text, tt.pos, result, tt.expected)
			}
		})
	}
}
