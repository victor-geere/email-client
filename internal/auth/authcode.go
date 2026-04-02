package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var authorizeURLFunc = func(tenantID string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", tenantID)
}

// generateCodeVerifier creates a random PKCE code verifier (43-128 chars from unreserved set).
func generateCodeVerifier() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// codeChallenge computes the S256 code challenge from the verifier.
func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// AuthCodeResult is sent over a channel when the callback is received.
type AuthCodeResult struct {
	Code string
	Err  error
}

// AuthCodeFlow performs the OAuth 2.0 authorization code flow with PKCE.
// It starts a local HTTP server, opens the browser for sign-in, and exchanges
// the returned code for tokens. This works with MFA and conditional access
// policies that block device code flow.
func AuthCodeFlow(ctx context.Context, httpClient *http.Client, clientID, tenantID, clientSecret string) (*TokenResponse, error) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	challenge := codeChallenge(verifier)

	// Find an available port and start the callback server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start callback listener: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	resultCh := make(chan AuthCodeResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errMsg := r.URL.Query().Get("error")
		errDesc := r.URL.Query().Get("error_description")

		if errMsg != "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>%s</p><p>You can close this tab.</p></body></html>", errDesc)
			resultCh <- AuthCodeResult{Err: fmt.Errorf("%s — %s", errMsg, errDesc)}
			return
		}
		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body><h2>Authentication failed</h2><p>No authorization code received.</p><p>You can close this tab.</p></body></html>")
			resultCh <- AuthCodeResult{Err: fmt.Errorf("no authorization code in callback")}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
		resultCh <- AuthCodeResult{Code: code}
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			resultCh <- AuthCodeResult{Err: fmt.Errorf("callback server: %w", err)}
		}
	}()
	defer server.Shutdown(context.Background())

	// Build authorization URL and open browser.
	params := url.Values{
		"client_id":             {clientID},
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"scope":                 {scopes},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"prompt":                {"select_account"},
	}
	authURL := authorizeURLFunc(tenantID) + "?" + params.Encode()

	fmt.Fprintf(os.Stderr, "Opening browser for Microsoft sign-in...\n")
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser automatically.\nOpen this URL manually:\n  %s\n", authURL)
	}

	// Wait for callback or context cancellation.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		if result.Err != nil {
			return nil, fmt.Errorf("authorization failed: %w", result.Err)
		}
		return exchangeCodeForToken(ctx, httpClient, clientID, tenantID, clientSecret, result.Code, verifier, redirectURI)
	}
}

// exchangeCodeForToken sends the authorization code to the token endpoint.
func exchangeCodeForToken(ctx context.Context, httpClient *http.Client, clientID, tenantID, clientSecret, code, verifier, redirectURI string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
		"scope":         {scopes},
	}
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURLFunc(tenantID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errResp); decodeErr == nil && errResp.ErrorDescription != "" {
			return nil, fmt.Errorf("token exchange failed: %s — %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token exchange failed (status %d)", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token exchange response: %w", err)
	}
	tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return &tokenResp, nil
}
