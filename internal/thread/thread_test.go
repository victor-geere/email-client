package thread

import (
	"testing"
	"time"

	"github.com/victor/email-linearize/internal/domain"
)

func makeMsg(id, convID, subject string, received time.Time) domain.Message {
	return domain.Message{
		ID:               id,
		ConversationID:   convID,
		Subject:          subject,
		From:             domain.EmailAddress{Name: "Test", Address: "test@example.com"},
		ReceivedDateTime: received,
		Body:             domain.Body{ContentType: "text", Content: "body of " + id},
	}
}

func TestGroupByConversation_GroupsCorrectly(t *testing.T) {
	msgs := []domain.Message{
		makeMsg("1", "conv-a", "A", time.Now()),
		makeMsg("2", "conv-b", "B", time.Now()),
		makeMsg("3", "conv-a", "Re: A", time.Now()),
	}
	groups := GroupByConversation(msgs)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups["conv-a"]) != 2 {
		t.Errorf("expected 2 messages in conv-a, got %d", len(groups["conv-a"]))
	}
	if len(groups["conv-b"]) != 1 {
		t.Errorf("expected 1 message in conv-b, got %d", len(groups["conv-b"]))
	}
}

func TestBuildThread_SortsByDate(t *testing.T) {
	now := time.Now()
	msgs := []domain.Message{
		makeMsg("3", "conv", "C", now.Add(2*time.Hour)),
		makeMsg("1", "conv", "A", now),
		makeMsg("2", "conv", "B", now.Add(1*time.Hour)),
	}
	thread := BuildThread("conv", msgs)
	if thread.Messages[0].ID != "1" || thread.Messages[1].ID != "2" || thread.Messages[2].ID != "3" {
		t.Errorf("messages not sorted: %s, %s, %s", thread.Messages[0].ID, thread.Messages[1].ID, thread.Messages[2].ID)
	}
}

func TestBuildThread_SingleMessage(t *testing.T) {
	msgs := []domain.Message{
		makeMsg("1", "conv", "Subject", time.Now()),
	}
	thread := BuildThread("conv", msgs)
	if len(thread.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(thread.Messages))
	}
	if thread.Subject != "Subject" {
		t.Errorf("expected subject 'Subject', got '%s'", thread.Subject)
	}
}

func TestAssignOrdinals_NumbersSequentially(t *testing.T) {
	now := time.Now()
	thread := domain.Thread{
		ConversationID: "conv",
		Subject:        "Test",
		Messages: []domain.Message{
			makeMsg("a", "conv", "A", now),
			makeMsg("b", "conv", "B", now.Add(time.Hour)),
			makeMsg("c", "conv", "C", now.Add(2*time.Hour)),
		},
	}
	annotated := AssignOrdinals(thread)
	for i, am := range annotated {
		if am.Ordinal != i+1 {
			t.Errorf("message %d: expected ordinal %d, got %d", i, i+1, am.Ordinal)
		}
	}
}

func TestBuildThread_IdenticalTimestamps(t *testing.T) {
	now := time.Now()
	msgs := []domain.Message{
		makeMsg("b", "conv", "B", now),
		makeMsg("a", "conv", "A", now),
	}
	thread := BuildThread("conv", msgs)
	// With stable sort, original order is preserved for equal timestamps
	if len(thread.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(thread.Messages))
	}
}
