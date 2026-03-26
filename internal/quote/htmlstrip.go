package quote

import (
	"strings"

	"golang.org/x/net/html"
)

// StripHTML extracts text from HTML content, preserving paragraph breaks as newlines.
// It does not fetch remote resources (images, stylesheets).
func StripHTML(htmlContent string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	var result strings.Builder
	skip := 0

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return strings.TrimSpace(result.String())

		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			switch tag {
			case "script", "style", "head":
				skip++
			case "br":
				result.WriteString("\n")
			}

		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			switch tag {
			case "script", "style", "head":
				if skip > 0 {
					skip--
				}
			case "p", "div", "tr", "li":
				result.WriteString("\n")
			}

		case html.SelfClosingTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "br" {
				result.WriteString("\n")
			}

		case html.TextToken:
			if skip == 0 {
				text := tokenizer.Token().Data
				result.WriteString(text)
			}
		}
	}
}
