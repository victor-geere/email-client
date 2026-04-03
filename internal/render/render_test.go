package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/victor/email-linearize/internal/domain"
)

func makeThread() domain.LinearizedThread {
	return domain.LinearizedThread{
		ConversationID: "conv-1",
		Subject:        "Q4 Budget Review",
		Messages: []domain.AnnotatedMessage{
			{
				Message: domain.Message{
					ID: "1", Subject: "Q4 Budget Review",
					From:             domain.EmailAddress{Name: "Alice Wang", Address: "alice@example.com"},
					ToRecipients:     []domain.EmailAddress{{Name: "Bob Chen", Address: "bob@example.com"}},
					CcRecipients:     []domain.EmailAddress{{Name: "Carol Davis", Address: "carol@example.com"}},
					ReceivedDateTime: time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
				},
				Ordinal:   1,
				CleanBody: "Please review the Q4 budget.",
			},
			{
				Message: domain.Message{
					ID: "2", Subject: "Re: Q4 Budget Review",
					From:             domain.EmailAddress{Name: "Bob Chen", Address: "bob@example.com"},
					ToRecipients:     []domain.EmailAddress{{Name: "Alice Wang", Address: "alice@example.com"}},
					ReceivedDateTime: time.Date(2025, 1, 14, 11, 0, 0, 0, time.UTC),
				},
				Ordinal:   2,
				CleanBody: "Looks good to me!\n\n[→ quoted text omitted. see message #1 from Alice, 14 Jan]",
			},
		},
	}
}

func TestRenderMarkdown_StructureCorrect(t *testing.T) {
	md := RenderMarkdown(makeThread(), "./attachments")

	if !strings.Contains(md, "# Q4 Budget Review") {
		t.Error("missing thread title")
	}
	if !strings.Contains(md, "### Message #1") {
		t.Error("missing message #1 header")
	}
	if !strings.Contains(md, "### Message #2") {
		t.Error("missing message #2 header")
	}
	if !strings.Contains(md, "**From:** Alice Wang <alice@example.com>") {
		t.Error("missing From line")
	}
	if !strings.Contains(md, "**Cc:** Carol Davis <carol@example.com>") {
		t.Error("missing Cc line")
	}
}

func TestRenderMarkdown_OmitsCcWhenEmpty(t *testing.T) {
	thread := makeThread()
	thread.Messages[1].CcRecipients = nil
	md := RenderMarkdown(thread, "./attachments")

	// Check that Cc line only appears once (for message #1)
	count := strings.Count(md, "**Cc:**")
	if count != 1 {
		t.Errorf("expected 1 Cc line, got %d", count)
	}
}

func TestRenderText_NoMarkdownSyntax(t *testing.T) {
	txt := RenderText(makeThread(), "./attachments")
	if strings.Contains(txt, "**") {
		t.Error("text output contains ** markdown syntax")
	}
	if strings.Contains(txt, "###") {
		t.Error("text output contains ### markdown syntax")
	}
}

func TestSlugify_SpecialChars(t *testing.T) {
	result := Slugify("Re: Q4 Budget!! Review")
	if result != "re-q4-budget-review" {
		t.Errorf("expected 're-q4-budget-review', got '%s'", result)
	}
}

func TestSlugify_LongSubject(t *testing.T) {
	long := strings.Repeat("a", 100)
	result := Slugify(long)
	if len(result) > 80 {
		t.Errorf("expected max 80 chars, got %d", len(result))
	}
}

func TestWriteFile_CollisionSafety(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()

	path1, err := WriteFile(dir, thread, "markdown")
	if err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	path2, err := WriteFile(dir, thread, "markdown")
	if err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	if path1 == path2 {
		t.Error("expected different paths for collision safety")
	}
	if !strings.HasSuffix(filepath.Base(path2), "-2.md") {
		t.Errorf("expected -2 suffix, got %s", filepath.Base(path2))
	}

	// Verify both files exist
	for _, p := range []string{path1, path2} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file %s does not exist: %v", p, err)
		}
	}
}

func TestWriteFile_DatePrefixFromLatestMessage(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	path, err := WriteFile(dir, thread, "markdown")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	base := filepath.Base(path)
	// Latest message is 2025-01-14 11:00 UTC = 2025-01-14 13:00 SAST
	if !strings.HasPrefix(base, "2025-01-14-13-00-") {
		t.Errorf("expected date prefix '2025-01-14-13-00-', got filename %s", base)
	}
}

func TestWriteFile_DailyActivityReportRoutedToSubfolder(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	thread.Subject = "Axiom Daily Activity Report 2026-02-05"
	path, err := WriteFile(dir, thread, "text")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	// Should be in daily-activity-reports subfolder
	relPath, _ := filepath.Rel(dir, path)
	if !strings.HasPrefix(relPath, "daily-activity-reports/") {
		t.Errorf("expected file in daily-activity-reports/, got %s", relPath)
	}
	// Should always be .md regardless of format flag
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("expected .md extension, got %s", filepath.Base(path))
	}
}

func TestWriteFile_SprintCompletionReportRoutedToSubfolder(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	thread.Subject = "Sprint Completion Report - March 2026"
	path, err := WriteFile(dir, thread, "text")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	relPath, _ := filepath.Rel(dir, path)
	if !strings.HasPrefix(relPath, "sprint-completion-reports/") {
		t.Errorf("expected file in sprint-completion-reports/, got %s", relPath)
	}
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("expected .md extension, got %s", filepath.Base(path))
	}
}

func TestCleanOutputDir_RemovesFilesNotDirs(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(sub, "file2.md"), []byte("y"), 0644)

	if err := CleanOutputDir(dir); err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	// Files should be gone
	if _, err := os.Stat(filepath.Join(dir, "file1.txt")); !os.IsNotExist(err) {
		t.Error("expected file1.txt to be deleted")
	}
	if _, err := os.Stat(filepath.Join(sub, "file2.md")); !os.IsNotExist(err) {
		t.Error("expected file2.md to be deleted")
	}
	// Directory should still exist
	info, err := os.Stat(sub)
	if err != nil || !info.IsDir() {
		t.Error("expected subdir to still exist")
	}
}

func TestHTMLToMarkdown_PreservesTableStructure(t *testing.T) {
	html := `<html><body>
<h2>Daily Activity Report</h2>
<table>
<tr><th>Issue</th><th>Synopsis</th><th>MRs</th></tr>
<tr><td>#123</td><td>Fix login bug</td><td>!674, !669</td></tr>
<tr><td>#104</td><td>Add search</td><td>!670</td></tr>
</table>
</body></html>`

	result := HTMLToMarkdown(html)
	if !strings.Contains(result, "| Issue | Synopsis | MRs |") {
		t.Errorf("expected markdown table header, got:\n%s", result)
	}
	if !strings.Contains(result, "| --- | --- | --- |") {
		t.Errorf("expected markdown table separator, got:\n%s", result)
	}
	if !strings.Contains(result, "| #123 | Fix login bug | !674, !669 |") {
		t.Errorf("expected markdown table row, got:\n%s", result)
	}
}

func TestHTMLToMarkdown_PreservesHeadings(t *testing.T) {
	html := `<html><body><h1>Title</h1><h2>Section</h2><p>Body text</p></body></html>`
	result := HTMLToMarkdown(html)
	if !strings.Contains(result, "# Title") {
		t.Errorf("expected h1 heading, got:\n%s", result)
	}
	if !strings.Contains(result, "## Section") {
		t.Errorf("expected h2 heading, got:\n%s", result)
	}
}

func TestRenderReport_UsesHTMLToMarkdownForBody(t *testing.T) {
	thread := domain.LinearizedThread{
		ConversationID: "conv-1",
		Subject:        "Daily Activity Report",
		Messages: []domain.AnnotatedMessage{
			{
				Message: domain.Message{
					ID: "1", Subject: "Daily Activity Report",
					From:             domain.EmailAddress{Name: "Bot", Address: "bot@example.com"},
					ReceivedDateTime: time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC),
					Body: domain.Body{
						ContentType: "html",
						Content:     "<html><body><table><tr><th>Metric</th><th>Value</th></tr><tr><td>Issues</td><td>5</td></tr></table></body></html>",
					},
				},
				Ordinal:   1,
				CleanBody: "MetricValueIssues5", // This is what StripHTML produces (mashed)
			},
		},
	}

	result := RenderReport(thread, "./attachments")
	if !strings.Contains(result, "| Metric | Value |") {
		t.Errorf("expected markdown table in report output, got:\n%s", result)
	}
	if strings.Contains(result, "MetricValueIssues5") {
		t.Error("report should use HTMLToMarkdown, not CleanBody")
	}
}

func TestLatestDatePrefix_FindsLatestDate(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "daily-activity-reports")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "2026-03-01-10-00-thread-a.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(sub, "2026-03-25-11-05-report.md"), []byte("y"), 0644)
	os.WriteFile(filepath.Join(dir, "2026-03-10-08-15-thread-b.txt"), []byte("z"), 0644)

	// Filenames are SAST-based; 2026-03-25 11:05 SAST = 2026-03-25T09:05:00Z UTC
	result := LatestDatePrefix(dir)
	if result != "2026-03-25T09:05:00Z" {
		t.Errorf("expected '2026-03-25T09:05:00Z', got %q", result)
	}
}

func TestLatestDatePrefix_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	result := LatestDatePrefix(dir)
	if result != "" {
		t.Errorf("expected empty string for empty dir, got %q", result)
	}
}

func TestRenderMarkdown_AttachmentLinks(t *testing.T) {
	thread := makeThread()
	thread.Messages[0].Attachments = []domain.Attachment{
		{Name: "report.pdf", ContentType: "application/pdf", Content: []byte("pdf")},
		{Name: "data.xlsx", ContentType: "application/xlsx", Content: []byte("xlsx")},
	}
	md := RenderMarkdown(thread, "./attachments")
	if !strings.Contains(md, "**Attachments:** [report.pdf](./attachments/2025-01-14-12-00-report.pdf), [data.xlsx](./attachments/2025-01-14-12-00-data.xlsx)") {
		t.Errorf("expected attachment links, got:\n%s", md)
	}
}

func TestRenderMarkdown_QuoteReferenceHyperlinks(t *testing.T) {
	md := RenderMarkdown(makeThread(), "./attachments")
	expected := "[[→ quoted text omitted. see message #1 from Alice, 14 Jan]](#message-1)"
	if !strings.Contains(md, expected) {
		t.Errorf("expected quote reference hyperlink %q, got:\n%s", expected, md)
	}
}

func TestRenderText_AttachmentPaths(t *testing.T) {
	thread := makeThread()
	thread.Messages[0].Attachments = []domain.Attachment{
		{Name: "report.pdf", ContentType: "application/pdf", Content: []byte("pdf")},
	}
	txt := RenderText(thread, "./attachments")
	if !strings.Contains(txt, "Attachments: [./attachments/2025-01-14-12-00-report.pdf]") {
		t.Errorf("expected attachment path in text output, got:\n%s", txt)
	}
}

func TestSaveAttachments_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	thread.Messages[0].Attachments = []domain.Attachment{
		{Name: "doc.pdf", ContentType: "application/pdf", Content: []byte("content")},
	}
	if err := SaveAttachments(dir, thread); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	path := filepath.Join(dir, "attachments", "2025-01-14-12-00-doc.pdf")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("attachment file not found: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("expected 'content', got %q", string(data))
	}
}

func TestSaveAttachments_NoOpWithoutAttachments(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	if err := SaveAttachments(dir, thread); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	// attachments dir should not be created
	if _, err := os.Stat(filepath.Join(dir, "attachments")); !os.IsNotExist(err) {
		t.Error("expected no attachments dir when no attachments")
	}
}

func TestWriteFile_SavesAttachments(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	thread.Messages[0].Attachments = []domain.Attachment{
		{Name: "file.txt", ContentType: "text/plain", Content: []byte("hello")},
	}
	_, err := WriteFile(dir, thread, "text")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	path := filepath.Join(dir, "attachments", "2025-01-14-12-00-file.txt")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("attachment file not saved: %v", err)
	}
}

func TestConvertOutputToHTML_ConvertsMarkdownFiles(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	_, err := WriteFile(dir, thread, "markdown")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := ConvertOutputToHTML(dir); err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	// Markdown file should be gone
	mdFiles, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	if len(mdFiles) > 0 {
		t.Errorf("expected no .md files in root, found: %v", mdFiles)
	}

	// HTML file should exist
	htmlFiles, _ := filepath.Glob(filepath.Join(dir, "*.html"))
	htmlCount := 0
	for _, f := range htmlFiles {
		if filepath.Base(f) != "index.html" {
			htmlCount++
		}
	}
	if htmlCount != 1 {
		t.Errorf("expected 1 thread .html file, found %d", htmlCount)
	}

	// CSS should exist
	if _, err := os.Stat(filepath.Join(dir, "style.css")); err != nil {
		t.Error("expected style.css to be created")
	}

	// index.html should exist
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		t.Error("expected index.html to be created")
	}
}

func TestConvertOutputToHTML_PreservesAttachmentMarkdown(t *testing.T) {
	dir := t.TempDir()
	attDir := filepath.Join(dir, "attachments")
	os.MkdirAll(attDir, 0755)
	attFile := filepath.Join(attDir, "2026-03-19-15-00-walkthrough.md")
	os.WriteFile(attFile, []byte("# Walkthrough"), 0644)

	// Also write a thread md
	thread := makeThread()
	WriteFile(dir, thread, "markdown")

	if err := ConvertOutputToHTML(dir); err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	// Attachment .md should still exist
	if _, err := os.Stat(attFile); err != nil {
		t.Error("attachment .md should not be deleted")
	}

	// No .html should be in attachments/
	htmlInAtt, _ := filepath.Glob(filepath.Join(attDir, "*.html"))
	if len(htmlInAtt) > 0 {
		t.Errorf("expected no .html in attachments/, found: %v", htmlInAtt)
	}
}

func TestConvertOutputToHTML_SubdirFiles(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	thread.Subject = "Axiom Daily Activity Report 2026-02-05"
	_, err := WriteFile(dir, thread, "markdown")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := ConvertOutputToHTML(dir); err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	// HTML should exist in subdir (1 thread + 1 index.html)
	subdir := filepath.Join(dir, "daily-activity-reports")
	htmlFiles, _ := filepath.Glob(filepath.Join(subdir, "*.html"))
	if len(htmlFiles) != 2 {
		t.Errorf("expected 2 html in daily-activity-reports/ (1 thread + index), found %d", len(htmlFiles))
	}

	// .md should be removed from subdir
	mdFiles, _ := filepath.Glob(filepath.Join(subdir, "*.md"))
	if len(mdFiles) > 0 {
		t.Errorf("expected no .md in subdir, found: %v", mdFiles)
	}

	// HTML should reference ../style.css
	content, _ := os.ReadFile(htmlFiles[0])
	if !strings.Contains(string(content), "../style.css") {
		t.Error("expected ../style.css reference for subdir file")
	}
}

func TestGenerateIndex_GroupsByDay(t *testing.T) {
	dir := t.TempDir()

	// Create two HTML files with different dates
	page1 := wrapHTMLPage([]byte("<h1>Thread A</h1>"), "Thread A", "style.css", "")
	page2 := wrapHTMLPage([]byte("<h1>Thread B</h1>"), "Thread B", "style.css", "")
	os.WriteFile(filepath.Join(dir, "2026-04-03-09-18-thread-a.html"), page1, 0644)
	os.WriteFile(filepath.Join(dir, "2026-04-02-20-08-thread-b.html"), page2, 0644)
	os.WriteFile(filepath.Join(dir, "style.css"), []byte(cssContent), 0644)

	if err := generateIndex(dir); err != nil {
		t.Fatalf("generate index failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html failed: %v", err)
	}
	idx := string(content)

	if !strings.Contains(idx, "<summary>2026-04-03</summary>") {
		t.Error("expected day group 2026-04-03")
	}
	if !strings.Contains(idx, "<summary>2026-04-02</summary>") {
		t.Error("expected day group 2026-04-02")
	}
	if !strings.Contains(idx, "Expand All") {
		t.Error("expected Expand All button")
	}
	if !strings.Contains(idx, "thread-a.html") {
		t.Error("expected link to thread-a.html")
	}
}

func TestConvertOutputToHTML_HTMLContainsHeadingAnchors(t *testing.T) {
	dir := t.TempDir()
	thread := makeThread()
	WriteFile(dir, thread, "markdown")

	if err := ConvertOutputToHTML(dir); err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	htmlFiles, _ := filepath.Glob(filepath.Join(dir, "2025-*.html"))
	if len(htmlFiles) == 0 {
		t.Fatal("no html files found")
	}

	content, _ := os.ReadFile(htmlFiles[0])
	html := string(content)
	// Auto heading IDs should generate id="message-1" for ### Message #1
	if !strings.Contains(html, `id="message-1"`) {
		t.Error("expected heading anchor id=\"message-1\"")
	}
	// Quote reference hyperlink should link to #message-1
	if !strings.Contains(html, `href="#message-1"`) {
		t.Error("expected quote reference link href=\"#message-1\"")
	}
}
