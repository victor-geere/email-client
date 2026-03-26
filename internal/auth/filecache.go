package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileCache stores tokens as a JSON file with restrictive permissions.
type FileCache struct {
	Path string
}

// NewFileCache creates a FileCache using the default path (~/.email-linearize/token.json).
func NewFileCache() (*FileCache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}
	return &FileCache{Path: filepath.Join(home, ".email-linearize", "token.json")}, nil
}

// Save writes the token to disk with permissions 0600.
func (f *FileCache) Save(token TokenResponse) error {
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create token cache directory: %w", err)
	}
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshal token for file cache: %w", err)
	}
	if err := os.WriteFile(f.Path, data, 0600); err != nil {
		return fmt.Errorf("write token cache file: %w", err)
	}
	return nil
}

// Load reads the token from disk.
func (f *FileCache) Load() (TokenResponse, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("read token cache file: %w", err)
	}
	var token TokenResponse
	if err := json.Unmarshal(data, &token); err != nil {
		return TokenResponse{}, fmt.Errorf("unmarshal token from file cache: %w", err)
	}
	return token, nil
}

// Clear removes the token cache file.
func (f *FileCache) Clear() error {
	if err := os.Remove(f.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove token cache file: %w", err)
	}
	return nil
}
