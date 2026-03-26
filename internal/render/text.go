package render

import (
	"fmt"
	"strings"

	"github.com/victor/email-linearize/internal/domain"
)

// RenderText produces a plain-text rendering of a linearized thread.
func RenderText(thread domain.LinearizedThread) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Thread: %s\n", thread.Subject))
	sb.WriteString("============================================================\n\n")

	for _, msg := range thread.Messages {
		sb.WriteString(fmt.Sprintf("Message #%d\n", msg.Ordinal))
		sb.WriteString(fmt.Sprintf("From: %s <%s>\n", msg.From.Name, msg.From.Address))

		if len(msg.ToRecipients) > 0 {
			sb.WriteString(fmt.Sprintf("To: %s\n", formatRecipientsText(msg.ToRecipients)))
		}
		if len(msg.CcRecipients) > 0 {
			sb.WriteString(fmt.Sprintf("Cc: %s\n", formatRecipientsText(msg.CcRecipients)))
		}

		sb.WriteString(fmt.Sprintf("Date: %s\n", msg.ReceivedDateTime.UTC().Format("2006-01-02 15:04 UTC")))
		sb.WriteString(fmt.Sprintf("Subject: %s\n", msg.Subject))
		sb.WriteString("\n")
		sb.WriteString(msg.CleanBody)
		sb.WriteString("\n\n============================================================\n\n")
	}

	return sb.String()
}

func formatRecipientsText(addrs []domain.EmailAddress) string {
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
