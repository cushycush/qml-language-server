package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"

	"github.com/odvcencio/gotreesitter"
	"github.com/owenrumney/go-lsp/lsp"
	"github.com/owenrumney/go-lsp/server"
)

type Handler struct {
	logger *slog.Logger

	docMu     sync.RWMutex
	documents map[lsp.DocumentURI]string

	parser    *QMLParser
	server    *server.Server
	workspace *workspaceIndex
}

func New(logger *slog.Logger) *Handler {
	parser := NewQMLParser()
	if parser == nil && logger != nil {
		logger.Error("failed to load QML grammar; all language features will be disabled")
	}
	return &Handler{
		logger:    logger,
		documents: make(map[lsp.DocumentURI]string),
		parser:    parser,
		workspace: newWorkspaceIndex(),
	}
}

func (h *Handler) Serve(ctx context.Context) error {
	h.server = server.NewServer(h)
	// Override go-lsp's built-in shutdown handler: its default returns untyped
	// nil, which makes the response omit `result` entirely, and Neovim rejects
	// that as INVALID_SERVER_MESSAGE. LSP requires `result: null` explicitly.
	h.server.HandleMethod("shutdown", func(ctx context.Context, _ json.RawMessage) (any, error) {
		_ = h.Shutdown(ctx)
		return json.RawMessage("null"), nil
	})
	// go-lsp's dispatcher ignores errors from notification handlers, so the
	// library's built-in `exit` handler never actually terminates. Without
	// this override the process only exits on stdin EOF, causing the editor
	// to block on its shutdown timeout.
	h.server.HandleNotification("exit", func(context.Context, json.RawMessage) error {
		os.Exit(0)
		return nil
	})
	return h.server.Run(ctx, server.RunStdio())
}

// getDocument returns the current text for uri. Returns "", false when the
// document isn't open.
func (h *Handler) getDocument(uri lsp.DocumentURI) (string, bool) {
	h.docMu.RLock()
	defer h.docMu.RUnlock()
	doc, ok := h.documents[uri]
	return doc, ok
}

func (h *Handler) setDocument(uri lsp.DocumentURI, text string) {
	h.docMu.Lock()
	h.documents[uri] = text
	h.docMu.Unlock()
}

func (h *Handler) deleteDocument(uri lsp.DocumentURI) {
	h.docMu.Lock()
	delete(h.documents, uri)
	h.docMu.Unlock()
}

func (h *Handler) Initialize(_ context.Context, params *lsp.InitializeParams) (*lsp.InitializeResult, error) {
	roots := workspaceRootsFromInitialize(params)
	if h.workspace != nil {
		h.workspace.setRoots(roots)
		go h.workspace.scan()
	}
	go DiscoverAndRegisterQMLTypes(h.logger, roots)
	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptions{
				OpenClose: boolPtr(true),
				Change:    lsp.SyncFull,
				Save:      &lsp.SaveOptions{IncludeText: boolPtr(true)},
			},
			HoverProvider: boolPtr(true),
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", ":", "<", "\"", "/"},
			},
			DefinitionProvider:        boolPtr(true),
			ReferencesProvider:        boolPtr(true),
			DocumentSymbolProvider:    boolPtr(true),
			DocumentHighlightProvider: boolPtr(true),
			SignatureHelpProvider: &lsp.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
			CodeActionProvider: &lsp.CodeActionOptions{},
			RenameProvider: &lsp.RenameOptions{
				PrepareProvider: boolPtr(true),
			},
			DiagnosticProvider: &lsp.DiagnosticOptions{},
			InlayHintProvider:  &lsp.InlayHintOptions{},
			SemanticTokensProvider: &lsp.SemanticTokensOptions{
				Legend: SemanticTokensLegend(),
				Full:   &lsp.SemanticTokensFull{},
			},
			WorkspaceSymbolProvider:    boolPtr(true),
			DocumentFormattingProvider: boolPtr(true),
		},
		ServerInfo: &lsp.ServerInfo{
			Name:    "qml-language-server",
			Version: "0.1.0",
		},
	}, nil
}

func (h *Handler) Initialized(_ context.Context, _ *lsp.InitializedParams) error { return nil }
func (h *Handler) Shutdown(_ context.Context) error                              { return nil }
func (h *Handler) Exit(_ context.Context) error                                  { return nil }

func (h *Handler) publishDiagnostics(uri lsp.DocumentURI, diagnostics []lsp.Diagnostic) {
	if h.server == nil || h.server.Client == nil {
		return
	}
	if diagnostics == nil {
		diagnostics = []lsp.Diagnostic{}
	}
	_ = h.server.Client.PublishDiagnostics(context.Background(), &lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func (h *Handler) reparseAndPublish(uri lsp.DocumentURI, text string) {
	if h.parser == nil {
		return
	}
	h.parser.Parse(uri, text)
	h.publishDiagnostics(uri, h.getDiagnostics(uri))
}

func (h *Handler) DidOpen(_ context.Context, params *lsp.DidOpenTextDocumentParams) error {
	h.setDocument(params.TextDocument.URI, params.TextDocument.Text)
	h.reparseAndPublish(params.TextDocument.URI, params.TextDocument.Text)
	if h.workspace != nil {
		h.workspace.registerURI(params.TextDocument.URI)
	}
	return nil
}

func (h *Handler) DidChange(_ context.Context, params *lsp.DidChangeTextDocumentParams) error {
	// With Change: SyncFull the last change carries the full document. Earlier
	// entries are dropped in the same spirit as the previous code.
	if len(params.ContentChanges) == 0 {
		return nil
	}
	text := params.ContentChanges[len(params.ContentChanges)-1].Text
	h.setDocument(params.TextDocument.URI, text)
	h.reparseAndPublish(params.TextDocument.URI, text)
	return nil
}

func (h *Handler) DidClose(_ context.Context, params *lsp.DidCloseTextDocumentParams) error {
	h.deleteDocument(params.TextDocument.URI)
	if h.parser != nil {
		h.parser.Invalidate(params.TextDocument.URI)
	}
	h.publishDiagnostics(params.TextDocument.URI, nil)
	return nil
}

func (h *Handler) DidSave(_ context.Context, params *lsp.DidSaveTextDocumentParams) error {
	if params.Text == nil {
		return nil
	}
	h.setDocument(params.TextDocument.URI, *params.Text)
	h.reparseAndPublish(params.TextDocument.URI, *params.Text)
	return nil
}

func (h *Handler) DidChangeWatchedFiles(_ context.Context, _ *lsp.DidChangeWatchedFilesParams) error {
	if h.workspace != nil {
		go h.workspace.scan()
	}
	return nil
}

func (h *Handler) DocumentHighlight(_ context.Context, params *lsp.DocumentHighlightParams) ([]lsp.DocumentHighlight, error) {
	uri := params.TextDocument.URI
	doc, ok := h.getDocument(uri)
	if !ok || h.parser == nil {
		return nil, nil
	}
	tree := h.parser.GetTree(uri)
	if tree == nil {
		return nil, nil
	}
	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

	lang := h.parser.Language()
	content := []byte(doc)
	node := h.parser.GetNodeAt(uri, params.Position, content)
	if node == nil || node.Type(lang) != "identifier" {
		return nil, nil
	}
	target := string(content[node.StartByte():node.EndByte()])

	var highlights []lsp.DocumentHighlight
	walkTree(root, func(n *gotreesitter.Node) bool {
		if n.Type(lang) == "identifier" && string(content[n.StartByte():n.EndByte()]) == target {
			highlights = append(highlights, lsp.DocumentHighlight{Range: nodeRange(content, n)})
		}
		return true
	})
	return highlights, nil
}

func (h *Handler) DocumentDiagnostic(_ context.Context, params *lsp.DocumentDiagnosticParams) (any, error) {
	diagnostics := h.getDiagnostics(params.TextDocument.URI)
	if diagnostics == nil {
		diagnostics = []lsp.Diagnostic{}
	}
	return lsp.FullDocumentDiagnosticReport{Items: diagnostics}, nil
}

func (h *Handler) getDiagnostics(uri lsp.DocumentURI) []lsp.Diagnostic {
	if h.parser == nil {
		return nil
	}
	tree := h.parser.GetTree(uri)
	if tree == nil {
		return nil
	}
	doc, ok := h.getDocument(uri)
	if !ok {
		return nil
	}
	var diagnostics []lsp.Diagnostic
	collectDiagnostics(tree.RootNode(), h.parser.Language(), []byte(doc), &diagnostics)
	return diagnostics
}

func boolPtr(b bool) *bool { return &b }
