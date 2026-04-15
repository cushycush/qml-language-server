package handler

import (
	"context"
	"testing"

	"github.com/owenrumney/go-lsp/lsp"
)

func newWorkspaceTestHandler(t *testing.T, components ...workspaceComponent) *Handler {
	t.Helper()
	h := New(nil)
	if h.parser == nil {
		t.Skip("parser unavailable")
	}
	for _, c := range components {
		h.workspace.byName[c.Name] = c
	}
	return h
}

func TestWorkspaceSymbolReturnsAllOnEmptyQuery(t *testing.T) {
	h := newWorkspaceTestHandler(t,
		workspaceComponent{Name: "ShellRoot", Path: "/p/ShellRoot.qml", URI: "file:///p/ShellRoot.qml"},
		workspaceComponent{Name: "TopBar", Path: "/p/TopBar.qml", URI: "file:///p/TopBar.qml"},
	)

	syms, err := h.WorkspaceSymbol(context.Background(), &lsp.WorkspaceSymbolParams{Query: ""})
	if err != nil {
		t.Fatalf("WorkspaceSymbol: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(syms))
	}
	for _, s := range syms {
		if s.Kind != lsp.SymbolKindClass {
			t.Errorf("symbol %q kind = %v, want Class", s.Name, s.Kind)
		}
	}
}

func TestWorkspaceSymbolFiltersBySubstring(t *testing.T) {
	h := newWorkspaceTestHandler(t,
		workspaceComponent{Name: "ShellRoot", Path: "/p/ShellRoot.qml", URI: "file:///p/ShellRoot.qml"},
		workspaceComponent{Name: "TopBar", Path: "/p/TopBar.qml", URI: "file:///p/TopBar.qml"},
		workspaceComponent{Name: "Sidebar", Path: "/p/Sidebar.qml", URI: "file:///p/Sidebar.qml"},
	)

	syms, err := h.WorkspaceSymbol(context.Background(), &lsp.WorkspaceSymbolParams{Query: "bar"})
	if err != nil {
		t.Fatalf("WorkspaceSymbol: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("expected 2 results matching 'bar', got %d (%v)", len(syms), syms)
	}
}

func TestWorkspaceSymbolIsCaseInsensitive(t *testing.T) {
	h := newWorkspaceTestHandler(t,
		workspaceComponent{Name: "ShellRoot", Path: "/p/ShellRoot.qml", URI: "file:///p/ShellRoot.qml"},
	)

	syms, err := h.WorkspaceSymbol(context.Background(), &lsp.WorkspaceSymbolParams{Query: "SHELL"})
	if err != nil {
		t.Fatalf("WorkspaceSymbol: %v", err)
	}
	if len(syms) != 1 {
		t.Errorf("expected 1 result (case-insensitive), got %d", len(syms))
	}
}

func TestWorkspaceSymbolNoMatchReturnsEmpty(t *testing.T) {
	h := newWorkspaceTestHandler(t,
		workspaceComponent{Name: "ShellRoot", Path: "/p/ShellRoot.qml", URI: "file:///p/ShellRoot.qml"},
	)

	syms, err := h.WorkspaceSymbol(context.Background(), &lsp.WorkspaceSymbolParams{Query: "nothingmatches"})
	if err != nil {
		t.Fatalf("WorkspaceSymbol: %v", err)
	}
	if len(syms) != 0 {
		t.Errorf("expected no results, got %d", len(syms))
	}
}
