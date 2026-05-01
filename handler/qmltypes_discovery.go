package handler

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/owenrumney/go-lsp/lsp"
)

// DiscoverAndRegisterQMLTypes walks known Qt installation paths, parses
// every qmldir/qmltypes pair it finds, and registers the types into the
// global symbol registry. Hard-coded entries are never removed — qmltypes
// data augments them so users never lose what they had.
//
// If workspaceRoots is non-empty, the function also looks for a .qmlls.ini
// file (the same format Qt's qmlls reads) and adds its buildDir and
// importPaths to the search.
func DiscoverAndRegisterQMLTypes(logger *slog.Logger, workspaceRoots []string) {
	paths := qmlImportPaths()

	// Merge paths from .qmlls.ini if present.
	if cfg := FindAndParseQMLLSIni(workspaceRoots); cfg != nil {
		if logger != nil {
			logger.Info("loaded .qmlls.ini", "buildDir", cfg.BuildDir, "importPaths", cfg.ImportPaths)
		}
		if cfg.BuildDir != "" {
			paths = appendUnique(paths, cfg.BuildDir)
		}
		for _, p := range cfg.ImportPaths {
			paths = appendUnique(paths, p)
		}
	}

	if len(paths) == 0 {
		if logger != nil {
			logger.Info("no QML import paths found; skipping qmltypes discovery")
		}
		return
	}

	var allModules []discoveredModule
	for _, root := range paths {
		allModules = append(allModules, discoverModules(root)...)
	}

	if logger != nil {
		logger.Info("discovered QML modules", "count", len(allModules))
	}

	for _, dm := range allModules {
		mod, err := ParseQMLTypesFile(dm.qmltypesPath)
		if err != nil {
			if logger != nil {
				logger.Warn("failed to parse qmltypes", "path", dm.qmltypesPath, "err", err)
			}
			continue
		}
		registerQMLTypesModule(mod, dm.moduleName)
	}

	// Build inheritance chains from the prototype index so
	// typePropertyCompletions walks inherited properties.
	buildInheritanceChains()

	if logger != nil {
		logger.Info("built inheritance chains", "types", len(baseTypes))
	}
}

func appendUnique(paths []string, p string) []string {
	p = filepath.Clean(p)
	for _, existing := range paths {
		if existing == p {
			return paths
		}
	}
	if info, err := os.Stat(p); err == nil && info.IsDir() {
		paths = append(paths, p)
	}
	return paths
}

type discoveredModule struct {
	moduleName    string
	qmltypesPath  string
	qmldirPath    string
}

// moduleQMLDirs maps module name to the absolute path of the qmldir file that
// declared it. Populated by DiscoverAndRegisterQMLTypes and used by document
// links to resolve `import Foo` targets. First writer wins so Qt6 takes
// precedence over Qt5 when both are installed.
var (
	moduleDirsMu sync.RWMutex
	moduleQMLDirs = map[string]string{}
)

func recordModuleQMLDir(name, path string) {
	if name == "" || path == "" {
		return
	}
	moduleDirsMu.Lock()
	if _, ok := moduleQMLDirs[name]; !ok {
		moduleQMLDirs[name] = path
	}
	moduleDirsMu.Unlock()
}

// LookupModuleQMLDir returns the qmldir path registered for a module, or "".
func LookupModuleQMLDir(name string) string {
	moduleDirsMu.RLock()
	defer moduleDirsMu.RUnlock()
	return moduleQMLDirs[name]
}

// qmlImportPaths returns directories to scan for QML modules.
func qmlImportPaths() []string {
	var paths []string
	seen := map[string]bool{}
	add := func(p string) {
		p = filepath.Clean(p)
		if !seen[p] {
			if info, err := os.Stat(p); err == nil && info.IsDir() {
				seen[p] = true
				paths = append(paths, p)
			}
		}
	}

	// Qt6, then Qt5 — first writer wins for duplicate exports.
	add("/usr/lib/qt6/qml")
	add("/usr/lib/qt/qml")
	add("/usr/lib64/qt6/qml")
	add("/usr/lib64/qt/qml")
	add("/usr/local/lib/qt6/qml")

	if p := os.Getenv("QML_IMPORT_PATH"); p != "" {
		for _, dir := range filepath.SplitList(p) {
			add(dir)
		}
	}
	if p := os.Getenv("QML2_IMPORT_PATH"); p != "" {
		for _, dir := range filepath.SplitList(p) {
			add(dir)
		}
	}
	return paths
}

// discoverModules walks a root directory for qmldir files and returns each
// module that has a typeinfo reference.
func discoverModules(root string) []discoveredModule {
	var modules []discoveredModule
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "qmldir" {
			return nil
		}
		qmld, err := ParseQMLDirFile(path)
		if err != nil || qmld.TypeInfo == "" {
			return nil
		}
		typesPath := filepath.Join(qmld.Dir, qmld.TypeInfo)
		if _, err := os.Stat(typesPath); err != nil {
			return nil
		}
		modules = append(modules, discoveredModule{
			moduleName:   qmld.Name,
			qmltypesPath: typesPath,
			qmldirPath:   path,
		})
		recordModuleQMLDir(qmld.Name, path)
		return nil
	})
	return modules
}

// prototypeIndex maps C++ class names to their parsed component so we can
// walk inheritance chains to resolve properties.
var (
	protoMu    sync.RWMutex
	protoIndex = map[string]*QMLTypesComponent{}
)

func indexPrototype(comp *QMLTypesComponent) {
	protoMu.Lock()
	protoIndex[comp.Name] = comp
	protoMu.Unlock()
}


// buildInheritanceChains walks the prototype index and populates baseTypes
// for every exported QML type so that typePropertyCompletions can walk
// inherited properties. Called once after all modules are registered.
func buildInheritanceChains() {
	protoMu.RLock()
	defer protoMu.RUnlock()

	for _, comp := range protoIndex {
		qmlName := comp.ExportedName()
		if qmlName == "" {
			continue
		}
		// Already has a hand-coded chain — don't overwrite.
		if _, exists := baseTypes[qmlName]; exists {
			continue
		}
		chain := resolvePrototypeChain(comp.Prototype)
		if len(chain) > 0 {
			baseTypes[qmlName] = chain
		}
	}
}

// resolvePrototypeChain walks the prototype field up the C++ inheritance
// hierarchy and returns QML-visible type names. Stops at QObject or after
// 32 hops (safety).
func resolvePrototypeChain(cppPrototype string) []string {
	var chain []string
	visited := map[string]bool{}
	current := cppPrototype
	for i := 0; i < 32 && current != ""; i++ {
		if visited[current] {
			break
		}
		visited[current] = true
		parent, ok := protoIndex[current]
		if !ok {
			break
		}
		if name := parent.ExportedName(); name != "" {
			chain = append(chain, name)
		}
		current = parent.Prototype
	}
	return chain
}

// registerQMLTypesModule converts parsed qmltypes components into QMLSymbol
// entries and writes them into the global symbol registry. Existing entries
// (hard-coded or from earlier modules) are not overwritten — only new
// symbols or richer property sets are added.
func registerQMLTypesModule(mod *QMLTypesModule, fallbackModule string) {
	// First pass: index all components by C++ name for prototype resolution.
	for i := range mod.Components {
		indexPrototype(&mod.Components[i])
	}

	// Second pass: register types that have QML exports.
	for i := range mod.Components {
		comp := &mod.Components[i]
		qmlName := comp.ExportedName()
		if qmlName == "" {
			continue
		}
		// Only register types that are creatable or singletons — skip internal
		// value types and sequence wrappers.
		if comp.AccessSemantics == "sequence" {
			continue
		}

		module := comp.ExportedModule()
		if module == "" {
			module = fallbackModule
		}

		// Register the type itself if not already known.
		if _, exists := lookupSymbol(qmlName); !exists {
			sym := QMLSymbol{
				Label:    qmlName,
				Kind:     lsp.CompletionItemKindClass,
				Detail:   module + " — " + qmlName,
				Module:   module,
				Category: "type",
			}
			if comp.IsSingleton {
				sym.Kind = lsp.CompletionItemKindModule
				sym.Category = "js"
			}
			registerSymbols(sym)
		}

		// Register the import module if not already known.
		if module != "" {
			if _, exists := lookupSymbol(module); !exists {
				registerSymbols(QMLSymbol{
					Label:     module,
					Kind:      lsp.CompletionItemKindModule,
					Detail:    module + " module",
					Signature: "import " + module,
					Module:    module,
					Category:  "import",
				})
			}
		}

		// Register type-specific properties from qmltypes.
		registerQMLTypesProperties(comp, qmlName)

		// Register method signatures for signature help.
		registerQMLTypesSignatures(comp, qmlName)
	}
}

// registerQMLTypesProperties registers properties, signals, methods, and
// enums from a parsed component as type-specific completions.
func registerQMLTypesProperties(comp *QMLTypesComponent, qmlName string) {
	for _, prop := range comp.Properties {
		label := prop.Name
		qmlType := cppTypeToQML(prop.Type)
		detail := qmlType + " — " + label + " (" + qmlName + ")"
		sig := qmlName + "." + label + ": " + qmlType
		sym := QMLSymbol{
			Label:      label,
			Kind:       lsp.CompletionItemKindProperty,
			Detail:     detail,
			Signature:  sig,
			Module:     "",
			Category:   "property",
			InsertText: label + ": ",
		}
		addTypeProperty(qmlName, sym)
	}

	for _, sig := range comp.Signals {
		handlerName := "on" + capitalize(sig.Name)
		detail := "signal — " + sig.Name + " (" + qmlName + ")"
		sym := QMLSymbol{
			Label:      handlerName,
			Kind:       lsp.CompletionItemKindEvent,
			Detail:     detail,
			Signature:  formatSignalSignature(sig),
			Module:     "",
			Category:   "property",
			InsertText: handlerName + ": ",
		}
		addTypeProperty(qmlName, sym)
	}

	for _, m := range comp.Methods {
		if m.Name == "" {
			continue
		}
		detail := "method — " + m.Name + " (" + qmlName + ")"
		sym := QMLSymbol{
			Label:     m.Name,
			Kind:      lsp.CompletionItemKindMethod,
			Detail:    detail,
			Signature: formatMethodSignature(m),
			Module:    "",
			Category:  "property",
		}
		addTypeProperty(qmlName, sym)
	}

	for _, e := range comp.Enums {
		if e.Name == "" {
			continue
		}
		detail := "enum — " + e.Name + " (" + qmlName + ")"
		desc := strings.Join(e.Values, ", ")
		if len(desc) > 120 {
			desc = desc[:120] + "…"
		}
		sym := QMLSymbol{
			Label:       qmlName + "." + e.Name,
			Kind:        lsp.CompletionItemKindEnum,
			Detail:      detail,
			Description: desc,
			Module:      "",
			Category:    "type",
		}
		if _, exists := lookupSymbol(sym.Label); !exists {
			registerSymbols(sym)
		}
	}
}

// registerQMLTypesSignatures registers method signatures from a parsed
// component into the functionSignatures map used by signature help. Methods
// are registered under both "method" (bare) and "TypeName.method" (dotted)
// keys. Existing hand-coded entries are never overwritten.
func registerQMLTypesSignatures(comp *QMLTypesComponent, qmlName string) {
	for _, m := range comp.Methods {
		if m.Name == "" {
			continue
		}
		sig := buildSignatureInfo(m)

		// "TypeName.method" for dotted calls (e.g. Qt.binding). Bare "method"
		// for unqualified calls. Hand-coded entries and earlier types win.
		registerSignatureIfAbsent(qmlName+"."+m.Name, sig)
		registerSignatureIfAbsent(m.Name, sig)
	}
}

func buildSignatureInfo(m QMLTypesMethod) lsp.SignatureInformation {
	var params []lsp.ParameterInformation
	var paramLabels []string
	for _, p := range m.Parameters {
		qmlType := cppTypeToQML(p.Type)
		label := p.Name + ": " + qmlType
		paramLabels = append(paramLabels, label)
		params = append(params, lsp.ParameterInformation{
			Label:         label,
			Documentation: plainText(qmlType),
		})
	}
	ret := cppTypeToQML(m.ReturnType)
	sigLabel := m.Name + "(" + strings.Join(paramLabels, ", ") + ")"
	if ret != "void" {
		sigLabel += ": " + ret
	}
	return lsp.SignatureInformation{
		Label:      sigLabel,
		Parameters: params,
	}
}

// addTypeProperty adds a property to the per-type catalog (typeProperties map)
// without overwriting existing entries.
func addTypeProperty(typeName string, sym QMLSymbol) {
	existing := typeProperties[typeName]
	for _, e := range existing {
		if e.Label == sym.Label {
			return
		}
	}
	typeProperties[typeName] = append(typeProperties[typeName], sym)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatSignalSignature(sig QMLTypesSignal) string {
	var params []string
	for _, p := range sig.Parameters {
		params = append(params, fmt.Sprintf("%s: %s", p.Name, cppTypeToQML(p.Type)))
	}
	return "signal " + sig.Name + "(" + strings.Join(params, ", ") + ")"
}

func formatMethodSignature(m QMLTypesMethod) string {
	var params []string
	for _, p := range m.Parameters {
		params = append(params, fmt.Sprintf("%s: %s", p.Name, cppTypeToQML(p.Type)))
	}
	ret := cppTypeToQML(m.ReturnType)
	if ret == "void" {
		return "function " + m.Name + "(" + strings.Join(params, ", ") + ")"
	}
	return "function " + m.Name + "(" + strings.Join(params, ", ") + "): " + ret
}
