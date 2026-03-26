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
				CleanBody: "Looks good to me!\n\n[→ see msg #1 from Alice, 14 Jan]",
			},
		},
	}
}

func TestRenderMarkdown_StructureCorrect(t *testing.T) {
	md := RenderMarkdown(makeThread())

	if !strings.Contains(md, "# Thread: Q4 Budget Review") {
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
	md := RenderMarkdown(thread)

	// Check that Cc line only appears once (for message #1)
	count := strings.Count(md, "**Cc:**")
	if count != 1 {
		t.Errorf("expected 1 Cc line, got %d", count)
	}
}

func TestRenderText_NoMarkdownSyntax(t *testing.T) {
	txt := RenderText(makeThread())
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
