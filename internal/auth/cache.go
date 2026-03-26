package auth

// TokenCache defines the interface for persisting OAuth tokens.
type TokenCache interface {
	Save(token TokenResponse) error
	Load() (TokenResponse, error)
	Clear() error
}
