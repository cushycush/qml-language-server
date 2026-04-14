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

var (
	qmlLang  *gotreesitter.Language
	langErr  error
	langOnce sync.Once
)

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
	g, err := grammargen.ImportGrammarJSON(grammarJSON)
	if err != nil {
		return nil, err
	}

	lang, err := grammargen.GenerateLanguage(g)
	if err != nil {
		return nil, err
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
