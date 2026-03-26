package linearize

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/victor/email-linearize/internal/domain"
	"github.com/victor/email-linearize/internal/quote"
)

var multiBlankLineRe = regexp.MustCompile(`\n{3,}`)

// Linearize processes a thread by detecting quotes and replacing them with references.
func Linearize(thread domain.Thread, annotated []domain.AnnotatedMessage) domain.LinearizedThread {
	for i := range annotated {
		earlier := annotated[:i]
		regions := quote.Detect(annotated[i], earlier)
		annotated[i].QuoteRegions = regions

		body := annotated[i].Body.Content
		if annotated[i].Body.ContentType == "html" {
			body = quote.StripHTML(body)
		}

		// Replace regions in reverse order to preserve offsets
		sortedRegions := make([]domain.QuoteRegion, len(regions))
		copy(sortedRegions, regions)
		for j := len(sortedRegions)/2 - 1; j >= 0; j-- {
			opp := len(sortedRegions) - 1 - j
			sortedRegions[j], sortedRegions[opp] = sortedRegions[opp], sortedRegions[j]
		}

		for _, region := range sortedRegions {
			if region.StartOffset >= len(body) || region.EndOffset > len(body) {
				continue
			}
			ref := formatReference(region, annotated)
			body = body[:region.StartOffset] + ref + body[region.EndOffset:]
		}

		body = multiBlankLineRe.ReplaceAllString(body, "\n\n")
		body = strings.TrimSpace(body)
		annotated[i].CleanBody = body
	}

	return domain.LinearizedThread{
		ConversationID: thread.ConversationID,
		Subject:        thread.Subject,
		Messages:       annotated,
	}
}

func formatReference(region domain.QuoteRegion, messages []domain.AnnotatedMessage) string {
	if region.AttributedToOrdinal == nil {
		return "[→ quoted text omitted]"
	}

	ordinal := *region.AttributedToOrdinal
	for _, msg := range messages {
		if msg.Ordinal == ordinal {
			firstName := strings.Fields(msg.From.Name)
			name := msg.From.Name
			if len(firstName) > 0 {
				name = firstName[0]
			}
			date := msg.ReceivedDateTime.Format("02 Jan")
			return fmt.Sprintf("[→ see msg #%d from %s, %s]", ordinal, name, date)
		}
	}
	return fmt.Sprintf("[→ see msg #%d]", ordinal)
}
