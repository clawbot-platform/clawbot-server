package runs

import (
	"context"
	"encoding/json"
	"time"

	"clawbot-server/internal/platform/store"
)

type RunType string

type ExecutionMode string

type ExecutionRing string

type RunStatus string

type CycleStatus string

type ReviewStatus string

type GuardrailStatus string

const (
	RunTypeReplayRun RunType = "replay_run"
	RunTypeAgentRun  RunType = "agent_run"
	RunTypeWeekRun   RunType = "week_run"
)

const (
	ExecutionModeDeterministic ExecutionMode = "deterministic"
	ExecutionModeLLM           ExecutionMode = "llm"
	ExecutionModeDual          ExecutionMode = "dual"
)

const (
	ExecutionRing0 ExecutionRing = "ring_0"
	ExecutionRing1 ExecutionRing = "ring_1"
	ExecutionRing2 ExecutionRing = "ring_2"
	ExecutionRing3 ExecutionRing = "ring_3"
)

const (
	RunStatusPending           RunStatus = "pending"
	RunStatusScheduled         RunStatus = "scheduled"
	RunStatusRunning           RunStatus = "running"
	RunStatusGuardrailDeferred RunStatus = "guardrail_deferred"
	RunStatusFailedRuntime     RunStatus = "failed_runtime"
	RunStatusFailedPolicy      RunStatus = "failed_policy"
	RunStatusReviewPending     RunStatus = "review_pending"
	RunStatusApproved          RunStatus = "approved"
	RunStatusRejected          RunStatus = "rejected"
	RunStatusOverridden        RunStatus = "overridden"
	RunStatusDeferred          RunStatus = "deferred"
	RunStatusCompleted         RunStatus = "completed"
	RunStatusFailed            RunStatus = "failed"
	RunStatusCancelled         RunStatus = "cancelled"
	RunStatusClosed            RunStatus = "closed"
)

const (
	CycleStatusPending           CycleStatus = "pending"
	CycleStatusRunning           CycleStatus = "running"
	CycleStatusGuardrailDeferred CycleStatus = "guardrail_deferred"
	CycleStatusFailedRuntime     CycleStatus = "failed_runtime"
	CycleStatusFailedPolicy      CycleStatus = "failed_policy"
	CycleStatusReviewPending     CycleStatus = "review_pending"
	CycleStatusApproved          CycleStatus = "approved"
	CycleStatusRejected          CycleStatus = "rejected"
	CycleStatusOverridden        CycleStatus = "overridden"
	CycleStatusDeferred          CycleStatus = "deferred"
	CycleStatusCompleted         CycleStatus = "completed"
	CycleStatusFailed            CycleStatus = "failed"
	CycleStatusCancelled         CycleStatus = "cancelled"
)

const (
	ReviewStatusReviewPending ReviewStatus = "review_pending"
	ReviewStatusApproved      ReviewStatus = "approved"
	ReviewStatusRejected      ReviewStatus = "rejected"
	ReviewStatusOverridden    ReviewStatus = "overridden"
)

const (
	GuardrailStatusPassed      GuardrailStatus = "guardrail_passed"
	GuardrailStatusFlagged     GuardrailStatus = "guardrail_flagged"
	GuardrailStatusTimeout     GuardrailStatus = "guardrail_timeout"
	GuardrailStatusUnavailable GuardrailStatus = "guardrail_unavailable"
	GuardrailStatusDisabled    GuardrailStatus = "guardrail_disabled"
)

type MemoryNamespace struct {
	RepoNamespace  string `json:"repo_namespace"`
	RunNamespace   string `json:"run_namespace"`
	CycleNamespace string `json:"cycle_namespace,omitempty"`
	AgentNamespace string `json:"agent_namespace,omitempty"`
}

type Run struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Status             string          `json:"status"`
	ScenarioType       string          `json:"scenario_type"`
	RunType            string          `json:"run_type"`
	ExecutionMode      string          `json:"execution_mode"`
	ExecutionRing      string          `json:"execution_ring"`
	GuardrailStatus    string          `json:"guardrail_status"`
	Repo               string          `json:"repo"`
	Domain             string          `json:"domain"`
	DatasetRefs        []string        `json:"dataset_refs"`
	PromptPackVersion  string          `json:"prompt_pack_version"`
	RulePackVersion    string          `json:"rule_pack_version"`
	ModelProfile       string          `json:"model_profile"`
	GuardrailProfile   string          `json:"guardrail_profile"`
	MemoryNamespace    MemoryNamespace `json:"memory_namespace"`
	RequestedBy        string          `json:"requested_by"`
	CreatedBy          string          `json:"created_by"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	StartedAt          *time.Time      `json:"started_at,omitempty"`
	FinishedAt         *time.Time      `json:"finished_at,omitempty"`
	CompletedAt        *time.Time      `json:"completed_at,omitempty"`
	ArtifactBundleRefs []string        `json:"artifact_bundle_refs"`
	MemorySnapshotRefs []string        `json:"memory_snapshot_refs"`
	ReviewMetadataJSON json.RawMessage `json:"review_metadata_json"`
	Notes              string          `json:"notes"`
	MetadataJSON       json.RawMessage `json:"metadata_json"`
}

type CreateInput struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Status             string          `json:"status"`
	ScenarioType       string          `json:"scenario_type"`
	RunType            string          `json:"run_type"`
	ExecutionMode      string          `json:"execution_mode"`
	ExecutionRing      string          `json:"execution_ring"`
	GuardrailStatus    string          `json:"guardrail_status"`
	Repo               string          `json:"repo"`
	Domain             string          `json:"domain"`
	DatasetRefs        []string        `json:"dataset_refs"`
	PromptPackVersion  string          `json:"prompt_pack_version"`
	RulePackVersion    string          `json:"rule_pack_version"`
	ModelProfile       string          `json:"model_profile"`
	GuardrailProfile   string          `json:"guardrail_profile"`
	MemoryNamespace    MemoryNamespace `json:"memory_namespace"`
	RequestedBy        string          `json:"requested_by"`
	CreatedBy          string          `json:"created_by"`
	StartedAt          *time.Time      `json:"started_at,omitempty"`
	FinishedAt         *time.Time      `json:"finished_at,omitempty"`
	ArtifactBundleRefs []string        `json:"artifact_bundle_refs"`
	MemorySnapshotRefs []string        `json:"memory_snapshot_refs"`
	ReviewMetadataJSON json.RawMessage `json:"review_metadata_json"`
	Notes              string          `json:"notes"`
	MetadataJSON       json.RawMessage `json:"metadata_json"`
}

type UpdateInput struct {
	Name               *string          `json:"name,omitempty"`
	Description        *string          `json:"description,omitempty"`
	Status             *string          `json:"status,omitempty"`
	ScenarioType       *string          `json:"scenario_type,omitempty"`
	RunType            *string          `json:"run_type,omitempty"`
	ExecutionMode      *string          `json:"execution_mode,omitempty"`
	ExecutionRing      *string          `json:"execution_ring,omitempty"`
	GuardrailStatus    *string          `json:"guardrail_status,omitempty"`
	Repo               *string          `json:"repo,omitempty"`
	Domain             *string          `json:"domain,omitempty"`
	DatasetRefs        *[]string        `json:"dataset_refs,omitempty"`
	PromptPackVersion  *string          `json:"prompt_pack_version,omitempty"`
	RulePackVersion    *string          `json:"rule_pack_version,omitempty"`
	ModelProfile       *string          `json:"model_profile,omitempty"`
	GuardrailProfile   *string          `json:"guardrail_profile,omitempty"`
	MemoryNamespace    *MemoryNamespace `json:"memory_namespace,omitempty"`
	RequestedBy        *string          `json:"requested_by,omitempty"`
	StartedAt          *time.Time       `json:"started_at,omitempty"`
	FinishedAt         *time.Time       `json:"finished_at,omitempty"`
	CompletedAt        *time.Time       `json:"completed_at,omitempty"`
	ArtifactBundleRefs *[]string        `json:"artifact_bundle_refs,omitempty"`
	MemorySnapshotRefs *[]string        `json:"memory_snapshot_refs,omitempty"`
	ReviewMetadataJSON *json.RawMessage `json:"review_metadata_json,omitempty"`
	Notes              *string          `json:"notes,omitempty"`
	MetadataJSON       *json.RawMessage `json:"metadata_json,omitempty"`
}

type ModelProfile struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Provider           string          `json:"provider"`
	BaseURL            string          `json:"base_url"`
	PrimaryModel       string          `json:"primary_model"`
	GuardrailModel     string          `json:"guardrail_model"`
	HelperModel        string          `json:"helper_model"`
	TimeoutSeconds     int             `json:"timeout_seconds"`
	Temperature        float64         `json:"temperature"`
	MaxTokens          int             `json:"max_tokens"`
	JSONMode           bool            `json:"json_mode"`
	StructuredOutput   bool            `json:"structured_output"`
	EnableGuardrails   bool            `json:"enable_guardrails"`
	EnableHelperModel  bool            `json:"enable_helper_model"`
	ConnectionMetadata json.RawMessage `json:"connection_metadata"`
	CreatedBy          string          `json:"created_by"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type RegisterModelProfileInput struct {
	Name               string          `json:"name"`
	Provider           string          `json:"provider"`
	BaseURL            string          `json:"base_url"`
	PrimaryModel       string          `json:"primary_model"`
	GuardrailModel     string          `json:"guardrail_model"`
	HelperModel        string          `json:"helper_model"`
	TimeoutSeconds     int             `json:"timeout_seconds"`
	Temperature        float64         `json:"temperature"`
	MaxTokens          int             `json:"max_tokens"`
	JSONMode           *bool           `json:"json_mode,omitempty"`
	StructuredOutput   *bool           `json:"structured_output,omitempty"`
	EnableGuardrails   *bool           `json:"enable_guardrails,omitempty"`
	EnableHelperModel  *bool           `json:"enable_helper_model,omitempty"`
	ConnectionMetadata json.RawMessage `json:"connection_metadata"`
}

type Artifact struct {
	ID           string          `json:"id"`
	RunID        string          `json:"run_id"`
	CycleID      *string         `json:"cycle_id,omitempty"`
	ArtifactType string          `json:"artifact_type"`
	URI          string          `json:"uri"`
	ContentType  string          `json:"content_type"`
	Version      string          `json:"version"`
	Checksum     string          `json:"checksum"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AttachArtifactInput struct {
	CycleID      *string         `json:"cycle_id,omitempty"`
	ArtifactType string          `json:"artifact_type"`
	URI          string          `json:"uri"`
	ContentType  string          `json:"content_type"`
	Version      string          `json:"version"`
	Checksum     string          `json:"checksum"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
}

type Cycle struct {
	ID                     string          `json:"id"`
	RunID                  string          `json:"run_id"`
	CycleKey               string          `json:"cycle_key"`
	Focus                  string          `json:"focus"`
	Objective              string          `json:"objective"`
	DetectorPack           string          `json:"detector_pack"`
	ExecutionRing          string          `json:"execution_ring"`
	GuardrailStatus        string          `json:"guardrail_status"`
	SummaryRef             string          `json:"summary_ref"`
	CarryForwardSummaryRef string          `json:"carry_forward_summary_ref"`
	MemorySnapshotRef      string          `json:"memory_snapshot_ref"`
	Status                 string          `json:"status"`
	StartedAt              *time.Time      `json:"started_at,omitempty"`
	FinishedAt             *time.Time      `json:"finished_at,omitempty"`
	MetadataJSON           json.RawMessage `json:"metadata_json"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

type CreateCycleInput struct {
	CycleKey               string          `json:"cycle_key"`
	Focus                  string          `json:"focus"`
	Objective              string          `json:"objective"`
	DetectorPack           string          `json:"detector_pack"`
	ExecutionRing          string          `json:"execution_ring"`
	SummaryRef             string          `json:"summary_ref"`
	CarryForwardSummaryRef string          `json:"carry_forward_summary_ref"`
	Status                 string          `json:"status"`
	MetadataJSON           json.RawMessage `json:"metadata_json"`
}

type UpdateCycleInput struct {
	Focus                  *string          `json:"focus,omitempty"`
	Objective              *string          `json:"objective,omitempty"`
	DetectorPack           *string          `json:"detector_pack,omitempty"`
	ExecutionRing          *string          `json:"execution_ring,omitempty"`
	GuardrailStatus        *string          `json:"guardrail_status,omitempty"`
	SummaryRef             *string          `json:"summary_ref,omitempty"`
	CarryForwardSummaryRef *string          `json:"carry_forward_summary_ref,omitempty"`
	Status                 *string          `json:"status,omitempty"`
	StartedAt              *time.Time       `json:"started_at,omitempty"`
	FinishedAt             *time.Time       `json:"finished_at,omitempty"`
	MetadataJSON           *json.RawMessage `json:"metadata_json,omitempty"`
	MemoryNote             *string          `json:"memory_note,omitempty"`
	AgentNamespace         *string          `json:"agent_namespace,omitempty"`
}

type Comparison struct {
	ID                   string          `json:"id"`
	RunID                string          `json:"run_id"`
	CycleID              *string         `json:"cycle_id,omitempty"`
	DeterministicSummary json.RawMessage `json:"deterministic_summary"`
	LLMSummary           json.RawMessage `json:"llm_summary"`
	GuardrailSummary     json.RawMessage `json:"guardrail_summary"`
	Deltas               json.RawMessage `json:"deltas"`
	ReviewStatus         string          `json:"review_status"`
	ReviewerNotes        string          `json:"reviewer_notes"`
	FinalDisposition     string          `json:"final_disposition"`
	FinalOutput          json.RawMessage `json:"final_output"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type UpsertComparisonInput struct {
	CycleID              *string         `json:"cycle_id,omitempty"`
	DeterministicSummary json.RawMessage `json:"deterministic_summary"`
	LLMSummary           json.RawMessage `json:"llm_summary"`
	GuardrailSummary     json.RawMessage `json:"guardrail_summary"`
	Deltas               json.RawMessage `json:"deltas"`
	ReviewStatus         string          `json:"review_status"`
	ReviewerNotes        string          `json:"reviewer_notes"`
	FinalDisposition     string          `json:"final_disposition"`
	FinalOutput          json.RawMessage `json:"final_output"`
}

type ExecuteRunInput struct {
	CycleID        *string         `json:"cycle_id,omitempty"`
	AgentNamespace *string         `json:"agent_namespace,omitempty"`
	Prompt         string          `json:"prompt"`
	SystemPrompt   string          `json:"system_prompt"`
	InputJSON      json.RawMessage `json:"input_json"`
	MemoryNote     string          `json:"memory_note"`
}

type ExecuteRunResult struct {
	RunID                string          `json:"run_id"`
	RunType              string          `json:"run_type"`
	ExecutionMode        string          `json:"execution_mode"`
	Status               string          `json:"status"`
	CycleID              *string         `json:"cycle_id,omitempty"`
	MemorySnapshotRef    string          `json:"memory_snapshot_ref,omitempty"`
	DeterministicSummary json.RawMessage `json:"deterministic_summary"`
	LLMSummary           json.RawMessage `json:"llm_summary"`
	GuardrailSummary     json.RawMessage `json:"guardrail_summary"`
	GuardrailStatus      string          `json:"guardrail_status"`
	Artifacts            []Artifact      `json:"artifacts"`
	Comparison           *Comparison     `json:"comparison,omitempty"`
}

type PolicyDecisionInput struct {
	ActionType          string          `json:"action_type"`
	TargetRunID         *string         `json:"target_run_id,omitempty"`
	TargetCycleID       *string         `json:"target_cycle_id,omitempty"`
	ActorID             string          `json:"actor_id"`
	ActorType           string          `json:"actor_type"`
	PolicyInput         json.RawMessage `json:"policy_input"`
	Allow               bool            `json:"allow"`
	PolicyBundleID      string          `json:"policy_bundle_id"`
	PolicyBundleVersion string          `json:"policy_bundle_version"`
	ReasonCode          string          `json:"reason_code"`
	ConditionsApplied   []string        `json:"conditions_applied"`
	FallbackMode        string          `json:"fallback_mode"`
}

type PolicyDecision struct {
	ID                  string          `json:"id"`
	ActionType          string          `json:"action_type"`
	TargetRunID         *string         `json:"target_run_id,omitempty"`
	TargetCycleID       *string         `json:"target_cycle_id,omitempty"`
	ActorID             string          `json:"actor_id"`
	ActorType           string          `json:"actor_type"`
	PolicyInput         json.RawMessage `json:"policy_input"`
	Allow               bool            `json:"allow"`
	PolicyBundleID      string          `json:"policy_bundle_id"`
	PolicyBundleVersion string          `json:"policy_bundle_version"`
	ReasonCode          string          `json:"reason_code"`
	ConditionsApplied   []string        `json:"conditions_applied"`
	FallbackMode        string          `json:"fallback_mode"`
	CreatedAt           time.Time       `json:"created_at"`
}

type ReviewActionInput struct {
	Action           string  `json:"action"`
	ReviewerID       string  `json:"reviewer_id"`
	ReviewerType     string  `json:"reviewer_type"`
	Rationale        string  `json:"rationale"`
	CycleID          *string `json:"cycle_id,omitempty"`
	PolicyDecisionID *string `json:"policy_decision_id,omitempty"`
}

type ReviewActionRecord struct {
	ID               string    `json:"id"`
	RunID            string    `json:"run_id"`
	CycleID          *string   `json:"cycle_id,omitempty"`
	ReviewerID       string    `json:"reviewer_id"`
	ReviewerType     string    `json:"reviewer_type"`
	ActionType       string    `json:"action_type"`
	PriorStatus      string    `json:"prior_status"`
	NewStatus        string    `json:"new_status"`
	Rationale        string    `json:"rationale"`
	PolicyDecisionID *string   `json:"policy_decision_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type GovernanceAuditEventInput struct {
	ActorID          string          `json:"actor_id"`
	ActorType        string          `json:"actor_type"`
	ActionType       string          `json:"action_type"`
	TargetRunID      *string         `json:"target_run_id,omitempty"`
	TargetCycleID    *string         `json:"target_cycle_id,omitempty"`
	TargetArtifactID *string         `json:"target_artifact_id,omitempty"`
	PolicyDecisionID *string         `json:"policy_decision_id,omitempty"`
	PayloadSummary   json.RawMessage `json:"payload_summary"`
}

type GovernanceAuditEvent struct {
	ID                string          `json:"id"`
	PreviousEventHash string          `json:"previous_event_hash"`
	CurrentEventHash  string          `json:"current_event_hash"`
	ActorID           string          `json:"actor_id"`
	ActorType         string          `json:"actor_type"`
	ActionType        string          `json:"action_type"`
	TargetRunID       *string         `json:"target_run_id,omitempty"`
	TargetCycleID     *string         `json:"target_cycle_id,omitempty"`
	TargetArtifactID  *string         `json:"target_artifact_id,omitempty"`
	PolicyDecisionID  *string         `json:"policy_decision_id,omitempty"`
	PayloadSummary    json.RawMessage `json:"payload_summary"`
	CreatedAt         time.Time       `json:"created_at"`
}

type DependencyStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
	Error    string `json:"error,omitempty"`
}

type DependencyHealth struct {
	Status       string             `json:"status"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

type Repository interface {
	List(context.Context, store.DBTX) ([]Run, error)
	Get(context.Context, store.DBTX, string) (Run, error)
	Create(context.Context, store.DBTX, CreateInput) (Run, error)
	Update(context.Context, store.DBTX, Run) (Run, error)

	CreateCycle(context.Context, store.DBTX, string, CreateCycleInput) (Cycle, error)
	GetCycle(context.Context, store.DBTX, string, string) (Cycle, error)
	UpdateCycle(context.Context, store.DBTX, Cycle) (Cycle, error)

	CreateArtifact(context.Context, store.DBTX, string, AttachArtifactInput) (Artifact, error)
	ListArtifacts(context.Context, store.DBTX, string) ([]Artifact, error)

	UpsertComparison(context.Context, store.DBTX, string, UpsertComparisonInput) (Comparison, error)
	GetComparison(context.Context, store.DBTX, string) (Comparison, error)

	RegisterModelProfile(context.Context, store.DBTX, RegisterModelProfileInput, string) (ModelProfile, error)
	GetModelProfile(context.Context, store.DBTX, string) (ModelProfile, error)

	RecordPolicyDecision(context.Context, store.DBTX, PolicyDecisionInput) (PolicyDecision, error)
	RecordReviewAction(context.Context, store.DBTX, string, ReviewActionInput, string, string) (ReviewActionRecord, error)
	AppendGovernanceAuditEvent(context.Context, store.DBTX, GovernanceAuditEventInput) (GovernanceAuditEvent, error)
}

type Service interface {
	List(context.Context) ([]Run, error)
	Get(context.Context, string) (Run, error)
	Create(context.Context, CreateInput, string) (Run, error)
	Update(context.Context, string, UpdateInput, string) (Run, error)

	CreateCycle(context.Context, string, CreateCycleInput, string) (Cycle, error)
	GetCycle(context.Context, string, string) (Cycle, error)
	UpdateCycle(context.Context, string, string, UpdateCycleInput, string) (Cycle, error)

	AttachArtifact(context.Context, string, AttachArtifactInput, string) (Artifact, error)
	ListArtifacts(context.Context, string) ([]Artifact, error)

	UpsertComparison(context.Context, string, UpsertComparisonInput, string) (Comparison, error)
	GetComparison(context.Context, string) (Comparison, error)

	RegisterModelProfile(context.Context, RegisterModelProfileInput, string) (ModelProfile, error)
	GetModelProfile(context.Context, string) (ModelProfile, error)
	StartRun(context.Context, string, ExecuteRunInput, string) (ExecuteRunResult, error)
	ExecuteCycleRun(context.Context, string, string, ExecuteRunInput, string) (ExecuteRunResult, error)
	ReviewAction(context.Context, string, ReviewActionInput, string) (Run, error)

	DependencyHealth(context.Context) (DependencyHealth, error)
}
