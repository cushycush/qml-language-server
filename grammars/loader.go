package qmlgrammars

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/gob"
	"sync"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammargen"
	"github.com/odvcencio/gotreesitter/grammars"
)

//go:embed qmljs.grammar.json
var grammarJSON []byte

//go:embed queries/highlights.scm
var highlightQuery string

//go:embed grammar_blobs/qmljs.bin
var qmljsBlob []byte

//go:embed grammar_blobs/qmljs.generated.bin
var qmljsGenerated []byte

var (
	qmlLang  *gotreesitter.Language
	langErr  error
	langOnce sync.Once
)

// QmljsLanguage returns the cached gotreesitter.Language for QML/JS.
//
// The parse tables are loaded from a gob blob that was generated ahead of
// time by `go run ./grammars/internal/gen`. The generator feeds the JSON
// grammar through grammargen — work that takes ~11s and we don't want to
// pay on every startup. Decoding the gob is a few ms.
//
// `qmljs.bin` (the upstream reference Language) is still used to adapt
// external-scanner symbol ordering: the generated tables and the hand-ported
// Go scanner have to agree on which token ID means which terminal, and the
// reference Language is the source of truth for that mapping.
//
// If the generated blob is missing (e.g. the generator hasn't been run after
// a grammar change), we fall back to regenerating at startup.
func QmljsLanguage() (*gotreesitter.Language, error) {
	langOnce.Do(func() {
		grammars.RegisterExternalScanner("qmljs", newQmljsScanner())

		grammars.RegisterExtension(grammars.ExtensionEntry{
			Name:              "qmljs",
			Extensions:        []string{".qml"},
			GenerateLanguage:  qmlGenerateLanguage,
			HighlightQuery:    highlightQuery,
			InheritHighlights: "javascript",
		})

		entry := grammars.DetectLanguageByName("qmljs")
		if entry != nil {
			qmlLang = entry.Language()
		} else {
			langErr = &LanguageNotFoundError{name: "qmljs"}
		}
	})
	return qmlLang, langErr
}

func qmlGenerateLanguage() (*gotreesitter.Language, error) {
	lang, err := loadGeneratedLanguage()
	if err != nil {
		// Fallback: regenerate from JSON. Slow (~11s), but keeps us functional
		// if the prebuilt blob is absent.
		g, err := grammargen.ImportGrammarJSON(grammarJSON)
		if err != nil {
			return nil, err
		}
		lang, err = grammargen.GenerateLanguage(g)
		if err != nil {
			return nil, err
		}
	}

	refLang, err := decodeLanguageBlob(qmljsBlob)
	if err != nil {
		return nil, err
	}

	adapted, ok := gotreesitter.AdaptExternalScannerByExternalOrder(refLang, lang)
	if !ok {
		return nil, &ScannerAdaptationError{}
	}

	lang.ExternalScanner = adapted
	return lang, nil
}

func loadGeneratedLanguage() (*gotreesitter.Language, error) {
	if len(qmljsGenerated) == 0 {
		return nil, &GeneratedBlobMissingError{}
	}
	gzr, err := gzip.NewReader(bytes.NewReader(qmljsGenerated))
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	var lang gotreesitter.Language
	if err := gob.NewDecoder(gzr).Decode(&lang); err != nil {
		return nil, err
	}
	return &lang, nil
}

func decodeLanguageBlob(data []byte) (*gotreesitter.Language, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	dec := gob.NewDecoder(gzr)
	var lang gotreesitter.Language
	if err := dec.Decode(&lang); err != nil {
		return nil, err
	}

	if lang.ExternalScanner == nil {
		lang.ExternalScanner = newQmljsScanner()
	}

	return &lang, nil
}

type LanguageNotFoundError struct {
	name string
}

func (e *LanguageNotFoundError) Error() string {
	return "language not found: " + e.name
}

type ScannerAdaptationError struct{}

func (e *ScannerAdaptationError) Error() string {
	return "failed to adapt external scanner for qmljs"
}

type GeneratedBlobMissingError struct{}

func (e *GeneratedBlobMissingError) Error() string {
	return "generated language blob missing: run `go run ./grammars/internal/gen`"
}
