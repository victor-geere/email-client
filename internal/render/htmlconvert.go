package render

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

var goldmarkConverter = goldmark.New(
	goldmark.WithExtensions(extension.Table),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(goldmarkhtml.WithHardWraps()),
)

const cssFilename = "style.css"

const cssContent = `* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  line-height: 1.6;
  color: #24292e;
  background: #f6f8fa;
}

main {
  max-width: 900px;
  margin: 2rem auto;
  padding: 0 1.5rem;
  text-align: left;
}

h1 {
  margin-bottom: 1.5rem;
  font-size: 1.75rem;
}

h2 {
  margin-top: 2rem;
  margin-bottom: 0.75rem;
}

h3 {
  margin-top: 1.5rem;
  margin-bottom: 0.5rem;
}

p {
  margin-bottom: 0.75rem;
}

hr {
  margin: 2rem 0;
  border: none;
  border-top: 1px solid #e1e4e8;
}

table {
  border-collapse: collapse;
  width: 100%;
  margin: 1rem 0;
}

th, td {
  border: 1px solid #e1e4e8;
  padding: 0.5rem 0.75rem;
  text-align: left;
}

th {
  background: #f6f8fa;
  font-weight: 600;
}

a {
  color: #0366d6;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}

details {
  margin-bottom: 0.5rem;
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  background: #fff;
}

summary {
  padding: 0.75rem 1rem;
  cursor: pointer;
  font-weight: 600;
  user-select: none;
}

summary:hover {
  background: #f6f8fa;
}

details[open] summary {
  border-bottom: 1px solid #e1e4e8;
}

details ul {
  list-style: none;
  padding: 0.5rem 1rem;
}

details li {
  padding: 0.35rem 0;
}

.expand-btn {
  margin-bottom: 1rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  background: #fff;
  font-size: 0.875rem;
  font-family: inherit;
  color: #24292e;
}

.expand-btn:hover {
  background: #f6f8fa;
}

strong {
  font-weight: 600;
}

blockquote {
  margin: 1rem 0;
  padding-left: 1rem;
  border-left: 3px solid #e1e4e8;
  color: #6a737d;
}

.banner {
  background: #24292e;
  padding: 0.75rem 1.5rem;
  display: flex;
  gap: 1.5rem;
}

.banner a {
  color: #fff;
  font-weight: 500;
  font-size: 0.875rem;
}

.banner a:hover {
  text-decoration: underline;
}
`

func convertMarkdown(source []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := goldmarkConverter.Convert(source, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func extractTitle(mdContent string) string {
	for _, line := range strings.Split(mdContent, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Email Thread"
}

// bannerHTML generates the navigation banner with correct relative links.
func bannerHTML(fileDir, outputDir string) string {
	rel, err := filepath.Rel(fileDir, outputDir)
	if err != nil {
		rel = "."
	}
	prefix := filepath.ToSlash(rel)
	if prefix == "." {
		prefix = ""
	} else {
		prefix += "/"
	}
	return fmt.Sprintf(`<nav class="banner">
<a href="%sindex.html">Home</a>
<a href="%ssprint-completion-reports/index.html">Sprint Reports</a>
<a href="%sdaily-activity-reports/index.html">Activity Reports</a>
</nav>`, prefix, prefix, prefix)
}

func wrapHTMLPage(body []byte, title, cssPath, banner string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<link rel="stylesheet" href="%s">
</head>
<body>
%s
<main>
%s
</main>
</body>
</html>
`, html.EscapeString(title), cssPath, banner, string(body))
	return buf.Bytes()
}

func isAttachmentsDir(path, outputDir string) bool {
	rel, err := filepath.Rel(outputDir, path)
	if err != nil {
		return false
	}
	parts := strings.Split(rel, string(filepath.Separator))
	return len(parts) > 0 && parts[0] == "attachments"
}

// ConvertOutputToHTML converts all markdown files in the output directory to
// HTML, writes a shared CSS file, and generates an index.html with an
// accordion grouping threads by day. Markdown files in attachments/ are not
// converted or deleted.
func ConvertOutputToHTML(outputDir string) error {
	var mdFiles []string
	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".md" && !isAttachmentsDir(path, outputDir) {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	for _, mdPath := range mdFiles {
		content, err := os.ReadFile(mdPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", mdPath, err)
		}

		htmlBody, err := convertMarkdown(content)
		if err != nil {
			return fmt.Errorf("convert %s: %w", mdPath, err)
		}

		title := extractTitle(string(content))

		dir := filepath.Dir(mdPath)
		cssRel, _ := filepath.Rel(dir, filepath.Join(outputDir, cssFilename))
		banner := bannerHTML(dir, outputDir)

		page := wrapHTMLPage(htmlBody, title, cssRel, banner)

		htmlPath := strings.TrimSuffix(mdPath, ".md") + ".html"
		if err := os.WriteFile(htmlPath, page, 0644); err != nil {
			return fmt.Errorf("write %s: %w", htmlPath, err)
		}

		os.Remove(mdPath)
	}

	cssPath := filepath.Join(outputDir, cssFilename)
	if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
		return fmt.Errorf("write CSS: %w", err)
	}

	if err := generateIndex(outputDir); err != nil {
		return err
	}
	if err := generateSubdirIndex(outputDir, "sprint-completion-reports", "Sprint Completion Reports"); err != nil {
		return err
	}
	return generateSubdirIndex(outputDir, "daily-activity-reports", "Daily Activity Reports")
}

type htmlFileEntry struct {
	Path  string
	Title string
	Date  string
}

var htmlTitleRe = regexp.MustCompile(`<title>([^<]+)</title>`)

func extractHTMLTitle(content string) string {
	match := htmlTitleRe.FindStringSubmatch(content)
	if len(match) >= 2 {
		title := html.UnescapeString(match[1])

		return title
	}
	return "Email Thread"
}

func generateIndex(outputDir string) error {
	var entries []htmlFileEntry

	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
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
		if filepath.Base(rel) == "index.html" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		title := extractHTMLTitle(string(content))

		base := filepath.Base(rel)
		date := ""
		if len(base) >= 10 {
			date = base[:10]
		}

		entries = append(entries, htmlFileEntry{
			Path:  rel,
			Title: title,
			Date:  date,
		})
		return nil
	})

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	groups := make(map[string][]htmlFileEntry)
	var dayOrder []string
	for _, e := range entries {
		if _, exists := groups[e.Date]; !exists {
			dayOrder = append(dayOrder, e.Date)
		}
		groups[e.Date] = append(groups[e.Date], e)
	}

	banner := bannerHTML(outputDir, outputDir)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Email Threads</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
%s
<main>
<h1>Email Threads</h1>
<button class="expand-btn" onclick="document.querySelectorAll('details').forEach(d=>d.open=true)">Expand All</button>
`, banner))

	for _, day := range dayOrder {
		sb.WriteString(fmt.Sprintf("<details>\n<summary>%s</summary>\n<ul>\n", day))
		for _, e := range groups[day] {
			sb.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", e.Path, html.EscapeString(e.Title)))
		}
		sb.WriteString("</ul>\n</details>\n")
	}

	sb.WriteString("</main>\n</body>\n</html>\n")

	return os.WriteFile(filepath.Join(outputDir, "index.html"), []byte(sb.String()), 0644)
}

func generateSubdirIndex(outputDir, subdir, title string) error {
	subdirPath := filepath.Join(outputDir, subdir)
	info, err := os.Stat(subdirPath)
	if err != nil || !info.IsDir() {
		return nil
	}

	var entries []htmlFileEntry
	files, err := os.ReadDir(subdirPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", subdirPath, err)
	}

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".html" || f.Name() == "index.html" {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(subdirPath, f.Name()))
		if readErr != nil {
			continue
		}
		fTitle := extractHTMLTitle(string(content))
		month := ""
		if len(f.Name()) >= 7 {
			month = f.Name()[:7]
		}
		entries = append(entries, htmlFileEntry{
			Path:  f.Name(),
			Title: fTitle,
			Date:  month,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	groups := make(map[string][]htmlFileEntry)
	var monthOrder []string
	for _, e := range entries {
		if _, exists := groups[e.Date]; !exists {
			monthOrder = append(monthOrder, e.Date)
		}
		groups[e.Date] = append(groups[e.Date], e)
	}

	banner := bannerHTML(subdirPath, outputDir)
	cssRel, _ := filepath.Rel(subdirPath, filepath.Join(outputDir, cssFilename))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<link rel="stylesheet" href="%s">
</head>
<body>
%s
<main>
<h1>%s</h1>
<button class="expand-btn" onclick="document.querySelectorAll('details').forEach(d=>d.open=true)">Expand All</button>
`, html.EscapeString(title), cssRel, banner, html.EscapeString(title)))

	for _, month := range monthOrder {
		sb.WriteString(fmt.Sprintf("<details>\n<summary>%s</summary>\n<ul>\n", month))
		for _, e := range groups[month] {
			sb.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", e.Path, html.EscapeString(e.Title)))
		}
		sb.WriteString("</ul>\n</details>\n")
	}

	sb.WriteString("</main>\n</body>\n</html>\n")

	return os.WriteFile(filepath.Join(subdirPath, "index.html"), []byte(sb.String()), 0644)
}
