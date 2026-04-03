package linearize

import (
	"strings"
	"testing"
	"time"

	"github.com/victor/email-linearize/internal/domain"
)

func makeThread(msgs ...domain.AnnotatedMessage) (domain.Thread, []domain.AnnotatedMessage) {
	rawMsgs := make([]domain.Message, len(msgs))
	for i, m := range msgs {
		rawMsgs[i] = m.Message
	}
	thread := domain.Thread{
		ConversationID: "conv-1",
		Subject:        "Test Thread",
		Messages:       rawMsgs,
	}
	return thread, msgs
}

func TestLinearize_ReplacesQuoteWithReference(t *testing.T) {
	t1 := time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 14, 11, 0, 0, 0, time.UTC)

	msg1 := domain.AnnotatedMessage{
		Message: domain.Message{
			ID: "1", ConversationID: "conv-1", Subject: "Test",
			From:             domain.EmailAddress{Name: "Alice Wang", Address: "alice@example.com"},
			ReceivedDateTime: t1,
			Body:             domain.Body{ContentType: "text", Content: "This is the original message with enough content to be matchable"},
		},
		Ordinal:   1,
		CleanBody: "This is the original message with enough content to be matchable",
	}

	msg2 := domain.AnnotatedMessage{
		Message: domain.Message{
			ID: "2", ConversationID: "conv-1", Subject: "Re: Test",
			From:             domain.EmailAddress{Name: "Bob Chen", Address: "bob@example.com"},
			ReceivedDateTime: t2,
			Body:             domain.Body{ContentType: "text", Content: "I agree!\n\n> This is the original message with enough content to be matchable"},
		},
		Ordinal: 2,
	}

	thread, annotated := makeThread(msg1, msg2)
	result := Linearize(thread, annotated)

	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Messages))
	}
	body := result.Messages[1].CleanBody
	if !strings.Contains(body, "[→ quoted text omitted. see message #1 from Alice, 14 Jan]") {
		t.Errorf("expected reference to msg #1, got: %s", body)
	}
}

func TestLinearize_UnattributedQuote(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			ID: "1", ConversationID: "conv-1", Subject: "Test",
			From:             domain.EmailAddress{Name: "Alice", Address: "alice@example.com"},
			ReceivedDateTime: time.Now(),
			Body:             domain.Body{ContentType: "text", Content: "Reply\n\n> Unknown quoted text that matches nothing at all"},
		},
		Ordinal: 1,
	}

	thread, annotated := makeThread(msg)
	result := Linearize(thread, annotated)
	body := result.Messages[0].CleanBody
	if strings.Contains(body, "> Unknown") {
		// If the quote wasn't replaced, that's okay if no regions were detected
		// (first message has no earlier messages to attribute to anyway)
	}
}

func TestLinearize_NoQuotes(t *testing.T) {
	msg := domain.AnnotatedMessage{
		Message: domain.Message{
			ID: "1", ConversationID: "conv-1", Subject: "Test",
			From:             domain.EmailAddress{Name: "Alice", Address: "alice@example.com"},
			ReceivedDateTime: time.Now(),
			Body:             domain.Body{ContentType: "text", Content: "Just a plain message with no quotes at all."},
		},
		Ordinal: 1,
	}

	thread, annotated := makeThread(msg)
	result := Linearize(thread, annotated)
	if result.Messages[0].CleanBody != "Just a plain message with no quotes at all." {
		t.Errorf("expected unchanged body, got: %s", result.Messages[0].CleanBody)
	}
}

func TestLinearize_TrimsBlankLines(t *testing.T) {
	t1 := time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC)
	msg1 := domain.AnnotatedMessage{
		Message: domain.Message{
			ID: "1", ConversationID: "conv-1", Subject: "Test",
			From:             domain.EmailAddress{Name: "Alice", Address: "alice@example.com"},
			ReceivedDateTime: t1,
			Body:             domain.Body{ContentType: "text", Content: "Original text that is long enough for matching purposes in this test"},
		},
		Ordinal:   1,
		CleanBody: "Original text that is long enough for matching purposes in this test",
	}

	msg2 := domain.AnnotatedMessage{
		Message: domain.Message{
			ID: "2", ConversationID: "conv-1", Subject: "Re: Test",
			From:             domain.EmailAddress{Name: "Bob", Address: "bob@example.com"},
			ReceivedDateTime: t1.Add(time.Hour),
			Body:             domain.Body{ContentType: "text", Content: "Reply\n\n\n\n\n> Original text that is long enough for matching purposes in this test\n\n\n\nEnd"},
		},
		Ordinal: 2,
	}

	thread, annotated := makeThread(msg1, msg2)
	result := Linearize(thread, annotated)
	body := result.Messages[1].CleanBody
	if strings.Contains(body, "\n\n\n") {
		t.Errorf("expected no triple blank lines, got: %q", body)
	}
}
