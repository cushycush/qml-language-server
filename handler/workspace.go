package handler

import (
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/owenrumney/go-lsp/lsp"
)

// workspaceIndex tracks QML files discovered under the workspace roots so we
// can offer user-defined components (e.g. `ShellRoot.qml`, project widgets) as
// completions. Convention: top-level object in `Foo.qml` is named `Foo`.
type workspaceIndex struct {
	mu     sync.RWMutex
	roots  []string
	byName map[string]workspaceComponent
}

type workspaceComponent struct {
	Name string
	Path string
	URI  lsp.DocumentURI
}

func newWorkspaceIndex() *workspaceIndex {
	return &workspaceIndex{byName: map[string]workspaceComponent{}}
}

func (w *workspaceIndex) setRoots(roots []string) {
	w.mu.Lock()
	w.roots = append(w.roots[:0], roots...)
	w.mu.Unlock()
}

// scan walks every root looking for *.qml files whose basename starts with an
// uppercase letter. Each such file registers a component named after the
// basename. Skips common vendor/build directories.
func (w *workspaceIndex) scan() {
	w.mu.Lock()
	roots := append([]string{}, w.roots...)
	w.mu.Unlock()

	found := map[string]workspaceComponent{}
	for _, root := range roots {
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if shouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.EqualFold(filepath.Ext(path), ".qml") {
				return nil
			}
			base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			if base == "" || !startsUpper(base) {
				return nil
			}
			if _, exists := found[base]; exists {
				return nil
			}
			found[base] = workspaceComponent{
				Name: base,
				Path: path,
				URI:  pathToURI(path),
			}
			return nil
		})
	}

	w.mu.Lock()
	w.byName = found
	w.mu.Unlock()
	for _, c := range found {
		publishWorkspaceSymbol(c)
	}
}

// publishWorkspaceSymbol mirrors a workspace component into the shared symbol
// registry so hover and completion-resolve can look it up by label.
func publishWorkspaceSymbol(c workspaceComponent) {
	registerSymbols(QMLSymbol{
		Label:       c.Name,
		Kind:        lsp.CompletionItemKindClass,
		Detail:      "workspace component — " + filepath.Base(c.Path),
		Signature:   c.Name + " { ... }",
		Description: "Defined in `" + c.Path + "`.",
		Category:    "workspace",
	})
	recordWorkspaceURI(c.Name, c.URI)
}

// workspaceURIs maps component name → file URI so go-to-definition can jump
// cross-file without coupling definition.go to the Handler. Populated every
// time a workspace component is published.
var (
	workspaceURIsMu sync.RWMutex
	workspaceURIs   = map[string]lsp.DocumentURI{}
)

func recordWorkspaceURI(name string, uri lsp.DocumentURI) {
	if name == "" || uri == "" {
		return
	}
	workspaceURIsMu.Lock()
	workspaceURIs[name] = uri
	workspaceURIsMu.Unlock()
}

// LookupWorkspaceURI returns the file URI for a workspace component, or "".
func LookupWorkspaceURI(name string) lsp.DocumentURI {
	workspaceURIsMu.RLock()
	defer workspaceURIsMu.RUnlock()
	return workspaceURIs[name]
}

// registerURI indexes a single QML document URI, used when a file is opened
// that lives outside the initial scan (e.g. an editor opens a sibling dir).
func (w *workspaceIndex) registerURI(uri lsp.DocumentURI) {
	path := uriToPath(uri)
	if path == "" || !strings.EqualFold(filepath.Ext(path), ".qml") {
		return
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if base == "" || !startsUpper(base) {
		return
	}
	comp := workspaceComponent{Name: base, Path: path, URI: uri}
	w.mu.Lock()
	w.byName[base] = comp
	w.mu.Unlock()
	publishWorkspaceSymbol(comp)
}

func (w *workspaceIndex) all() []workspaceComponent {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]workspaceComponent, 0, len(w.byName))
	for _, c := range w.byName {
		out = append(out, c)
	}
	return out
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", "node_modules", "build", "build-debug", "build-release",
		".cache", "dist", "out", "target", ".venv", "venv", ".idea", ".vscode":
		return true
	}
	return strings.HasPrefix(name, ".") && name != "."
}

func startsUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

func pathToURI(path string) lsp.DocumentURI {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := &url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return lsp.DocumentURI(u.String())
}

func uriToPath(uri lsp.DocumentURI) string {
	u, err := url.Parse(string(uri))
	if err != nil || u.Scheme != "file" {
		return ""
	}
	return u.Path
}

// workspaceCompletions returns CompletionItems for every user-defined QML
// component we've indexed. Each item carries a lightweight description so the
// doc pane shows the file path it came from.
func (h *Handler) workspaceCompletions() []lsp.CompletionItem {
	if h.workspace == nil {
		return nil
	}
	cs := h.workspace.all()
	if len(cs) == 0 {
		return nil
	}
	h.workspace.mu.RLock()
	roots := append([]string{}, h.workspace.roots...)
	h.workspace.mu.RUnlock()

	kind := lsp.CompletionItemKindClass
	items := make([]lsp.CompletionItem, 0, len(cs))
	for _, c := range cs {
		rel := c.Path
		for _, root := range roots {
			if r, err := filepath.Rel(root, c.Path); err == nil && !strings.HasPrefix(r, "..") {
				rel = r
				break
			}
		}
		items = append(items, lsp.CompletionItem{
			Label:  c.Name,
			Kind:   &kind,
			Detail: "workspace component — " + rel,
			Documentation: &lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: "**" + c.Name + "** — workspace component\n\nDefined in `" + rel + "`.",
			},
		})
	}
	return items
}

// workspaceRootsFromInitialize extracts root directories from the Initialize
// params, preferring WorkspaceFolders over the legacy RootURI/RootPath.
func workspaceRootsFromInitialize(params *lsp.InitializeParams) []string {
	if params == nil {
		return nil
	}
	var roots []string
	for _, wf := range params.WorkspaceFolders {
		if p := uriToPath(wf.URI); p != "" {
			roots = append(roots, p)
		}
	}
	if len(roots) == 0 && params.RootURI != nil {
		if p := uriToPath(*params.RootURI); p != "" {
			roots = append(roots, p)
		}
	}
	if len(roots) == 0 && params.RootPath != nil && *params.RootPath != "" {
		roots = append(roots, *params.RootPath)
	}
	if len(roots) == 0 {
		if cwd, err := os.Getwd(); err == nil {
			roots = append(roots, cwd)
		}
	}
	return roots
}
