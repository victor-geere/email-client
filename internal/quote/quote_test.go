package quote

import (
	"testing"
	"time"

	"github.com/victor/email-linearize/internal/domain"
)

func TestStripHTML_PreservesText(t *testing.T) {
	html := "<html><body><p>Hello world</p><p>Second paragraph</p></body></html>"
	result := StripHTML(html)
	if result != "Hello world\nSecond paragraph" {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestStripHTML_IgnoresScripts(t *testing.T) {
	html := "<html><body><script>alert('xss')</script><p>Safe text</p></body></html>"
	result := StripHTML(html)
	if result != "Safe text" {
		t.Errorf("expected 'Safe text', got %q", result)
	}
}

func TestStripHTML_HandlesBr(t *testing.T) {
	html := "<p>Line one<br>Line two</p>"
	result := StripHTML(html)
	if result != "Line one\nLine two" {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestDetect_PlainTextPrefix(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content:     "My reply\n\n> Original message line 1\n> Original message line 2\n\nAnother point",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region")
	}
}

func TestDetect_HTMLBlockquote(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "html",
				Content:     "<html><body><p>My reply</p><blockquote>Quoted original text here</blockquote></body></html>",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region from blockquote")
	}
}

func TestDetect_OutlookHeader(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content:     "My reply\n\nFrom: Alice Wang\nSent: Monday, Jan 14, 2025\nSubject: Budget Review\n\nOriginal message text",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region from Outlook header")
	}
}

func TestDetect_GmailHeader(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content:     "My reply\n\nOn Jan 14, 2025, Alice Wang wrote:\nOriginal message text",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region from Gmail header")
	}
}

func TestDetect_AttributesToCorrectMessage(t *testing.T) {
	earlier := []domain.AnnotatedMessage{
		{
			Message: domain.Message{
				Body:             domain.Body{ContentType: "text", Content: "This is the original message with enough text to match reliably"},
				From:             domain.EmailAddress{Name: "Alice", Address: "alice@example.com"},
				ReceivedDateTime: time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC),
			},
			Ordinal: 1,
		},
	}

	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content:     "My reply\n\n> This is the original message with enough text to match reliably",
			},
		},
		Ordinal: 2,
	}

	regions := Detect(msg, earlier)
	if len(regions) == 0 {
		t.Fatal("expected at least one region")
	}
	found := false
	for _, r := range regions {
		if r.AttributedToOrdinal != nil && *r.AttributedToOrdinal == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected quote to be attributed to message 1")
	}
}

func TestDetect_UnattributableQuote(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content:     "My reply\n\n> Some quote that doesn't match any earlier message text whatsoever",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	for _, r := range regions {
		if r.AttributedToOrdinal != nil {
			t.Error("expected unattributable quote to have nil AttributedToOrdinal")
		}
	}
}
