// Package domain defines the core types used throughout the email-linearize tool.
package domain

import "time"

// EmailAddress represents an email participant with display name and address.
type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Body holds the content and content type of an email message body.
type Body struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// Message represents a single email message as retrieved from Microsoft Graph.
type Message struct {
	ID               string         `json:"id"`
	ConversationID   string         `json:"conversationId"`
	Subject          string         `json:"subject"`
	From             EmailAddress   `json:"from"`
	ToRecipients     []EmailAddress `json:"toRecipients"`
	CcRecipients     []EmailAddress `json:"ccRecipients"`
	ReceivedDateTime time.Time      `json:"receivedDateTime"`
	Body             Body           `json:"body"`
	BodyPreview      string         `json:"bodyPreview"`
}

// Thread represents a group of messages sharing a conversation ID.
type Thread struct {
	ConversationID string
	Subject        string
	Messages       []Message
}

// QuoteRegion marks a region of quoted text within a message body.
type QuoteRegion struct {
	StartOffset         int
	EndOffset           int
	AttributedToOrdinal *int
	OriginalSnippet     string
}

// AnnotatedMessage wraps a Message with linearization metadata.
type AnnotatedMessage struct {
	Message
	Ordinal      int
	QuoteRegions []QuoteRegion
	CleanBody    string
}

// LinearizedThread holds the final processed thread ready for rendering.
type LinearizedThread struct {
	ConversationID string
	Subject        string
	Messages       []AnnotatedMessage
}
