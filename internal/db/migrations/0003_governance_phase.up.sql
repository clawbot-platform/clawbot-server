ALTER TABLE runs
  ADD COLUMN IF NOT EXISTS execution_ring TEXT NOT NULL DEFAULT 'ring_1',
  ADD COLUMN IF NOT EXISTS guardrail_status TEXT NOT NULL DEFAULT 'guardrail_disabled';

ALTER TABLE run_cycles
  ADD COLUMN IF NOT EXISTS execution_ring TEXT NOT NULL DEFAULT 'ring_1',
  ADD COLUMN IF NOT EXISTS guardrail_status TEXT NOT NULL DEFAULT 'guardrail_disabled';

CREATE INDEX IF NOT EXISTS idx_runs_execution_ring ON runs (execution_ring);
CREATE INDEX IF NOT EXISTS idx_run_cycles_execution_ring ON run_cycles (execution_ring);

CREATE TABLE IF NOT EXISTS policy_decisions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  action_type TEXT NOT NULL,
  target_run_id UUID NULL REFERENCES runs(id) ON DELETE SET NULL,
  target_cycle_id UUID NULL REFERENCES run_cycles(id) ON DELETE SET NULL,
  actor_id TEXT NOT NULL DEFAULT '',
  actor_type TEXT NOT NULL DEFAULT 'user',
  policy_input_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  allow BOOLEAN NOT NULL,
  policy_bundle_id TEXT NOT NULL DEFAULT '',
  policy_bundle_version TEXT NOT NULL DEFAULT '',
  reason_code TEXT NOT NULL DEFAULT '',
  conditions_applied_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  fallback_mode TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policy_decisions_target_run ON policy_decisions (target_run_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_policy_decisions_action ON policy_decisions (action_type, created_at DESC);

CREATE TABLE IF NOT EXISTS run_review_actions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_id UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  cycle_id UUID NULL REFERENCES run_cycles(id) ON DELETE SET NULL,
  reviewer_id TEXT NOT NULL,
  reviewer_type TEXT NOT NULL DEFAULT 'human',
  action_type TEXT NOT NULL,
  prior_status TEXT NOT NULL,
  new_status TEXT NOT NULL,
  rationale TEXT NOT NULL DEFAULT '',
  policy_decision_id UUID NULL REFERENCES policy_decisions(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_run_review_actions_run ON run_review_actions (run_id, created_at DESC);

CREATE TABLE IF NOT EXISTS governance_audit_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  previous_event_hash TEXT NOT NULL DEFAULT '',
  current_event_hash TEXT NOT NULL,
  actor_id TEXT NOT NULL DEFAULT '',
  actor_type TEXT NOT NULL DEFAULT 'user',
  action_type TEXT NOT NULL,
  target_run_id UUID NULL REFERENCES runs(id) ON DELETE SET NULL,
  target_cycle_id UUID NULL REFERENCES run_cycles(id) ON DELETE SET NULL,
  target_artifact_id UUID NULL REFERENCES run_artifacts(id) ON DELETE SET NULL,
  policy_decision_id UUID NULL REFERENCES policy_decisions(id) ON DELETE SET NULL,
  payload_summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_governance_audit_events_run ON governance_audit_events (target_run_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_governance_audit_events_hash ON governance_audit_events (current_event_hash);
