package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// TokenSource provides valid access tokens, handling caching and refresh.
type TokenSource struct {
	ClientID   string
	TenantID   string
	Cache      TokenCache
	HTTPClient *http.Client
}

// NewTokenSource creates a TokenSource with the given configuration.
func NewTokenSource(clientID, tenantID string, cache TokenCache) *TokenSource {
	return &TokenSource{
		ClientID:   clientID,
		TenantID:   tenantID,
		Cache:      cache,
		HTTPClient: http.DefaultClient,
	}
}

// Token returns a valid access token. It loads from cache, refreshes if
// expired, or initiates a device code flow as a last resort.
func (ts *TokenSource) Token(ctx context.Context) (string, error) {
	cached, err := ts.Cache.Load()
	if err == nil && !cached.IsExpired() {
		return cached.AccessToken, nil
	}

	// Try refresh if we have a refresh token
	if err == nil && cached.RefreshToken != "" {
		refreshed, refreshErr := RefreshAccessToken(ctx, ts.HTTPClient, ts.ClientID, ts.TenantID, cached.RefreshToken)
		if refreshErr == nil {
			if saveErr := ts.Cache.Save(*refreshed); saveErr != nil {
				return "", fmt.Errorf("save refreshed token: %w", saveErr)
			}
			return refreshed.AccessToken, nil
		}
		// Refresh failed — fall through to device code flow
	}

	// Device code flow
	dcResp, err := RequestDeviceCode(ctx, ts.HTTPClient, ts.ClientID, ts.TenantID)
	if err != nil {
		return "", fmt.Errorf("request device code: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n%s\n\n", dcResp.Message)

	token, err := PollForToken(ctx, ts.HTTPClient, ts.ClientID, ts.TenantID, dcResp.DeviceCode, dcResp.Interval)
	if err != nil {
		return "", fmt.Errorf("poll for token: %w", err)
	}
	if saveErr := ts.Cache.Save(*token); saveErr != nil {
		return "", fmt.Errorf("save new token: %w", saveErr)
	}
	return token.AccessToken, nil
}
