package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func requestedNames(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("file name is required")
	}
	if strings.EqualFold(raw, "all") {
		return []string{"all"}, nil
	}
	parts := strings.Split(raw, ",")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			return nil, errors.New("empty file name")
		}
		if err := validFilename(name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func validFilename(name string) error {
	if name == "all" || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) || name == "." || name == ".." {
		return errors.New("invalid file name")
	}
	return nil
}

func directory(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("directory path is required")
	}
	path, err := filepath.Abs(filepath.Clean(strings.TrimSpace(raw)))
	if err != nil {
		return "", err
	}
	if !isDir(path) {
		return "", fmt.Errorf("directory does not exist: %s", path)
	}
	return path, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isBackupFile(entry os.DirEntry) bool {
	return entry.Type().IsRegular() && entry.Name() != ".DS_Store" && !strings.HasPrefix(entry.Name(), "._")
}
