package render

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/victor/email-linearize/internal/domain"
)

// SAST is South Africa Standard Time (UTC+2).
var SAST = time.FixedZone("SAST", 2*60*60)

// quoteRefRe matches [→ quoted text omitted. see message #N from Name, DD Mon].
var quoteRefRe = regexp.MustCompile(`\[→ quoted text omitted\. see message #(\d+) from [^\]]+\]`)

// consecutiveNewlines caps runs of 3+ newlines to 2.
var consecutiveNewlines = regexp.MustCompile(`\n{3,}`)

// preserveNewlines converts body newlines for markdown:
// - \n\n stays as-is (paragraph break)
// - single \n becomes two trailing spaces + \n (hard line break)
// - runs of 3+ newlines are capped at 2
func preserveNewlines(s string) string {
	s = consecutiveNewlines.ReplaceAllString(s, "\n\n")

	var sb strings.Builder
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			// If previous line was empty or current line is empty, keep
			// the newline as-is (paragraph break).
			prevEmpty := lines[i-1] == "" || strings.TrimSpace(lines[i-1]) == ""
			currEmpty := line == "" || strings.TrimSpace(line) == ""
			if prevEmpty || currEmpty {
				sb.WriteString("\n")
			} else {
				// Hard line break: two trailing spaces before newline
				sb.WriteString("  \n")
			}
		}
		sb.WriteString(line)
	}
	return sb.String()
}

// RenderMarkdown produces a full Markdown document from a linearized thread.
func RenderMarkdown(thread domain.LinearizedThread, attachDir string) string {
	return renderThread(thread, false, attachDir)
}

// RenderReport produces a Markdown document where message bodies are rendered
// from the original HTML with table-aware conversion, producing readable
// markdown tables for structured report emails.
func RenderReport(thread domain.LinearizedThread, attachDir string) string {
	return renderThread(thread, true, attachDir)
}

func renderThread(thread domain.LinearizedThread, reportMode bool, attachDir string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", thread.Subject))
	sb.WriteString("---\n\n")

	for _, msg := range thread.Messages {
		sb.WriteString(fmt.Sprintf("### Message #%d\n\n", msg.Ordinal))
		sb.WriteString(fmt.Sprintf("**From:** %s <%s>  \n", msg.From.Name, msg.From.Address))

		if len(msg.ToRecipients) > 0 {
			sb.WriteString(fmt.Sprintf("**To:** %s  \n", formatRecipients(msg.ToRecipients)))
		}
		if len(msg.CcRecipients) > 0 {
			sb.WriteString(fmt.Sprintf("**Cc:** %s  \n", formatRecipients(msg.CcRecipients)))
		}

		sb.WriteString(fmt.Sprintf("**Date:** %s  \n", msg.ReceivedDateTime.In(SAST).Format("2006-01-02 15:04 SAST")))
		sb.WriteString(fmt.Sprintf("**Subject:** %s  \n", msg.Subject))

		if len(msg.Attachments) > 0 {
			links := make([]string, len(msg.Attachments))
			for i, att := range msg.Attachments {
				filename := AttachmentFilename(msg.ReceivedDateTime, att.Name)
				links[i] = fmt.Sprintf("[%s](%s/%s)", att.Name, attachDir, filename)
			}
			sb.WriteString(fmt.Sprintf("**Attachments:** %s  \n", strings.Join(links, ", ")))
		}

		sb.WriteString("\n")

		body := msg.CleanBody
		if reportMode && msg.Body.ContentType == "html" {
			body = HTMLToMarkdown(msg.Body.Content)
		}
		// Hyperlink quote references to message anchors
		body = quoteRefRe.ReplaceAllStringFunc(body, func(match string) string {
			sub := quoteRefRe.FindStringSubmatch(match)
			if len(sub) < 2 {
				return match
			}
			return fmt.Sprintf("[%s](#message-%s)", match, sub[1])
		})
		sb.WriteString(preserveNewlines(body))
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
