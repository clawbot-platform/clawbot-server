DROP TABLE IF EXISTS run_comparisons;
DROP TABLE IF EXISTS run_artifacts;
DROP TABLE IF EXISTS run_cycles;
DROP TABLE IF EXISTS model_profiles;

DROP INDEX IF EXISTS idx_runs_execution_mode;
DROP INDEX IF EXISTS idx_runs_run_type;

ALTER TABLE runs
  DROP COLUMN IF EXISTS memory_snapshot_refs_json,
  DROP COLUMN IF EXISTS notes,
  DROP COLUMN IF EXISTS review_metadata_json,
  DROP COLUMN IF EXISTS artifact_bundle_refs_json,
  DROP COLUMN IF EXISTS finished_at,
  DROP COLUMN IF EXISTS requested_by,
  DROP COLUMN IF EXISTS memory_namespace_json,
  DROP COLUMN IF EXISTS guardrail_profile,
  DROP COLUMN IF EXISTS model_profile,
  DROP COLUMN IF EXISTS rule_pack_version,
  DROP COLUMN IF EXISTS prompt_pack_version,
  DROP COLUMN IF EXISTS dataset_refs_json,
  DROP COLUMN IF EXISTS domain,
  DROP COLUMN IF EXISTS repo,
  DROP COLUMN IF EXISTS execution_mode,
  DROP COLUMN IF EXISTS run_type;
