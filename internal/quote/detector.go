package quote

import (
	"sort"

	"github.com/victor/email-linearize/internal/domain"
)

// Detect finds all quoted regions in a message by running all detection strategies.
func Detect(msg domain.AnnotatedMessage, earlier []domain.AnnotatedMessage) []domain.QuoteRegion {
	var regions []domain.QuoteRegion

	// Run HTML-based detection if the body is HTML
	if msg.Body.ContentType == "html" {
		regions = append(regions, detectBlockquotes(msg.Body.Content)...)
	}

	// Get text content for text-based strategies
	textContent := msg.Body.Content
	if msg.Body.ContentType == "html" {
		textContent = StripHTML(msg.Body.Content)
	}

	// Run text-based strategies
	regions = append(regions, detectLinePrefixQuotes(textContent)...)
	regions = append(regions, detectHeaderQuotes(textContent)...)

	// Merge overlapping regions
	regions = mergeRegions(regions)

	// Attribute each region to an earlier message
	for i := range regions {
		if regions[i].AttributedToOrdinal == nil {
			quoteText := ""
			if regions[i].StartOffset < len(textContent) && regions[i].EndOffset <= len(textContent) {
				quoteText = textContent[regions[i].StartOffset:regions[i].EndOffset]
			}
			regions[i].AttributedToOrdinal = attributeQuote(quoteText, earlier)
		}
	}

	return regions
}

// mergeRegions combines overlapping quote regions.
func mergeRegions(regions []domain.QuoteRegion) []domain.QuoteRegion {
	if len(regions) == 0 {
		return nil
	}

	sort.Slice(regions, func(i, j int) bool {
		return regions[i].StartOffset < regions[j].StartOffset
	})

	merged := []domain.QuoteRegion{regions[0]}
	for _, r := range regions[1:] {
		last := &merged[len(merged)-1]
		if r.StartOffset <= last.EndOffset {
			if r.EndOffset > last.EndOffset {
				last.EndOffset = r.EndOffset
			}
			// Keep the longer snippet
			if len(r.OriginalSnippet) > len(last.OriginalSnippet) {
				last.OriginalSnippet = r.OriginalSnippet
			}
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}
