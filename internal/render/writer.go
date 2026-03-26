package render

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

// WriteFile writes a rendered thread to disk with collision safety.
func WriteFile(outputDir string, thread domain.LinearizedThread, format string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	slug := Slugify(thread.Subject)
	if slug == "" {
		slug = "thread"
	}

	ext := ".md"
	if format == "text" {
		ext = ".txt"
	}

	var content string
	switch format {
	case "text":
		content = RenderText(thread)
	default:
		content = RenderMarkdown(thread)
	}

	path := filepath.Join(outputDir, slug+ext)
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
