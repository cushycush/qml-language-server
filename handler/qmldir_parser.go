package handler

import (
	"os"
	"path/filepath"
	"strings"
)

// ParseQMLDirFile reads and parses a qmldir file from disk.
func ParseQMLDirFile(path string) (*QMLDirModule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := ParseQMLDir(string(data))
	m.Dir = filepath.Dir(path)
	return m, nil
}

// ParseQMLDir parses the content of a qmldir file.
func ParseQMLDir(content string) *QMLDirModule {
	m := &QMLDirModule{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "module":
			if len(fields) >= 2 {
				m.Name = fields[1]
			}
		case "typeinfo":
			if len(fields) >= 2 {
				m.TypeInfo = fields[1]
			}
		case "depends":
			if len(fields) >= 2 {
				m.Depends = append(m.Depends, fields[1])
			}
		case "import":
			if len(fields) >= 2 {
				m.Imports = append(m.Imports, fields[1])
			}
		}
	}
	return m
}
