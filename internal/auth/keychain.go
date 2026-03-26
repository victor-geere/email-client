package auth

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keychainService = "email-linearize"
	keychainUser    = "default"
)

// KeychainCache stores tokens in the OS keychain.
type KeychainCache struct{}

// Save stores the token in the OS keychain.
func (k *KeychainCache) Save(token TokenResponse) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshal token for keychain: %w", err)
	}
	if err := keyring.Set(keychainService, keychainUser, string(data)); err != nil {
		return fmt.Errorf("save to keychain: %w", err)
	}
	return nil
}

// Load retrieves the token from the OS keychain.
func (k *KeychainCache) Load() (TokenResponse, error) {
	data, err := keyring.Get(keychainService, keychainUser)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("load from keychain: %w", err)
	}
	var token TokenResponse
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return TokenResponse{}, fmt.Errorf("unmarshal token from keychain: %w", err)
	}
	return token, nil
}

// Clear removes the token from the OS keychain.
func (k *KeychainCache) Clear() error {
	if err := keyring.Delete(keychainService, keychainUser); err != nil {
		return fmt.Errorf("clear keychain: %w", err)
	}
	return nil
}
