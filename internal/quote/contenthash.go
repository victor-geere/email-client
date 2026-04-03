package quote

import (
	"strings"

	"github.com/victor/email-linearize/internal/domain"
)

// normalize reduces text to a canonical form for comparison:
// lowercase, collapse whitespace, strip quote prefixes.
func normalize(s string) string {
	s = strings.ToLower(s)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := line
		for strings.HasPrefix(strings.TrimSpace(trimmed), ">") {
			trimmed = strings.TrimSpace(trimmed)
			trimmed = strings.TrimPrefix(trimmed, ">")
			trimmed = strings.TrimSpace(trimmed)
		}
		lines[i] = trimmed
	}
	s = strings.Join(lines, " ")
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// attributeQuote tries to match quote text to an earlier message.
// Returns the ordinal of the matching message, or nil if no match.
// Checks both directions: whether the quote appears in a message body
// (standard quoting) and whether a message body appears in the quote
// (header-prefixed quoting where the quote includes From/Sent/Subject headers).
func attributeQuote(quoteText string, earlier []domain.AnnotatedMessage) *int {
	normalizedQuote := normalize(quoteText)
	if len(normalizedQuote) < 10 {
		return nil // too short to match reliably
	}

	bestMatch := -1
	bestOverlap := 0

	for _, msg := range earlier {
		body := msg.Body.Content
		if msg.Body.ContentType == "html" {
			body = StripHTML(body)
		}
		normalizedBody := normalize(body)
		if len(normalizedBody) < 10 {
			continue
		}

		overlap := 0
		if strings.Contains(normalizedBody, normalizedQuote) {
			// Quote is a subset of the body (standard quoting)
			overlap = len(normalizedQuote)
		} else if strings.Contains(normalizedQuote, normalizedBody) {
			// Body is a subset of the quote (header-prefixed quoting)
			overlap = len(normalizedBody)
		}

		if overlap > bestOverlap {
			bestOverlap = overlap
			bestMatch = msg.Ordinal
		}
	}

	if bestMatch >= 0 {
		return &bestMatch
	}
	return nil
}
