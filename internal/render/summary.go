package render

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// topicEntry holds metadata about an HTML file that matches a topic search.
type topicEntry struct {
	RelPath string // relative to outputDir
	Title   string
	Date    string // yyyy-MM-dd from filename prefix
	Snippet string // first paragraph of body text
}

var bodyTextRe = regexp.MustCompile(`(?s)<main>(.*?)</main>`)
var paraRe = regexp.MustCompile(`(?s)<p>(.*?)</p>`)
var tagStripper = regexp.MustCompile(`<[^>]+>`)

// extractSnippet pulls the first non-metadata paragraph from the HTML body.
func extractSnippet(htmlContent string, maxLen int) string {
	bodyMatch := bodyTextRe.FindStringSubmatch(htmlContent)
	if len(bodyMatch) < 2 {
		return ""
	}
	body := bodyMatch[1]

	paras := paraRe.FindAllStringSubmatch(body, -1)
	for _, p := range paras {
		text := tagStripper.ReplaceAllString(p[1], "")
		text = strings.TrimSpace(text)
		// Skip metadata lines (From:, To:, Date:, Subject:, Attachments:, Cc:)
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "From:") || strings.HasPrefix(text, "To:") ||
			strings.HasPrefix(text, "Date:") || strings.HasPrefix(text, "Subject:") ||
			strings.HasPrefix(text, "Attachments:") || strings.HasPrefix(text, "Cc:") {
			continue
		}
		if len(text) > maxLen {
			text = text[:maxLen] + "…"
		}
		return text
	}
	return ""
}

// findTopicEntries scans all HTML files in outputDir for those matching the
// topic (case-insensitive substring match in title or filename).
func findTopicEntries(outputDir, topic string) ([]topicEntry, error) {
	topicLower := strings.ToLower(topic)
	var entries []topicEntry

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".html" {
			return nil
		}
		if isAttachmentsDir(path, outputDir) {
			return nil
		}

		rel, _ := filepath.Rel(outputDir, path)
		base := filepath.Base(rel)
		if base == "index.html" {
			return nil
		}
		// Skip existing topic files
		if strings.HasSuffix(base, "-index.html") || isTopicSummaryFile(base) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		contentStr := string(content)

		title := extractHTMLTitle(contentStr)
		titleLower := strings.ToLower(title)
		baseLower := strings.ToLower(base)

		if !strings.Contains(titleLower, topicLower) && !strings.Contains(baseLower, topicLower) {
			return nil
		}

		date := ""
		if len(base) >= 10 {
			date = base[:10]
		}

		snippet := extractSnippet(contentStr, 200)

		entries = append(entries, topicEntry{
			RelPath: rel,
			Title:   title,
			Date:    date,
			Snippet: snippet,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	return entries, nil
}

func isTopicSummaryFile(base string) bool {
	// Topic summary files end in .html but not -index.html and don't have
	// the date prefix pattern. We can't perfectly distinguish, so we skip
	// files that don't start with a date prefix.
	if len(base) < 10 {
		return true
	}
	return false
}

// GenerateTopicPages creates two HTML files in outputDir:
//   - <slug>-index.html: a flat list (no accordion) of matching email threads
//   - <slug>.html: a summary with dates and inline hyperlinks to each thread
func GenerateTopicPages(outputDir, topic string) error {
	entries, err := findTopicEntries(outputDir, topic)
	if err != nil {
		return fmt.Errorf("scan output directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no email threads found matching topic %q", topic)
	}

	slug := Slugify(topic)

	if err := writeTopicIndex(outputDir, slug, topic, entries); err != nil {
		return err
	}
	return writeTopicSummary(outputDir, slug, topic, entries)
}

func writeTopicIndex(outputDir, slug, topic string, entries []topicEntry) error {
	banner := bannerHTML(outputDir, outputDir)
	title := fmt.Sprintf("%s — Index", topic)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
%s
<main>
<h1>%s</h1>
<p>%d email threads matching "%s".</p>
<ul>
`, html.EscapeString(title), banner, html.EscapeString(title), len(entries), html.EscapeString(topic)))

	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a> <small>(%s)</small></li>\n",
			e.RelPath, html.EscapeString(e.Title), e.Date))
	}

	sb.WriteString("</ul>\n</main>\n</body>\n</html>\n")

	path := filepath.Join(outputDir, slug+"-index.html")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func writeTopicSummary(outputDir, slug, topic string, entries []topicEntry) error {
	banner := bannerHTML(outputDir, outputDir)
	title := fmt.Sprintf("%s — Summary", topic)
	indexFile := slug + "-index.html"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
%s
<main>
<h1>%s</h1>
<p>Summary of %d email threads about "<strong>%s</strong>". See also the <a href="%s">full index</a>.</p>
`, html.EscapeString(title), banner, html.EscapeString(title), len(entries),
		html.EscapeString(topic), indexFile))

	// Group entries by month for the summary
	months := make(map[string][]topicEntry)
	var monthOrder []string
	for _, e := range entries {
		month := ""
		if len(e.Date) >= 7 {
			month = e.Date[:7]
		}
		if _, exists := months[month]; !exists {
			monthOrder = append(monthOrder, month)
		}
		months[month] = append(months[month], e)
	}

	for _, month := range monthOrder {
		sb.WriteString(fmt.Sprintf("<h2>%s</h2>\n", month))
		for _, e := range months[month] {
			sb.WriteString(fmt.Sprintf("<h3><a href=\"%s\">%s</a></h3>\n",
				e.RelPath, html.EscapeString(e.Title)))
			sb.WriteString(fmt.Sprintf("<p><strong>Date:</strong> %s</p>\n", e.Date))
			if e.Snippet != "" {
				sb.WriteString(fmt.Sprintf("<p>%s</p>\n", html.EscapeString(e.Snippet)))
			}
		}
	}

	sb.WriteString("</main>\n</body>\n</html>\n")

	path := filepath.Join(outputDir, slug+".html")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
