package handler

import (
	"context"
	"strings"

	"github.com/owenrumney/go-lsp/lsp"
)

// WorkspaceSymbol resolves workspace/symbol queries by scanning the indexed
// QML components and returning any whose top-level type name matches the
// query as a case-insensitive substring. The workspace index is built at
// startup (and refreshed via DidChangeWatchedFiles) so this is a cheap
// in-memory lookup.
func (h *Handler) WorkspaceSymbol(_ context.Context, params *lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	if h.workspace == nil {
		return []lsp.SymbolInformation{}, nil
	}
	query := strings.ToLower(strings.TrimSpace(params.Query))
	components := h.workspace.all()

	results := make([]lsp.SymbolInformation, 0, len(components))
	for _, c := range components {
		if query != "" && !strings.Contains(strings.ToLower(c.Name), query) {
			continue
		}
		results = append(results, lsp.SymbolInformation{
			Name: c.Name,
			Kind: lsp.SymbolKindClass,
			Location: lsp.Location{
				URI:   c.URI,
				Range: lsp.Range{}, // Top of the file; we don't track exact byte ranges in the index.
			},
			ContainerName: c.Path,
		})
	}
	return results, nil
}
