package handler

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// QMLLSConfig represents settings parsed from a .qmlls.ini file. This is the
// same configuration file format that Qt's qmlls reads, so projects that
// already generate one (e.g. Quickshell) work out of the box.
type QMLLSConfig struct {
	BuildDir    string
	ImportPaths []string
}

// FindAndParseQMLLSIni searches workspace roots for a .qmlls.ini file and
// parses the first one found. Returns nil if none exists.
func FindAndParseQMLLSIni(roots []string) *QMLLSConfig {
	for _, root := range roots {
		path := filepath.Join(root, ".qmlls.ini")
		cfg, err := ParseQMLLSIni(path)
		if err == nil {
			return cfg
		}
	}
	return nil
}

// ParseQMLLSIni parses a .qmlls.ini file. The format is a simple INI file
// with a [General] section containing keys like buildDir and importPaths.
func ParseQMLLSIni(path string) (*QMLLSConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	cfg := &QMLLSConfig{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"")

		switch key {
		case "buildDir":
			cfg.BuildDir = value
		case "importPaths":
			for _, p := range strings.Split(value, ":") {
				p = strings.TrimSpace(p)
				if p != "" {
					cfg.ImportPaths = append(cfg.ImportPaths, p)
				}
			}
		}
	}
	return cfg, scanner.Err()
}
