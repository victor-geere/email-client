package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// memoryCache is a simple in-memory TokenCache for testing.
type memoryCache struct {
	token *TokenResponse
}

func (m *memoryCache) Save(token TokenResponse) error {
	m.token = &token
	return nil
}

func (m *memoryCache) Load() (TokenResponse, error) {
	if m.token == nil {
		return TokenResponse{}, fmt.Errorf("no token cached")
	}
	return *m.token, nil
}

func (m *memoryCache) Clear() error {
	m.token = nil
	return nil
}

func TestTokenSource_ReturnsCachedToken(t *testing.T) {
	cache := &memoryCache{
		token: &TokenResponse{
			AccessToken: "cached-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}
	ts := NewTokenSource("client-id", "tenant-id", cache)
	token, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cached-token" {
		t.Errorf("expected cached-token, got %s", token)
	}
}

func TestTokenSource_RefreshesExpiredToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"access_token":  "refreshed-token",
			"refresh_token": "new-refresh",
			"expires_in":    3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	origTokenURL := tokenURLFunc
	tokenURLFunc = func(tenantID string) string { return server.URL }
	defer func() { tokenURLFunc = origTokenURL }()

	cache := &memoryCache{
		token: &TokenResponse{
			AccessToken:  "expired-token",
			RefreshToken: "old-refresh",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		},
	}
	ts := NewTokenSource("client-id", "tenant-id", cache)
	token, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "refreshed-token" {
		t.Errorf("expected refreshed-token, got %s", token)
	}
}

func TestTokenSource_DeviceCodeFlowOnNoCache(t *testing.T) {
	step := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		r.ParseForm()
		grantType := r.FormValue("grant_type")

		if grantType == "" {
			// Device code request
			json.NewEncoder(w).Encode(map[string]interface{}{
				"device_code":      "test-device-code",
				"user_code":        "ABCD-1234",
				"verification_uri": "https://example.com/device",
				"expires_in":       900,
				"interval":         1,
				"message":          "Go to https://example.com/device and enter ABCD-1234",
			})
			return
		}

		// Token poll
		if step == 0 {
			step++
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new-token",
			"refresh_token": "new-refresh",
			"expires_in":    3600,
		})
	}))
	defer server.Close()

	origTokenURL := tokenURLFunc
	origDeviceCodeURL := deviceCodeURLFunc
	tokenURLFunc = func(tenantID string) string { return server.URL }
	deviceCodeURLFunc = func(tenantID string) string { return server.URL }
	defer func() {
		tokenURLFunc = origTokenURL
		deviceCodeURLFunc = origDeviceCodeURL
	}()

	cache := &memoryCache{}
	ts := NewTokenSource("client-id", "tenant-id", cache)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := ts.Token(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "new-token" {
		t.Errorf("expected new-token, got %s", token)
	}
	if cache.token == nil {
		t.Fatal("expected token to be cached")
	}
}

func TestRefreshAccessToken_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	origTokenURL := tokenURLFunc
	tokenURLFunc = func(tenantID string) string { return server.URL }
	defer func() { tokenURLFunc = origTokenURL }()

	_, err := RefreshAccessToken(context.Background(), server.Client(), "client-id", "tenant-id", "bad-refresh")
	if err == nil {
		t.Fatal("expected error on refresh failure")
	}
}
