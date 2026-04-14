package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

const (
	SubjectCompareCompleted       = "clawbot.identity.compare.completed.v1"
	SubjectOFACScreeningCompleted = "clawbot.watchlist.ofac.screening.completed.v1"
	SubjectOFACScreeningReview    = "clawbot.watchlist.ofac.screening.review.v1"
)

type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	TenantID      string          `json:"tenant_id"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	CaseID        string          `json:"case_id,omitempty"`
	OccurredAt    string          `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

type Event[T any] struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	TenantID      string `json:"tenant_id"`
	CorrelationID string `json:"correlation_id,omitempty"`
	CaseID        string `json:"case_id,omitempty"`
	OccurredAt    string `json:"occurred_at"`
	Payload       T      `json:"payload"`
}

type CompareCompletedPayload struct {
	TenantID            string `json:"tenant_id"`
	CaseID              string `json:"case_id,omitempty"`
	CorrelationID       string `json:"correlation_id,omitempty"`
	DecisionTraceID     string `json:"decision_trace_id"`
	ExplanationID       string `json:"explanation_id,omitempty"`
	Disposition         string `json:"disposition"`
	ConfidenceBand      string `json:"confidence_band,omitempty"`
	LeftSourceSystem    string `json:"left_source_system,omitempty"`
	LeftSourceRecordID  string `json:"left_source_record_id,omitempty"`
	RightSourceSystem   string `json:"right_source_system,omitempty"`
	RightSourceRecordID string `json:"right_source_record_id,omitempty"`
}

type OFACScreeningPayload struct {
	TenantID          string  `json:"tenant_id"`
	CaseID            string  `json:"case_id,omitempty"`
	CorrelationID     string  `json:"correlation_id,omitempty"`
	ScreeningID       string  `json:"screening_id"`
	Decision          string  `json:"decision"`
	DecisionTraceID   string  `json:"decision_trace_id,omitempty"`
	ExplanationID     string  `json:"explanation_id,omitempty"`
	SubjectName       string  `json:"subject_name"`
	CandidateCount    int     `json:"candidate_count"`
	TopCandidateName  string  `json:"top_candidate_name,omitempty"`
	TopCandidateScore float64 `json:"top_candidate_score,omitempty"`
}

type Handlers struct {
	OnCompareCompleted       func(context.Context, Event[CompareCompletedPayload]) error
	OnOFACScreeningCompleted func(context.Context, Event[OFACScreeningPayload]) error
	OnOFACScreeningReview    func(context.Context, Event[OFACScreeningPayload]) error
}

type Subscriber interface {
	Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error)
}

type Consumer struct {
	logger *slog.Logger
	subs   []*nats.Subscription
}

func NewConsumer(subscriber Subscriber, logger *slog.Logger, handlers Handlers) (*Consumer, error) {
	if subscriber == nil {
		return nil, fmt.Errorf("subscriber is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	c := &Consumer{logger: logger}

	if err := subscribe(c, subscriber, SubjectCompareCompleted, handlers.OnCompareCompleted); err != nil {
		return nil, err
	}
	if err := subscribe(c, subscriber, SubjectOFACScreeningCompleted, handlers.OnOFACScreeningCompleted); err != nil {
		return nil, err
	}
	if err := subscribe(c, subscriber, SubjectOFACScreeningReview, handlers.OnOFACScreeningReview); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) Close() error {
	var firstErr error
	for _, sub := range c.subs {
		if sub == nil {
			continue
		}
		if err := sub.Unsubscribe(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.subs = nil
	return firstErr
}

func subscribe[T any](c *Consumer, subscriber Subscriber, subject string, handler func(context.Context, Event[T]) error) error {
	sub, err := subscriber.Subscribe(subject, func(msg *nats.Msg) {
		event, decodeErr := decodeEvent[T](msg.Data)
		if decodeErr != nil {
			c.logger.Error("identity.event.decode_failed", "subject", subject, "error", decodeErr)
			return
		}

		if handler == nil {
			return
		}

		if err := handler(context.Background(), event); err != nil {
			c.logger.Error("identity.event.handler_failed", "subject", subject, "event_id", event.EventID, "error", err)
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", subject, err)
	}

	c.subs = append(c.subs, sub)
	return nil
}

func decodeEvent[T any](raw []byte) (Event[T], error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return Event[T]{}, fmt.Errorf("decode envelope: %w", err)
	}

	var payload T
	if len(env.Payload) > 0 {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return Event[T]{}, fmt.Errorf("decode envelope payload: %w", err)
		}
	}

	return Event[T]{
		EventID:       env.EventID,
		EventType:     env.EventType,
		TenantID:      env.TenantID,
		CorrelationID: env.CorrelationID,
		CaseID:        env.CaseID,
		OccurredAt:    env.OccurredAt,
		Payload:       payload,
	}, nil
}
