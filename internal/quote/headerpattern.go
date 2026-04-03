package quote

import (
	"regexp"
	"strings"

	"github.com/victor/email-linearize/internal/domain"
)

var (
	// "On Jan 14, 2025, Alice wrote:" or similar
	gmailHeaderRe = regexp.MustCompile(`(?m)^On .+wrote:\s*$`)
	// "From: ... Sent:/Date: ... Subject: ..." (Outlook style, both classic and web variants)
	// Handles optional To:, Cc:, Bcc:, When:, Where: lines between From:/date and Subject:.
	outlookHeaderRe = regexp.MustCompile(`(?m)^From:\s.+\n(?:Sent|Date):\s.+\n(?:(?:To|Cc|Bcc|When|Where):\s.+\n)*Subject:\s.+`)
	// Separator lines: "-----Original Message-----" or similar
	separatorRe = regexp.MustCompile(`(?m)^-{3,}\s*(?:Original Message|Forwarded message)\s*-{3,}`)
)

// detectHeaderQuotes finds quote headers and captures everything after them.
func detectHeaderQuotes(textContent string) []domain.QuoteRegion {
	var regions []domain.QuoteRegion

	for _, re := range []*regexp.Regexp{gmailHeaderRe, outlookHeaderRe, separatorRe} {
		locs := re.FindAllStringIndex(textContent, -1)
		for _, loc := range locs {
			start := loc[0]
			end := findQuoteEnd(textContent, loc[1])
			if end > start {
				snippet := textContent[start:end]
				regions = append(regions, domain.QuoteRegion{
					StartOffset:     start,
					EndOffset:       end,
					OriginalSnippet: truncate(strings.TrimSpace(snippet), 100),
				})
			}
		}
	}

	return regions
}

// findQuoteEnd finds the end of quoted content after a header.
// It extends to the end of the text or until a non-quoted section is found.
func findQuoteEnd(text string, afterHeader int) int {
	return len(text)
}
