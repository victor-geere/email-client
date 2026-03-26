package quote

import (
	"strings"

	"github.com/victor/email-linearize/internal/domain"
	"golang.org/x/net/html"
)

// detectBlockquotes finds <blockquote> elements in HTML and returns the text ranges.
func detectBlockquotes(htmlContent string) []domain.QuoteRegion {
	textContent := StripHTML(htmlContent)
	var regions []domain.QuoteRegion

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var bqTexts []string
	var findBlockquotes func(*html.Node)
	findBlockquotes = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "blockquote" {
			text := extractText(n)
			if trimmed := strings.TrimSpace(text); trimmed != "" {
				bqTexts = append(bqTexts, trimmed)
			}
			return // don't recurse into nested blockquotes separately
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBlockquotes(c)
		}
	}
	findBlockquotes(doc)

	for _, bqText := range bqTexts {
		idx := strings.Index(textContent, bqText)
		if idx >= 0 {
			regions = append(regions, domain.QuoteRegion{
				StartOffset:     idx,
				EndOffset:       idx + len(bqText),
				OriginalSnippet: truncate(bqText, 100),
			})
		}
	}
	return regions
}

func extractText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
