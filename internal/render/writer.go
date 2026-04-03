package render

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/victor/email-linearize/internal/domain"
)

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a string to a URL-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
		s = strings.TrimRight(s, "-")
	}
	return s
}

// CleanOutputDir removes all files (not directories) in dir and its subdirectories.
func CleanOutputDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			return os.Remove(path)
		}
		return nil
	})
}

// specialSubdir returns the subdirectory name and whether to force markdown
// for threads whose slug matches a routing rule.
func specialSubdir(slug string) (subdir string, forceMarkdown bool) {
	if strings.Contains(slug, "daily-activity-report") {
		return "daily-activity-reports", true
	}
	if strings.Contains(slug, "sprint-completion-report") {
		return "sprint-completion-reports", true
	}
	return "", false
}

// latestMessagePrefix returns a "yyyy-MM-dd-HH-mm-" prefix derived from the
// latest ReceivedDateTime among the thread's messages.
func latestMessagePrefix(thread domain.LinearizedThread) string {
	if len(thread.Messages) == 0 {
		return ""
	}
	latest := thread.Messages[0].ReceivedDateTime
	for _, msg := range thread.Messages[1:] {
		if msg.ReceivedDateTime.After(latest) {
			latest = msg.ReceivedDateTime
		}
	}
	return latest.In(SAST).Format("2006-01-02-15-04-")
}

// AttachmentFilename returns the dated filename for an attachment.
func AttachmentFilename(msgTime time.Time, attName string) string {
	prefix := msgTime.In(SAST).Format("2006-01-02-15-04-")
	return prefix + attName
}

// SaveAttachments writes all message attachments to outputDir/attachments/.
func SaveAttachments(outputDir string, thread domain.LinearizedThread) error {
	hasAny := false
	for _, msg := range thread.Messages {
		if len(msg.Attachments) > 0 {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return nil
	}

	attachDir := filepath.Join(outputDir, "attachments")
	if err := os.MkdirAll(attachDir, 0755); err != nil {
		return fmt.Errorf("create attachments directory: %w", err)
	}

	for _, msg := range thread.Messages {
		for _, att := range msg.Attachments {
			filename := AttachmentFilename(msg.ReceivedDateTime, att.Name)
			path := filepath.Join(attachDir, filename)
			if err := os.WriteFile(path, att.Content, 0644); err != nil {
				return fmt.Errorf("write attachment %q: %w", att.Name, err)
			}
		}
	}
	return nil
}

// WriteFile writes a rendered thread to disk with collision safety.
// Files are prefixed with the latest message timestamp. Threads matching
// special routing rules are placed in subdirectories as markdown.
func WriteFile(outputDir string, thread domain.LinearizedThread, format string) (string, error) {
	slug := Slugify(thread.Subject)
	if slug == "" {
		slug = "thread"
	}

	// Determine target directory and extension
	targetDir := outputDir
	ext := ".md"
	if format == "text" {
		ext = ".txt"
	}

	if subdir, forceMd := specialSubdir(slug); subdir != "" {
		targetDir = filepath.Join(outputDir, subdir)
		if forceMd {
			ext = ".md"
		}
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	// Save attachments
	if err := SaveAttachments(outputDir, thread); err != nil {
		return "", fmt.Errorf("save attachments: %w", err)
	}

	// Compute relative path from target directory to attachments dir
	attachDir := "./attachments"
	if targetDir != outputDir {
		attachDir = "../attachments"
	}

	// Render content — use report mode for structured reports, markdown when
	// forced by subdirectory routing, text otherwise
	var content string
	if isReport(slug) {
		content = RenderReport(thread, attachDir)
	} else if ext == ".md" {
		content = RenderMarkdown(thread, attachDir)
	} else {
		content = RenderText(thread, attachDir)
	}

	prefix := latestMessagePrefix(thread)
	filename := prefix + slug + ext
	path := filepath.Join(targetDir, filename)
	path = collisionSafe(path)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return path, nil
}

func collisionSafe(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// datePrefixRe matches the "yyyy-MM-dd-HH-mm-" prefix on output filenames.
var datePrefixRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}-\d{2}-\d{2})-`)

// LatestDatePrefix scans all files in outputDir (and subdirs) and returns the
// latest date prefix as an ISO 8601 datetime string suitable for Graph API
// $filter (e.g. "2026-03-25T11:05:00Z"). Returns "" if no dated files found.
func LatestDatePrefix(outputDir string) string {
	var latest time.Time

	_ = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		match := datePrefixRe.FindStringSubmatch(base)
		if match == nil {
			return nil
		}
		t, err := time.ParseInLocation("2006-01-02-15-04", match[1], SAST)
		if err != nil {
			return nil
		}
		if t.After(latest) {
			latest = t
		}
		return nil
	})

	if latest.IsZero() {
		return ""
	}
	return latest.UTC().Format("2006-01-02T15:04:05Z")
}
