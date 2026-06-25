/*
Copyright © 2026 54L1M
*/

// Package store handles persistence of snippets to a JSON file under the user's
// config directory (~/.config/snip/snippets.json).
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/54L1M/snip/internal/snippet"
)

const (
	dirName       = "snip"
	configName    = ".snip"
	snippetsName  = "snippets.json"
	dirPerm       = 0o755
	filePerm      = 0o600
)

// ConfigDir returns ~/.config/snip, creating nothing.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, ".config", dirName), nil
}

// ConfigFile returns the path to the viper settings file (~/.config/snip/.snip).
func ConfigFile() string {
	dir, err := ConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, configName)
}

// snippetsPath returns the path to the snippets JSON store.
func snippetsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, snippetsName), nil
}

// Load reads the snippet store from disk. A missing file yields an empty store
// (not an error), so the tool works before anything is saved.
func Load() (*snippet.Store, error) {
	path, err := snippetsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &snippet.Store{Snippets: map[string]*snippet.Snippet{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}

	var s snippet.Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}
	if s.Snippets == nil {
		s.Snippets = map[string]*snippet.Snippet{}
	}
	return &s, nil
}

// Save writes the store to disk atomically (temp file + rename) with 0600 perms.
func Save(s *snippet.Store) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}

	path, err := snippetsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode snippets: %w", err)
	}

	tmp, err := os.CreateTemp(dir, snippetsName+".*.tmp")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("could not write snippets: %w", err)
	}
	if err := tmp.Chmod(filePerm); err != nil {
		tmp.Close()
		return fmt.Errorf("could not set permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("could not save snippets: %w", err)
	}
	return nil
}
