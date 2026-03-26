package graph

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// NotFoundError indicates a resource was not found.
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.Resource)
}

// PermissionError indicates insufficient permissions.
type PermissionError struct {
	Message string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: %s", e.Message)
}

// handleErrors inspects the HTTP response status and applies retry logic.
// retryFn is called once on 401 to refresh the token and retry.
func handleErrors(ctx context.Context, resp *http.Response, retryFn func() (*http.Response, error)) (*http.Response, error) {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return resp, nil

	case resp.StatusCode == 401:
		resp.Body.Close()
		retried, err := retryFn()
		if err != nil {
			return nil, fmt.Errorf("retry after 401: %w", err)
		}
		if retried.StatusCode == 401 {
			retried.Body.Close()
			return nil, fmt.Errorf("authentication failed after token refresh, please re-authenticate with --auth-only")
		}
		return retried, nil

	case resp.StatusCode == 403:
		resp.Body.Close()
		return nil, &PermissionError{Message: "re-consent may be required, run with --auth-only"}

	case resp.StatusCode == 404:
		resp.Body.Close()
		return nil, &NotFoundError{Resource: "requested resource"}

	case resp.StatusCode == 429:
		resp.Body.Close()
		return retryAfterWait(ctx, resp, retryFn)

	case resp.StatusCode >= 500:
		resp.Body.Close()
		return retryWithBackoff(ctx, retryFn)

	default:
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

func retryAfterWait(ctx context.Context, resp *http.Response, retryFn func() (*http.Response, error)) (*http.Response, error) {
	waitSec := 1
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if v, err := strconv.Atoi(ra); err == nil {
			waitSec = v
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(waitSec) * time.Second):
	}

	retried, err := retryFn()
	if err != nil {
		return nil, fmt.Errorf("retry after 429: %w", err)
	}
	if retried.StatusCode == 429 {
		retried.Body.Close()
		return nil, fmt.Errorf("still rate-limited after retry")
	}
	return retried, nil
}

func retryWithBackoff(ctx context.Context, retryFn func() (*http.Response, error)) (*http.Response, error) {
	for attempt := 0; attempt < 3; attempt++ {
		wait := time.Duration(1<<uint(attempt)) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}

		resp, err := retryFn()
		if err != nil {
			return nil, fmt.Errorf("retry attempt %d: %w", attempt+1, err)
		}
		if resp.StatusCode < 500 {
			return resp, nil
		}
		resp.Body.Close()
	}
	return nil, fmt.Errorf("server error persisted after 3 retries")
}
