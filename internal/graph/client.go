package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/victor/email-linearize/internal/domain"
)

const defaultBaseURL = "https://graph.microsoft.com/v1.0"

const selectFields = "id,conversationId,subject,from,toRecipients,ccRecipients,receivedDateTime,body,bodyPreview"

// FetchOptions controls message fetching behaviour.
type FetchOptions struct {
	Top            int
	ConversationID string
}

// Client communicates with the Microsoft Graph API.
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	TokenSource func(ctx context.Context) (string, error)
}

// NewClient creates a Graph client with the given token source.
func NewClient(baseURL string, tokenSource func(ctx context.Context) (string, error)) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		HTTPClient:  http.DefaultClient,
		TokenSource: tokenSource,
	}
}

// graphEmailAddress matches Graph API's nested emailAddress shape.
type graphEmailAddress struct {
	EmailAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"emailAddress"`
}

type graphMessage struct {
	ID               string              `json:"id"`
	ConversationID   string              `json:"conversationId"`
	Subject          string              `json:"subject"`
	From             graphEmailAddress   `json:"from"`
	ToRecipients     []graphEmailAddress `json:"toRecipients"`
	CcRecipients     []graphEmailAddress `json:"ccRecipients"`
	ReceivedDateTime string              `json:"receivedDateTime"`
	Body             struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
	BodyPreview string `json:"bodyPreview"`
}

type messagesResponse struct {
	Value    []graphMessage `json:"value"`
	NextLink string         `json:"@odata.nextLink"`
}

type folderResponse struct {
	Value []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"value"`
}

func toDomainMessage(gm graphMessage) domain.Message {
	t, _ := parseTime(gm.ReceivedDateTime)
	msg := domain.Message{
		ID:             gm.ID,
		ConversationID: gm.ConversationID,
		Subject:        gm.Subject,
		From: domain.EmailAddress{
			Name:    gm.From.EmailAddress.Name,
			Address: gm.From.EmailAddress.Address,
		},
		ReceivedDateTime: t,
		Body: domain.Body{
			ContentType: gm.Body.ContentType,
			Content:     gm.Body.Content,
		},
		BodyPreview: gm.BodyPreview,
	}
	for _, r := range gm.ToRecipients {
		msg.ToRecipients = append(msg.ToRecipients, domain.EmailAddress{
			Name:    r.EmailAddress.Name,
			Address: r.EmailAddress.Address,
		})
	}
	for _, r := range gm.CcRecipients {
		msg.CcRecipients = append(msg.CcRecipients, domain.EmailAddress{
			Name:    r.EmailAddress.Name,
			Address: r.EmailAddress.Address,
		})
	}
	return msg
}

func (c *Client) doAuthRequest(ctx context.Context, fullURL string) (*http.Response, error) {
	token, err := c.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	return c.HTTPClient.Do(req)
}

// ResolveFolderID maps a folder display name to its Graph API ID.
func (c *Client) ResolveFolderID(ctx context.Context, folderName string) (string, error) {
	fullURL := c.BaseURL + "/me/mailFolders"
	resp, err := c.doAuthRequest(ctx, fullURL)
	if err != nil {
		return "", fmt.Errorf("fetch mail folders: %w", err)
	}

	resp, err = handleErrors(ctx, resp, func() (*http.Response, error) {
		return c.doAuthRequest(ctx, fullURL)
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read folder response: %w", err)
	}

	var folders folderResponse
	if err := json.Unmarshal(body, &folders); err != nil {
		return "", fmt.Errorf("decode folder response: %w", err)
	}

	target := strings.ToLower(folderName)
	for _, f := range folders.Value {
		if strings.ToLower(f.DisplayName) == target {
			return f.ID, nil
		}
	}
	return "", &NotFoundError{Resource: fmt.Sprintf("folder %q", folderName)}
}

// FetchMessages retrieves messages from a folder with pagination.
func (c *Client) FetchMessages(ctx context.Context, folderID string, opts FetchOptions) ([]domain.Message, error) {
	query := url.Values{
		"$select":  {selectFields},
		"$orderby": {"receivedDateTime asc"},
	}
	top := opts.Top
	if top <= 0 {
		top = 50
	}
	query.Set("$top", fmt.Sprintf("%d", top))
	if opts.ConversationID != "" {
		query.Set("$filter", fmt.Sprintf("conversationId eq '%s'", opts.ConversationID))
	}

	firstURL := c.BaseURL + fmt.Sprintf("/me/mailFolders/%s/messages", folderID) + "?" + query.Encode()
	return c.fetchAllPages(ctx, firstURL)
}

func (c *Client) fetchAllPages(ctx context.Context, pageURL string) ([]domain.Message, error) {
	var all []domain.Message

	for pageURL != "" {
		resp, err := c.doAuthRequest(ctx, pageURL)
		if err != nil {
			return nil, fmt.Errorf("fetch messages: %w", err)
		}

		resp, err = handleErrors(ctx, resp, func() (*http.Response, error) {
			return c.doAuthRequest(ctx, pageURL)
		})
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		var page messagesResponse
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		for _, gm := range page.Value {
			all = append(all, toDomainMessage(gm))
		}
		pageURL = page.NextLink
	}

	return all, nil
}

// FetchThreadMessages retrieves all messages for a specific conversation.
func (c *Client) FetchThreadMessages(ctx context.Context, conversationID string) ([]domain.Message, error) {
	return c.FetchMessages(ctx, "inbox", FetchOptions{
		ConversationID: conversationID,
	})
}
