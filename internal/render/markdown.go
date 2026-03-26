package render

import (
	"fmt"
	"strings"

	"github.com/victor/email-linearize/internal/domain"
)

// RenderMarkdown produces a full Markdown document from a linearized thread.
func RenderMarkdown(thread domain.LinearizedThread) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Thread: %s\n\n", thread.Subject))
	sb.WriteString("---\n\n")

	for _, msg := range thread.Messages {
		sb.WriteString(fmt.Sprintf("### Message #%d\n\n", msg.Ordinal))
		sb.WriteString(fmt.Sprintf("**From:** %s <%s>\n", msg.From.Name, msg.From.Address))

		if len(msg.ToRecipients) > 0 {
			sb.WriteString(fmt.Sprintf("**To:** %s\n", formatRecipients(msg.ToRecipients)))
		}
		if len(msg.CcRecipients) > 0 {
			sb.WriteString(fmt.Sprintf("**Cc:** %s\n", formatRecipients(msg.CcRecipients)))
		}

		sb.WriteString(fmt.Sprintf("**Date:** %s\n", msg.ReceivedDateTime.UTC().Format("2006-01-02 15:04 UTC")))
		sb.WriteString(fmt.Sprintf("**Subject:** %s\n", msg.Subject))
		sb.WriteString("\n")
		sb.WriteString(msg.CleanBody)
		sb.WriteString("\n\n---\n\n")
	}

	return sb.String()
}

func formatRecipients(addrs []domain.EmailAddress) string {
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		if a.Name != "" {
			parts[i] = fmt.Sprintf("%s <%s>", a.Name, a.Address)
		} else {
			parts[i] = a.Address
		}
	}
	return strings.Join(parts, ", ")
}
