package quote

import (
	"strings"

	"github.com/victor/email-linearize/internal/domain"
)

// detectLinePrefixQuotes finds lines starting with ">" and groups them into quote regions.
func detectLinePrefixQuotes(textContent string) []domain.QuoteRegion {
	var regions []domain.QuoteRegion
	lines := strings.Split(textContent, "\n")

	inQuote := false
	quoteStart := 0
	currentOffset := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isQuote := strings.HasPrefix(trimmed, ">")

		if isQuote && !inQuote {
			inQuote = true
			quoteStart = currentOffset
		} else if !isQuote && inQuote {
			inQuote = false
			quoteEnd := currentOffset - 1 // don't include the trailing newline
			if quoteEnd > quoteStart {
				snippet := textContent[quoteStart:quoteEnd]
				regions = append(regions, domain.QuoteRegion{
					StartOffset:     quoteStart,
					EndOffset:       quoteEnd,
					OriginalSnippet: truncate(strings.TrimSpace(snippet), 100),
				})
			}
		}

		currentOffset += len(lines[i]) + 1 // +1 for newline
	}

	// Close any trailing quote
	if inQuote {
		quoteEnd := len(textContent)
		snippet := textContent[quoteStart:quoteEnd]
		regions = append(regions, domain.QuoteRegion{
			StartOffset:     quoteStart,
			EndOffset:       quoteEnd,
			OriginalSnippet: truncate(strings.TrimSpace(snippet), 100),
		})
	}

	return regions
}
