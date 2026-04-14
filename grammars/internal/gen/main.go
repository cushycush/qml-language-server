// Command gen runs grammargen once and writes the resulting parse tables to
// grammar_blobs/qmljs.generated.bin as a gzipped gob. This is ~11s of work
// that we do not want to repeat on every process start; the loader embeds the
// output and decodes it in milliseconds.
//
// Usage: go run ./grammars/internal/gen
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/odvcencio/gotreesitter/grammargen"
)

func main() {
	src, err := os.ReadFile("grammars/qmljs.grammar.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read grammar json:", err)
		os.Exit(1)
	}
	g, err := grammargen.ImportGrammarJSON(src)
	if err != nil {
		fmt.Fprintln(os.Stderr, "import grammar json:", err)
		os.Exit(1)
	}
	lang, err := grammargen.GenerateLanguage(g)
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate language:", err)
		os.Exit(1)
	}
	lang.ExternalScanner = nil // scanner gets re-attached at load time

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if err := gob.NewEncoder(gz).Encode(lang); err != nil {
		fmt.Fprintln(os.Stderr, "gob encode:", err)
		os.Exit(1)
	}
	if err := gz.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "gzip close:", err)
		os.Exit(1)
	}

	out := "grammars/grammar_blobs/qmljs.generated.bin"
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d bytes)\n", out, buf.Len())
}
