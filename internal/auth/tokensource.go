package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// TokenSource provides valid access tokens, handling caching and refresh.
type TokenSource struct {
	ClientID     string
	TenantID     string
	ClientSecret string
	Cache        TokenCache
	HTTPClient   *http.Client
}

// NewTokenSource creates a TokenSource with the given configuration.
func NewTokenSource(clientID, tenantID, clientSecret string, cache TokenCache) *TokenSource {
	return &TokenSource{
		ClientID:     clientID,
		TenantID:     tenantID,
		ClientSecret: clientSecret,
		Cache:        cache,
		HTTPClient:   http.DefaultClient,
	}
}

// Token returns a valid access token. It loads from cache, refreshes if
// expired, or initiates a device code flow for new sign-in.
func (ts *TokenSource) Token(ctx context.Context) (string, error) {
	cached, err := ts.Cache.Load()
	if err == nil && !cached.IsExpired() {
		return cached.AccessToken, nil
	}

	// Try refresh if we have a refresh token
	if err == nil && cached.RefreshToken != "" {
		refreshed, refreshErr := RefreshAccessToken(ctx, ts.HTTPClient, ts.ClientID, ts.TenantID, ts.ClientSecret, cached.RefreshToken)
		if refreshErr == nil {
			if saveErr := ts.Cache.Save(*refreshed); saveErr != nil {
				return "", fmt.Errorf("save refreshed token: %w", saveErr)
			}
			return refreshed.AccessToken, nil
		}
		// Refresh failed — inform user and fall through to device code flow
		fmt.Fprintf(os.Stderr, "Token refresh failed: %v\nStarting new sign-in...\n", refreshErr)
	}

	// Device code flow
	dcResp, dcErr := RequestDeviceCode(ctx, ts.HTTPClient, ts.ClientID, ts.TenantID)
	if dcErr != nil {
		return "", fmt.Errorf("request device code: %w", dcErr)
	}

	promptDeviceCode(dcResp)

	fmt.Fprintf(os.Stderr, "Waiting for sign-in (including any MFA steps)...\n")
	token, err := PollForToken(ctx, ts.HTTPClient, ts.ClientID, ts.TenantID, ts.ClientSecret, dcResp.DeviceCode, dcResp.Interval)
	if err != nil {
		return "", fmt.Errorf("poll for token: %w", err)
	}

	if saveErr := ts.Cache.Save(*token); saveErr != nil {
		return "", fmt.Errorf("save new token: %w", saveErr)
	}
	return token.AccessToken, nil
}
