package handler

import (
	"context"
	"log/slog"

	"github.com/owenrumney/go-lsp/lsp"
	"github.com/owenrumney/go-lsp/server"
)

type Handler struct {
	logger    *slog.Logger
	documents map[lsp.DocumentURI]string
	parser    *QMLParser
}

func New(logger *slog.Logger) *Handler {
	return &Handler{
		logger:    logger,
		documents: make(map[lsp.DocumentURI]string),
		parser:    NewQMLParser(),
	}
}

func (h *Handler) Serve(ctx context.Context) error {
	srv := server.NewServer(h)
	return srv.Run(ctx, server.RunStdio())
}

func (h *Handler) Initialize(_ context.Context, _ *lsp.InitializeParams) (*lsp.InitializeResult, error) {
	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptions{
				OpenClose: boolPtr(true),
				Change:    lsp.SyncIncremental,
				WillSave:  boolPtr(true),
				Save:      &lsp.SaveOptions{IncludeText: boolPtr(true)},
			},
			HoverProvider: boolPtr(true),
			CompletionProvider: &lsp.CompletionOptions{
				ResolveProvider:   boolPtr(true),
				TriggerCharacters: []string{".", ":", "<", "\"", "/"},
			},
			DefinitionProvider:        boolPtr(true),
			TypeDefinitionProvider:    boolPtr(true),
			ReferencesProvider:        boolPtr(true),
			DocumentSymbolProvider:    boolPtr(true),
			WorkspaceSymbolProvider:   boolPtr(true),
			DocumentHighlightProvider: boolPtr(true),
			SignatureHelpProvider: &lsp.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
			CodeActionProvider: &lsp.CodeActionOptions{
				ResolveProvider: boolPtr(true),
			},
			RenameProvider: &lsp.RenameOptions{
				PrepareProvider: boolPtr(true),
			},
			ExecuteCommandProvider: &lsp.ExecuteCommandOptions{
				Commands: []string{},
			},
			DiagnosticProvider: &lsp.DiagnosticOptions{},
			SemanticTokensProvider: &lsp.SemanticTokensOptions{
				Full:  &lsp.SemanticTokensFull{},
				Range: boolPtr(true),
			},
			FoldingRangeProvider:       boolPtr(true),
			SelectionRangeProvider:     boolPtr(true),
			CallHierarchyProvider:      boolPtr(true),
			TypeHierarchyProvider:      boolPtr(true),
			InlayHintProvider:          &lsp.InlayHintOptions{},
			LinkedEditingRangeProvider: boolPtr(true),
		},
		ServerInfo: &lsp.ServerInfo{
			Name:    "qml-language-server",
			Version: "0.1.0",
		},
	}, nil
}

func (h *Handler) Initialized(_ context.Context, _ *lsp.InitializedParams) error {
	return nil
}

func (h *Handler) Shutdown(_ context.Context) error {
	return nil
}

func (h *Handler) Exit(_ context.Context) error {
	return nil
}

func (h *Handler) DidOpenTextDocument(_ context.Context, params *lsp.DidOpenTextDocumentParams) error {
	h.documents[params.TextDocument.URI] = params.TextDocument.Text
	if h.parser != nil {
		h.parser.Parse(params.TextDocument.URI, params.TextDocument.Text)
	}
	return nil
}

func (h *Handler) DidChangeTextDocument(_ context.Context, params *lsp.DidChangeTextDocumentParams) error {
	for _, change := range params.ContentChanges {
		h.documents[params.TextDocument.URI] = change.Text
	}
	if h.parser != nil {
		h.parser.Parse(params.TextDocument.URI, h.documents[params.TextDocument.URI])
	}
	return nil
}

func (h *Handler) DidCloseTextDocument(_ context.Context, params *lsp.DidCloseTextDocumentParams) error {
	delete(h.documents, params.TextDocument.URI)
	if h.parser != nil {
		h.parser.Invalidate(params.TextDocument.URI)
	}
	return nil
}

func (h *Handler) DidSaveTextDocument(_ context.Context, params *lsp.DidSaveTextDocumentParams) error {
	if params.Text != nil {
		h.documents[params.TextDocument.URI] = *params.Text
	}
	if h.parser != nil {
		h.parser.Parse(params.TextDocument.URI, h.documents[params.TextDocument.URI])
	}
	return nil
}

func (h *Handler) DidChangeWatchedFiles(_ context.Context, params *lsp.DidChangeWatchedFilesParams) error {
	return nil
}

func boolPtr(b bool) *bool { return &b }
