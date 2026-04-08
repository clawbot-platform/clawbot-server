DROP INDEX IF EXISTS idx_governance_audit_events_hash;
DROP INDEX IF EXISTS idx_governance_audit_events_run;
DROP TABLE IF EXISTS governance_audit_events;

DROP INDEX IF EXISTS idx_run_review_actions_run;
DROP TABLE IF EXISTS run_review_actions;

DROP INDEX IF EXISTS idx_policy_decisions_action;
DROP INDEX IF EXISTS idx_policy_decisions_target_run;
DROP TABLE IF EXISTS policy_decisions;

DROP INDEX IF EXISTS idx_run_cycles_execution_ring;
DROP INDEX IF EXISTS idx_runs_execution_ring;

ALTER TABLE run_cycles
  DROP COLUMN IF EXISTS guardrail_status,
  DROP COLUMN IF EXISTS execution_ring;

ALTER TABLE runs
  DROP COLUMN IF EXISTS guardrail_status,
  DROP COLUMN IF EXISTS execution_ring;
