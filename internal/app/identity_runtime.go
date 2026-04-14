package app

import (
	"context"
	"log/slog"
	"strings"

	"clawbot-server/internal/config"
	identityevents "clawbot-server/internal/events/identity"

	"github.com/nats-io/nats.go"
)

type identityRuntime struct {
	logger   *slog.Logger
	consumer *identityevents.Consumer
	natsConn *nats.Conn
}

func startIdentityEventRuntime(cfg config.Server, logger *slog.Logger) *identityRuntime {
	if logger == nil {
		logger = slog.Default()
	}

	natsURL := strings.TrimSpace(cfg.NATSURL)
	if natsURL == "" {
		logger.Info("identity.events.disabled", "reason", "missing_nats_url")
		return &identityRuntime{logger: logger}
	}

	natsConn, err := nats.Connect(natsURL, nats.Name("clawbot-server.identity-events"))
	if err != nil {
		logger.Warn("identity.events.nats_connect_failed", "nats_url", natsURL, "error", err)
		return &identityRuntime{logger: logger}
	}

	consumer, err := identityevents.NewConsumer(natsConn, logger, identityevents.Handlers{
		OnCompareCompleted: func(_ context.Context, event identityevents.Event[identityevents.CompareCompletedPayload]) error {
			logger.Info(
				"identity.event.compare_completed",
				"event_id", event.EventID,
				"tenant_id", event.TenantID,
				"case_id", event.CaseID,
				"decision_trace_id", event.Payload.DecisionTraceID,
				"disposition", event.Payload.Disposition,
				"confidence_band", event.Payload.ConfidenceBand,
			)
			return nil
		},
		OnOFACScreeningCompleted: func(_ context.Context, event identityevents.Event[identityevents.OFACScreeningPayload]) error {
			logger.Info(
				"identity.event.ofac_screening_completed",
				"event_id", event.EventID,
				"tenant_id", event.TenantID,
				"case_id", event.CaseID,
				"screening_id", event.Payload.ScreeningID,
				"decision", event.Payload.Decision,
				"top_candidate_name", event.Payload.TopCandidateName,
				"top_candidate_score", event.Payload.TopCandidateScore,
			)
			return nil
		},
		OnOFACScreeningReview: func(_ context.Context, event identityevents.Event[identityevents.OFACScreeningPayload]) error {
			logger.Info(
				"identity.event.ofac_screening_review",
				"event_id", event.EventID,
				"tenant_id", event.TenantID,
				"case_id", event.CaseID,
				"screening_id", event.Payload.ScreeningID,
				"decision", event.Payload.Decision,
				"top_candidate_name", event.Payload.TopCandidateName,
				"top_candidate_score", event.Payload.TopCandidateScore,
			)
			return nil
		},
	})
	if err != nil {
		logger.Warn("identity.events.subscribe_failed", "error", err)
		natsConn.Close()
		return &identityRuntime{logger: logger}
	}

	logger.Info(
		"identity.events.subscribed",
		"nats_url", natsURL,
		"subjects", []string{
			identityevents.SubjectCompareCompleted,
			identityevents.SubjectOFACScreeningCompleted,
			identityevents.SubjectOFACScreeningReview,
		},
	)

	return &identityRuntime{
		logger:   logger,
		consumer: consumer,
		natsConn: natsConn,
	}
}

func (r *identityRuntime) Close() {
	if r == nil {
		return
	}

	if r.consumer != nil {
		if err := r.consumer.Close(); err != nil {
			r.logger.Warn("identity.events.unsubscribe_failed", "error", err)
		}
	}

	if r.natsConn != nil {
		if err := r.natsConn.Drain(); err != nil {
			r.logger.Warn("identity.events.drain_failed", "error", err)
		}
		r.natsConn.Close()
	}
}
