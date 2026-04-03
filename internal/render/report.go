package render

import (
	"strings"

	"golang.org/x/net/html"
)

// isReport returns true if the slug matches a report routing pattern.
func isReport(slug string) bool {
	return strings.Contains(slug, "daily-activity-report") ||
		strings.Contains(slug, "sprint-completion-report")
}

// HTMLToMarkdown converts HTML to Markdown, preserving table structure as
// markdown tables. Used for report emails where tabular data must stay readable.
// Does not fetch remote resources.
func HTMLToMarkdown(htmlContent string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	var result strings.Builder
	skip := 0

	// Table state
	inTable := false
	var tableRows [][]string
	var currentRow []string
	var cellBuf strings.Builder
	inCell := false
	isHeaderRow := false
	headerRowIdx := -1

	flushTable := func() {
		if len(tableRows) == 0 {
			return
		}
		// Calculate column widths
		cols := 0
		for _, row := range tableRows {
			if len(row) > cols {
				cols = len(row)
			}
		}
		if cols == 0 {
			return
		}
		// Pad rows to same length
		for i := range tableRows {
			for len(tableRows[i]) < cols {
				tableRows[i] = append(tableRows[i], "")
			}
		}
		// Write markdown table
		for i, row := range tableRows {
			result.WriteString("| ")
			result.WriteString(strings.Join(row, " | "))
			result.WriteString(" |\n")
			// Add separator after header row
			if i == headerRowIdx || (headerRowIdx < 0 && i == 0) {
				result.WriteString("|")
				for range cols {
					result.WriteString(" --- |")
				}
				result.WriteString("\n")
			}
		}
		result.WriteString("\n")
		tableRows = nil
		headerRowIdx = -1
	}

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			flushTable()
			return strings.TrimSpace(result.String())

		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			switch tag {
			case "script", "style", "head":
				skip++
			case "br":
				if inCell {
					cellBuf.WriteString(" ")
				} else {
					result.WriteString("\n")
				}
			case "table":
				inTable = true
				tableRows = nil
				headerRowIdx = -1
			case "tr":
				currentRow = nil
				isHeaderRow = false
			case "th":
				inCell = true
				isHeaderRow = true
				cellBuf.Reset()
			case "td":
				inCell = true
				cellBuf.Reset()
			case "h1":
				if !inTable {
					result.WriteString("\n# ")
				}
			case "h2":
				if !inTable {
					result.WriteString("\n## ")
				}
			case "h3":
				if !inTable {
					result.WriteString("\n### ")
				}
			case "strong", "b":
				if !inCell {
					result.WriteString("**")
				}
			}

		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			switch tag {
			case "script", "style", "head":
				if skip > 0 {
					skip--
				}
			case "p", "div", "li":
				if !inTable {
					result.WriteString("\n")
				}
			case "tr":
				if inTable {
					tableRows = append(tableRows, currentRow)
					if isHeaderRow {
						headerRowIdx = len(tableRows) - 1
					}
				} else {
					result.WriteString("\n")
				}
			case "th", "td":
				if inTable {
					cell := strings.TrimSpace(cellBuf.String())
					currentRow = append(currentRow, cell)
				}
				inCell = false
			case "table":
				flushTable()
				inTable = false
			case "h1", "h2", "h3":
				if !inTable {
					result.WriteString("\n")
				}
			case "strong", "b":
				if !inCell {
					result.WriteString("**")
				}
			}

		case html.SelfClosingTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "br" {
				if inCell {
					cellBuf.WriteString(" ")
				} else {
					result.WriteString("\n")
				}
			}

		case html.TextToken:
			if skip == 0 {
				text := tokenizer.Token().Data
				if inCell {
					cellBuf.WriteString(text)
				} else {
					result.WriteString(text)
				}
			}
		}
	}
}
