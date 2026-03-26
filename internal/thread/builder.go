package thread

import (
	"sort"

	"github.com/victor/email-linearize/internal/domain"
)

// GroupByConversation groups messages by their ConversationID.
func GroupByConversation(messages []domain.Message) map[string][]domain.Message {
	groups := make(map[string][]domain.Message)
	for _, msg := range messages {
		groups[msg.ConversationID] = append(groups[msg.ConversationID], msg)
	}
	return groups
}

// BuildThread creates a Thread from messages sharing a conversation ID.
// Messages are sorted by ReceivedDateTime ascending.
func BuildThread(conversationID string, messages []domain.Message) domain.Thread {
	sorted := make([]domain.Message, len(messages))
	copy(sorted, messages)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].ReceivedDateTime.Before(sorted[j].ReceivedDateTime)
	})

	subject := ""
	if len(sorted) > 0 {
		subject = sorted[0].Subject
	}

	return domain.Thread{
		ConversationID: conversationID,
		Subject:        subject,
		Messages:       sorted,
	}
}

// AssignOrdinals assigns sequential ordinal numbers to messages in a thread.
func AssignOrdinals(thread domain.Thread) []domain.AnnotatedMessage {
	annotated := make([]domain.AnnotatedMessage, len(thread.Messages))
	for i, msg := range thread.Messages {
		annotated[i] = domain.AnnotatedMessage{
			Message:      msg,
			Ordinal:      i + 1,
			QuoteRegions: nil,
			CleanBody:    msg.Body.Content,
		}
	}
	return annotated
}
