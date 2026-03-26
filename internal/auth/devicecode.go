package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	envClientID = "EMAIL_LINEAR_CLIENT_ID"
	envTenantID = "EMAIL_LINEAR_TENANT_ID"
	scopes      = "Mail.Read offline_access"
)

// TokenResponse holds tokens returned by the OAuth token endpoint.
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired reports whether the access token has expired.
func (t TokenResponse) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// DeviceCodeResponse holds the response from the device code endpoint.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

// EnvConfig reads client ID and tenant ID from environment variables.
func EnvConfig() (clientID, tenantID string, err error) {
	clientID = os.Getenv(envClientID)
	if clientID == "" {
		return "", "", fmt.Errorf("environment variable %s is not set", envClientID)
	}
	tenantID = os.Getenv(envTenantID)
	if tenantID == "" {
		tenantID = "common"
	}
	return clientID, tenantID, nil
}

// URL functions are variables so tests can override them.
var deviceCodeURLFunc = func(tenantID string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/devicecode", tenantID)
}

var tokenURLFunc = func(tenantID string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
}

// RequestDeviceCode initiates the device code flow.
func RequestDeviceCode(ctx context.Context, httpClient *http.Client, clientID, tenantID string) (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {clientID},
		"scope":     {scopes},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceCodeURLFunc(tenantID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create device code request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed with status %d", resp.StatusCode)
	}

	var dcResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&dcResp); err != nil {
		return nil, fmt.Errorf("decode device code response: %w", err)
	}
	return &dcResp, nil
}

// PollForToken polls the token endpoint until the user approves or the context is cancelled.
func PollForToken(ctx context.Context, httpClient *http.Client, clientID, tenantID, deviceCode string, interval int) (*TokenResponse, error) {
	if interval < 1 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	data := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"client_id":   {clientID},
		"device_code": {deviceCode},
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			token, pending, err := tryTokenRequest(ctx, httpClient, tenantID, data)
			if err != nil {
				return nil, err
			}
			if pending {
				continue
			}
			return token, nil
		}
	}
}

func tryTokenRequest(ctx context.Context, httpClient *http.Client, tenantID string, data url.Values) (*TokenResponse, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURLFunc(tenantID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, false, fmt.Errorf("create token poll request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("token poll request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var tokenResp TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return nil, false, fmt.Errorf("decode token response: %w", err)
		}
		tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		return &tokenResp, false, nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, false, fmt.Errorf("decode error response: %w", err)
	}

	switch errResp.Error {
	case "authorization_pending", "slow_down":
		return nil, true, nil
	case "expired_token":
		return nil, false, fmt.Errorf("device code expired, please restart authentication")
	default:
		return nil, false, fmt.Errorf("token request error: %s", errResp.Error)
	}
}

// RefreshAccessToken uses a refresh token to obtain a new access token.
func RefreshAccessToken(ctx context.Context, httpClient *http.Client, clientID, tenantID, refreshToken string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
		"scope":         {scopes},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURLFunc(tenantID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (status %d), please re-authenticate with --auth-only", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}
	tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return &tokenResp, nil
}
