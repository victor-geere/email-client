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

func TestDetect_OutlookDateHeader(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content: "My reply\n\n" +
					"From: Jens von Bülow <jens@tnt.co.za>\n" +
					"Date: Tuesday, 10 March 2026 at 09:57\n" +
					"To: Neil Harvey <neil@tnt.co.za>\n" +
					"Subject: Agile...\n\n" +
					"Original message text here",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region from Outlook Date: header")
	}
	// The region should start at the "From:" line
	if regions[0].StartOffset != len("My reply\n\n") {
		t.Errorf("expected region to start at offset %d, got %d", len("My reply\n\n"), regions[0].StartOffset)
	}
}

func TestDetect_OutlookHeaderWithOptionalFields(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content: "I agree.\n\n" +
					"From: Gavin Harvey <gavin@tnt.co.za>\n" +
					"Sent: 21 March 2026 14:45\n" +
					"To: Neil Harvey <neil@tnt.co.za>; Jens von Bülow <jens@tnt.co.za>\n" +
					"Cc: Victor Geere <victor@kelevra.com>\n" +
					"Subject: Meeting with CapSource @\n" +
					"When: 24 March 2026 09:00-11:30.\n" +
					"Where: Atlantic Beach Links\n\n" +
					"Original message body here",
			},
		},
		Ordinal: 3,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region from header with optional Cc/When/Where fields")
	}
}

func TestDetect_MultipleQuotedMessagesInChain(t *testing.T) {
	earlier1 := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content:     "I have sent an invite to all making Neil required.",
			},
			From:             domain.EmailAddress{Name: "Gavin", Address: "gavin@tnt.co.za"},
			ReceivedDateTime: time.Date(2026, 3, 21, 12, 45, 0, 0, time.UTC),
		},
		Ordinal: 1,
	}
	earlier2 := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content: "I think I should attend.\n\n" +
					"From: Gavin <gavin@tnt.co.za>\n" +
					"Sent: 21 March 2026 14:45\n" +
					"To: Neil <neil@tnt.co.za>\n" +
					"Subject: Meeting\n\n" +
					"I have sent an invite to all making Neil required.",
			},
			From:             domain.EmailAddress{Name: "Kumesh", Address: "kumesh@kelevra.com"},
			ReceivedDateTime: time.Date(2026, 3, 23, 12, 45, 0, 0, time.UTC),
		},
		Ordinal: 2,
	}

	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content: "Lets see where this takes us\n\n" +
					"From: Kumesh <kumesh@kelevra.com>\n" +
					"Date: Monday, 23 March 2026 at 14:45\n" +
					"To: Gavin <gavin@tnt.co.za>\n" +
					"Subject: Re: Meeting\n\n" +
					"I think I should attend.\n\n" +
					"From: Gavin <gavin@tnt.co.za>\n" +
					"Sent: 21 March 2026 14:45\n" +
					"To: Neil <neil@tnt.co.za>\n" +
					"Subject: Meeting\n\n" +
					"I have sent an invite to all making Neil required.",
			},
		},
		Ordinal: 3,
	}

	regions := Detect(msg, []domain.AnnotatedMessage{earlier1, earlier2})
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region for nested quoted chain")
	}
	// The entire quoted chain from "From: Kumesh..." to end should be captured
	expectedStart := len("Lets see where this takes us\n\n")
	if regions[0].StartOffset != expectedStart {
		t.Errorf("expected region to start at offset %d, got %d", expectedStart, regions[0].StartOffset)
	}
}

func TestDetect_AttributesThroughHeaderPrefixedQuote(t *testing.T) {
	earlier := []domain.AnnotatedMessage{
		{
			Message: domain.Message{
				Body:             domain.Body{ContentType: "text", Content: "I want to understand why this duplication arose. Next time we meet."},
				From:             domain.EmailAddress{Name: "Neil Harvey", Address: "neil@tnt.co.za"},
				ReceivedDateTime: time.Date(2026, 3, 25, 11, 20, 0, 0, time.UTC),
			},
			Ordinal: 1,
		},
	}

	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content: "me too... I wonder how we find out.\n\n" +
					"From: Neil Harvey <neil@tnt.co.za>\n" +
					"Sent: Wednesday, 25 March 2026 13:20\n" +
					"To: Jens von Bülow <jens@tnt.co.za>\n" +
					"Subject: Re: Audit\n\n" +
					"I want to understand why this duplication arose. Next time we meet.",
			},
		},
		Ordinal: 2,
	}

	regions := Detect(msg, earlier)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region")
	}
	found := false
	for _, r := range regions {
		if r.AttributedToOrdinal != nil && *r.AttributedToOrdinal == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected header-prefixed quote to be attributed to message 1")
	}
}

func TestDetect_OutlookHeaderWithBcc(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			Body: domain.Body{
				ContentType: "text",
				Content: "My reply\n\n" +
					"From: Alice <alice@example.com>\n" +
					"Sent: Monday, Jan 14, 2025\n" +
					"To: Bob <bob@example.com>\n" +
					"Cc: Carol <carol@example.com>\n" +
					"Bcc: Dave <dave@example.com>\n" +
					"Subject: Budget Review\n\n" +
					"Original message text",
			},
		},
		Ordinal: 2,
	}
	regions := Detect(msg, nil)
	if len(regions) == 0 {
		t.Fatal("expected at least one quote region from header with Bcc field")
	}
	if regions[0].StartOffset != len("My reply\n\n") {
		t.Errorf("expected region to start at offset %d, got %d", len("My reply\n\n"), regions[0].StartOffset)
	}
}
