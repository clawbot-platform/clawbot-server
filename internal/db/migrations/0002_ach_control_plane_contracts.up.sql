ALTER TABLE runs
  ADD COLUMN IF NOT EXISTS run_type TEXT NOT NULL DEFAULT 'replay_run',
  ADD COLUMN IF NOT EXISTS execution_mode TEXT NOT NULL DEFAULT 'deterministic',
  ADD COLUMN IF NOT EXISTS repo TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS domain TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS dataset_refs_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS prompt_pack_version TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS rule_pack_version TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS model_profile TEXT NOT NULL DEFAULT 'ach-default',
  ADD COLUMN IF NOT EXISTS guardrail_profile TEXT NOT NULL DEFAULT 'ach-guardian-default',
  ADD COLUMN IF NOT EXISTS memory_namespace_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS requested_by TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS artifact_bundle_refs_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS review_metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS memory_snapshot_refs_json JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE runs
SET
  finished_at = COALESCE(finished_at, completed_at),
  requested_by = CASE WHEN requested_by = '' THEN created_by ELSE requested_by END
WHERE completed_at IS NOT NULL OR created_by <> '';

CREATE INDEX IF NOT EXISTS idx_runs_run_type ON runs (run_type);
CREATE INDEX IF NOT EXISTS idx_runs_execution_mode ON runs (execution_mode);

CREATE TABLE IF NOT EXISTS model_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  provider TEXT NOT NULL,
  base_url TEXT NOT NULL DEFAULT '',
  primary_model TEXT NOT NULL,
  guardrail_model TEXT NOT NULL DEFAULT '',
  helper_model TEXT NOT NULL DEFAULT '',
  timeout_seconds INT NOT NULL DEFAULT 45,
  temperature DOUBLE PRECISION NOT NULL DEFAULT 0.1,
  max_tokens INT NOT NULL DEFAULT 4096,
  json_mode BOOLEAN NOT NULL DEFAULT TRUE,
  structured_output BOOLEAN NOT NULL DEFAULT TRUE,
  enable_guardrails BOOLEAN NOT NULL DEFAULT TRUE,
  enable_helper_model BOOLEAN NOT NULL DEFAULT TRUE,
  connection_metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO model_profiles (
  name,
  provider,
  base_url,
  primary_model,
  guardrail_model,
  helper_model,
  timeout_seconds,
  temperature,
  max_tokens,
  json_mode,
  structured_output,
  enable_guardrails,
  enable_helper_model,
  connection_metadata_json,
  created_by
)
VALUES (
  'ach-default',
  'local_ollama',
  'http://ai-precision:11434',
  'ibm/granite3.3:8b',
  'ibm/granite3.3-guardian:8b',
  'granite4:3b',
  45,
  0.1,
  4096,
  TRUE,
  TRUE,
  TRUE,
  TRUE,
  '{}'::jsonb,
  'system'
)
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS run_cycles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_id UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  cycle_key TEXT NOT NULL,
  focus TEXT NOT NULL DEFAULT '',
  objective TEXT NOT NULL DEFAULT '',
  detector_pack TEXT NOT NULL DEFAULT '',
  summary_ref TEXT NOT NULL DEFAULT '',
  carry_forward_summary_ref TEXT NOT NULL DEFAULT '',
  memory_snapshot_ref TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  started_at TIMESTAMPTZ NULL,
  finished_at TIMESTAMPTZ NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (run_id, cycle_key)
);

CREATE INDEX IF NOT EXISTS idx_run_cycles_run_id ON run_cycles (run_id);
CREATE INDEX IF NOT EXISTS idx_run_cycles_status ON run_cycles (status);

CREATE TABLE IF NOT EXISTS run_artifacts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_id UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  cycle_id UUID NULL REFERENCES run_cycles(id) ON DELETE SET NULL,
  artifact_type TEXT NOT NULL,
  uri TEXT NOT NULL,
  content_type TEXT NOT NULL DEFAULT 'application/json',
  version TEXT NOT NULL DEFAULT '',
  checksum TEXT NOT NULL DEFAULT '',
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_run_artifacts_run_id ON run_artifacts (run_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_run_artifacts_cycle_id ON run_artifacts (cycle_id, created_at DESC);

CREATE TABLE IF NOT EXISTS run_comparisons (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_id UUID NOT NULL UNIQUE REFERENCES runs(id) ON DELETE CASCADE,
  cycle_id UUID NULL REFERENCES run_cycles(id) ON DELETE SET NULL,
  deterministic_summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  llm_summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  guardrail_summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  deltas_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  review_status TEXT NOT NULL DEFAULT 'review_pending',
  reviewer_notes TEXT NOT NULL DEFAULT '',
  final_disposition TEXT NOT NULL DEFAULT '',
  final_output_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
