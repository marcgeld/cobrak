package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sentinel errors for config path validation.
var (
	// ErrAbsolutePath is returned when a config override is an absolute path.
	ErrAbsolutePath = errors.New("config path must be a relative path, not absolute")
	// ErrPathTraversal is returned when a config override would escape ~/.cobrak.
	ErrPathTraversal = errors.New("config path must not escape the ~/.cobrak directory")
)

// ResolveConfigPath resolves the configuration file path using the following precedence:
//  1. flagPath (if non-empty): treated as a relative path under ~/.cobrak/
//  2. COBRAK_CONFIG environment variable (if set): treated as a relative path under ~/.cobrak/
//  3. default: ~/.cobrak/settings.toml
//
// Absolute paths and path traversal (e.g. "../x") are rejected with an error.
func ResolveConfigPath(flagPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}

	root := filepath.Join(home, ".cobrak")

	// 1. --config flag
	if flagPath != "" {
		return scopedConfigPath(root, flagPath)
	}

	// 2. COBRAK_CONFIG environment variable
	if env := os.Getenv("COBRAK_CONFIG"); env != "" {
		return scopedConfigPath(root, env)
	}

	// 3. Default path
	return filepath.Join(root, "settings.toml"), nil
}

// scopedConfigPath validates that rel is a safe relative path and returns the
// absolute path filepath.Join(root, filepath.Clean(rel)).
// It rejects absolute paths and any path that would escape root via traversal.
func scopedConfigPath(root, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", ErrAbsolutePath
	}

	clean := filepath.Clean(rel)

	// Reject paths that start with ".." (traversal attempt)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	full := filepath.Join(root, clean)

	// Double-check: the resolved path must remain under root
	rel2, err := filepath.Rel(root, full)
	if err != nil || rel2 == ".." || strings.HasPrefix(rel2, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	return full, nil
}
