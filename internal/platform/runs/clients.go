package runs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MemoryContext struct {
	PriorCycleSummaries []string `json:"prior_cycle_summaries"`
	CarryForwardRisks   []string `json:"carry_forward_risks"`
	UnresolvedGaps      []string `json:"unresolved_gaps"`
	BacklogItems        []string `json:"backlog_items"`
	ReviewerNotes       []string `json:"reviewer_notes"`
}

type MemoryWriteInput struct {
	Note                string   `json:"note"`
	PriorCycleSummaries []string `json:"prior_cycle_summaries"`
	CarryForwardRisks   []string `json:"carry_forward_risks"`
	UnresolvedGaps      []string `json:"unresolved_gaps"`
	BacklogItems        []string `json:"backlog_items"`
	ReviewerNotes       []string `json:"reviewer_notes"`
}

type MemoryClient interface {
	BaseURL() string
	FetchScopedContext(context.Context, MemoryNamespace) (MemoryContext, error)
	PersistScopedNotes(context.Context, MemoryNamespace, MemoryWriteInput) (string, error)
}

type NoopMemoryClient struct{}

func NewNoopMemoryClient() *NoopMemoryClient {
	return &NoopMemoryClient{}
}

func (c *NoopMemoryClient) BaseURL() string {
	return ""
}

func (c *NoopMemoryClient) FetchScopedContext(context.Context, MemoryNamespace) (MemoryContext, error) {
	return MemoryContext{}, nil
}

func (c *NoopMemoryClient) PersistScopedNotes(context.Context, MemoryNamespace, MemoryWriteInput) (string, error) {
	return "", nil
}

type HTTPMemoryClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPMemoryClient(baseURL string, timeout time.Duration) *HTTPMemoryClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPMemoryClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *HTTPMemoryClient) BaseURL() string {
	return c.baseURL
}

func (c *HTTPMemoryClient) FetchScopedContext(ctx context.Context, ns MemoryNamespace) (MemoryContext, error) {
	if c.baseURL == "" {
		return MemoryContext{}, nil
	}

	payload := map[string]any{"namespace": ns}
	body, err := json.Marshal(payload)
	if err != nil {
		return MemoryContext{}, fmt.Errorf("marshal memory context request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/scoped-memory/context", bytes.NewReader(body))
	if err != nil {
		return MemoryContext{}, fmt.Errorf("build memory context request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MemoryContext{}, fmt.Errorf("fetch scoped memory context: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return MemoryContext{}, fmt.Errorf("fetch scoped memory context returned status %d", resp.StatusCode)
	}

	var decoded struct {
		Data MemoryContext `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return MemoryContext{}, fmt.Errorf("decode scoped memory context response: %w", err)
	}

	return decoded.Data, nil
}

func (c *HTTPMemoryClient) PersistScopedNotes(ctx context.Context, ns MemoryNamespace, input MemoryWriteInput) (string, error) {
	if c.baseURL == "" {
		return "", nil
	}

	payload := map[string]any{"namespace": ns, "input": input}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal memory note request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/scoped-memory/notes", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build memory note request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("persist scoped memory note: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("persist scoped memory note returned status %d", resp.StatusCode)
	}

	var decoded struct {
		Data struct {
			SnapshotRef string `json:"snapshot_ref"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode scoped memory note response: %w", err)
	}

	return decoded.Data.SnapshotRef, nil
}

type InferenceRequest struct {
	Provider                string          `json:"provider"`
	BaseURL                 string          `json:"base_url,omitempty"`
	Prompt                  string          `json:"prompt"`
	SystemPrompt            string          `json:"system_prompt"`
	InputJSON               json.RawMessage `json:"input_json"`
	PrimaryModel            string          `json:"primary_model"`
	GuardrailModel          string          `json:"guardrail_model"`
	HelperModel             string          `json:"helper_model"`
	Temperature             float64         `json:"temperature"`
	MaxTokens               int             `json:"max_tokens"`
	ExpectJSON              bool            `json:"expect_json"`
	EnableGuardrails        bool            `json:"enable_guardrails"`
	EnableHelperModel       bool            `json:"enable_helper_model"`
	HelperRequested         bool            `json:"helper_requested"`
	PrimaryTimeoutSeconds   int             `json:"primary_timeout_seconds,omitempty"`
	GuardrailTimeoutSeconds int             `json:"guardrail_timeout_seconds,omitempty"`
	HelperTimeoutSeconds    int             `json:"helper_timeout_seconds,omitempty"`
	ConnectionMeta          json.RawMessage `json:"connection_metadata"`
}

type InferenceResponse struct {
	PrimaryOutput   json.RawMessage `json:"primary_output"`
	GuardrailOutput json.RawMessage `json:"guardrail_output"`
	Guardrail       json.RawMessage `json:"guardrail"`
	GuardrailStatus string          `json:"guardrail_status"`
	GuardrailScore  *float64        `json:"guardrail_score,omitempty"`
	HelperOutput    json.RawMessage `json:"helper_output"`
	PrimaryText     string          `json:"primary_text,omitempty"`
	GuardrailText   string          `json:"guardrail_text,omitempty"`
	HelperText      string          `json:"helper_text,omitempty"`
	LatencyMS       int64           `json:"latency_ms"`
	ModelTag        string          `json:"model_tag"`
	GuardrailTag    string          `json:"guardrail_tag"`
	HelperTag       string          `json:"helper_tag"`
}

type InferenceClient interface {
	BaseURL() string
	Execute(context.Context, InferenceRequest) (InferenceResponse, error)
}

type NoopInferenceClient struct{}

func NewNoopInferenceClient() *NoopInferenceClient {
	return &NoopInferenceClient{}
}

func (c *NoopInferenceClient) BaseURL() string {
	return ""
}

func (c *NoopInferenceClient) Execute(context.Context, InferenceRequest) (InferenceResponse, error) {
	return InferenceResponse{}, fmt.Errorf("inference client is not configured")
}

type HTTPInferenceClient struct {
	baseURL        string
	defaultTimeout time.Duration
	httpClient     *http.Client
}

func NewHTTPInferenceClient(baseURL string, timeout time.Duration) *HTTPInferenceClient {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return &HTTPInferenceClient{
		baseURL:        strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		defaultTimeout: timeout,
		// Context deadlines are authoritative for inference calls.
		httpClient: &http.Client{},
	}
}

func (c *HTTPInferenceClient) BaseURL() string {
	return c.baseURL
}

func (c *HTTPInferenceClient) Execute(ctx context.Context, input InferenceRequest) (InferenceResponse, error) {
	baseURL := c.resolveBaseURL(input.BaseURL)
	if baseURL == "" {
		return InferenceResponse{}, fmt.Errorf("inference base url is not configured")
	}

	switch strings.ToLower(strings.TrimSpace(input.Provider)) {
	case "local_ollama", "ollama":
		return c.executeLocalOllama(ctx, baseURL, input)
	default:
		return c.executeGateway(ctx, baseURL, input)
	}
}

func (c *HTTPInferenceClient) executeGateway(ctx context.Context, baseURL string, input InferenceRequest) (InferenceResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("marshal inference request: %w", err)
	}
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/inference/execute", bytes.NewReader(body))
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("build inference request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("execute inference request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return InferenceResponse{}, fmt.Errorf("execute inference request returned status %d", resp.StatusCode)
	}

	var decoded struct {
		Data InferenceResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return InferenceResponse{}, fmt.Errorf("decode inference response: %w", err)
	}

	if input.ExpectJSON && len(decoded.Data.PrimaryOutput) == 0 {
		return InferenceResponse{}, fmt.Errorf("structured output requested but primary_output is empty")
	}
	if len(decoded.Data.GuardrailOutput) == 0 && len(decoded.Data.Guardrail) != 0 {
		decoded.Data.GuardrailOutput = decoded.Data.Guardrail
	}
	if len(decoded.Data.Guardrail) == 0 && len(decoded.Data.GuardrailOutput) != 0 {
		decoded.Data.Guardrail = decoded.Data.GuardrailOutput
	}
	if strings.TrimSpace(decoded.Data.GuardrailStatus) == "" {
		if input.EnableGuardrails {
			decoded.Data.GuardrailStatus = string(GuardrailStatusPassed)
		} else {
			decoded.Data.GuardrailStatus = string(GuardrailStatusDisabled)
		}
	}

	slog.Default().Debug("inference.gateway.completed",
		"provider", strings.TrimSpace(input.Provider),
		"model", input.PrimaryModel,
		"payload_bytes", len(body),
		"elapsed_ms", time.Since(start).Milliseconds(),
		"guardrail_invoked", input.EnableGuardrails,
		"helper_invoked", input.EnableHelperModel && input.HelperRequested,
		"parse_ok", true,
	)

	return decoded.Data, nil
}

func (c *HTTPInferenceClient) executeLocalOllama(ctx context.Context, baseURL string, input InferenceRequest) (InferenceResponse, error) {
	start := time.Now()

	if strings.TrimSpace(input.PrimaryModel) == "" {
		return InferenceResponse{}, fmt.Errorf("primary_model is required for local_ollama")
	}

	primaryTimeout := c.resolveTimeout(input.PrimaryTimeoutSeconds)
	guardrailTimeout := c.resolveTimeout(input.GuardrailTimeoutSeconds)
	helperTimeout := c.resolveTimeout(input.HelperTimeoutSeconds)
	guardrailInvoked := input.EnableGuardrails && strings.TrimSpace(input.GuardrailModel) != ""
	helperInvoked := input.EnableHelperModel && input.HelperRequested && strings.TrimSpace(input.HelperModel) != ""
	phaseRootCtx, phaseRootCancel, phaseRootDetached := buildOllamaPhaseRootContext(ctx, primaryTimeout, guardrailTimeout, helperTimeout, guardrailInvoked, helperInvoked)
	defer phaseRootCancel()
	if phaseRootDetached {
		slog.Default().Debug("inference.local_ollama.phase_root_detached",
			"provider", "local_ollama",
			"guardrail_invoked", guardrailInvoked,
			"helper_invoked", helperInvoked,
			"primary_timeout_ms", primaryTimeout.Milliseconds(),
			"guardrail_timeout_ms", guardrailTimeout.Milliseconds(),
			"helper_timeout_ms", helperTimeout.Milliseconds(),
		)
	}

	primaryContent, primaryPayloadBytes, primaryElapsed, err := c.callOllamaChat(phaseRootCtx, baseURL, "primary", primaryTimeout, ollamaChatRequest{
		Model:    input.PrimaryModel,
		Messages: buildPrimaryMessages(input),
		Format:   localOllamaFormat(input.ExpectJSON),
		Stream:   false,
		Options:  buildOllamaOptions(input),
	})
	if err != nil {
		return InferenceResponse{}, err
	}

	primaryOutput, primaryText, err := mapOllamaContent(primaryContent, input.ExpectJSON)
	if err != nil {
		slog.Default().Debug("inference.local_ollama.parse_failed",
			"provider", "local_ollama",
			"model", input.PrimaryModel,
			"phase", "primary",
			"payload_bytes", primaryPayloadBytes,
			"elapsed_ms", primaryElapsed.Milliseconds(),
			"parse_ok", false,
		)
		return InferenceResponse{}, fmt.Errorf("map primary ollama response: %w", err)
	}
	slog.Default().Debug("inference.local_ollama.completed",
		"provider", "local_ollama",
		"model", input.PrimaryModel,
		"phase", "primary",
		"payload_bytes", primaryPayloadBytes,
		"elapsed_ms", primaryElapsed.Milliseconds(),
		"guardrail_invoked", guardrailInvoked,
		"helper_invoked", helperInvoked,
		"parse_ok", true,
	)

	response := InferenceResponse{
		PrimaryOutput:   primaryOutput,
		PrimaryText:     primaryText,
		ModelTag:        input.PrimaryModel,
		GuardrailStatus: string(GuardrailStatusDisabled),
	}

	if guardrailInvoked {
		thinkFalse := false
		guardrailContent, guardrailPayloadBytes, guardrailElapsed, guardrailErr := c.callOllamaChat(phaseRootCtx, baseURL, "guardrail", guardrailTimeout, ollamaChatRequest{
			Model: input.GuardrailModel,
			Messages: []ollamaMessage{
				{Role: "system", Content: "You are a strict guardrail scorer. Return only <score> yes </score> or <score> no </score>."},
				{Role: "user", Content: buildCompactGuardrailInput(input, primaryOutput)},
			},
			Stream:    false,
			Think:     &thinkFalse,
			KeepAlive: "10m",
			Options:   buildOllamaOptions(input),
		})
		if guardrailErr != nil {
			response.GuardrailStatus = classifyGuardrailFailure(guardrailErr)
			response.GuardrailText = guardrailErr.Error()
			slog.Default().Debug("inference.local_ollama.guardrail_failed",
				"provider", "local_ollama",
				"model", input.GuardrailModel,
				"phase", "guardrail",
				"payload_bytes", guardrailPayloadBytes,
				"elapsed_ms", guardrailElapsed.Milliseconds(),
				"guardrail_status", response.GuardrailStatus,
				"think_false", true,
				"parse_ok", false,
			)
		} else {
			guardrailOutput, guardrailText, guardrailStatus, guardrailScore, guardrailMapErr := parseLocalGuardrailContent(guardrailContent)
			if guardrailMapErr != nil {
				response.GuardrailStatus = string(GuardrailStatusUnavailable)
				response.GuardrailText = fmt.Sprintf("guardrail response mapping failed: %v", guardrailMapErr)
				slog.Default().Debug("inference.local_ollama.parse_failed",
					"provider", "local_ollama",
					"model", input.GuardrailModel,
					"phase", "guardrail",
					"payload_bytes", guardrailPayloadBytes,
					"elapsed_ms", guardrailElapsed.Milliseconds(),
					"guardrail_status", response.GuardrailStatus,
					"think_false", true,
					"parse_ok", false,
				)
			} else {
				response.GuardrailStatus = guardrailStatus
				response.GuardrailScore = guardrailScore
				response.GuardrailOutput = guardrailOutput
				response.Guardrail = guardrailOutput
				response.GuardrailText = guardrailText
				response.GuardrailTag = input.GuardrailModel
				slog.Default().Debug("inference.local_ollama.completed",
					"provider", "local_ollama",
					"model", input.GuardrailModel,
					"phase", "guardrail",
					"payload_bytes", guardrailPayloadBytes,
					"elapsed_ms", guardrailElapsed.Milliseconds(),
					"guardrail_status", guardrailStatus,
					"guardrail_score", guardrailScoreOrZero(guardrailScore),
					"think_false", true,
					"guardrail_invoked", true,
					"helper_invoked", helperInvoked,
					"parse_ok", true,
				)
			}
		}
	}

	if helperInvoked {
		helperPayload := map[string]any{
			"request_summary": truncateText(input.Prompt, 260),
			"context_summary": compactJSONSummary(input.InputJSON, 8, 260),
			"primary_summary": compactJSONSummary(primaryOutput, 8, 260),
		}
		helperBody, _ := json.Marshal(helperPayload)

		helperContent, helperPayloadBytes, helperElapsed, helperErr := c.callOllamaChat(phaseRootCtx, baseURL, "helper", helperTimeout, ollamaChatRequest{
			Model: input.HelperModel,
			Messages: []ollamaMessage{
				{Role: "system", Content: "Provide helper transformation output. Return JSON when possible."},
				{Role: "user", Content: string(helperBody)},
			},
			Format:  localOllamaFormat(input.ExpectJSON),
			Stream:  false,
			Options: buildOllamaOptions(input),
		})
		if helperErr != nil {
			slog.Default().Debug("inference.local_ollama.failed",
				"provider", "local_ollama",
				"model", input.HelperModel,
				"phase", "helper",
				"payload_bytes", helperPayloadBytes,
				"elapsed_ms", helperElapsed.Milliseconds(),
				"parse_ok", false,
			)
			response.HelperText = fmt.Sprintf("helper model call failed: %v", helperErr)
		} else {
			helperOutput, helperText, helperMapErr := mapOllamaContent(helperContent, input.ExpectJSON)
			if helperMapErr != nil {
				slog.Default().Debug("inference.local_ollama.parse_failed",
					"provider", "local_ollama",
					"model", input.HelperModel,
					"phase", "helper",
					"payload_bytes", helperPayloadBytes,
					"elapsed_ms", helperElapsed.Milliseconds(),
					"parse_ok", false,
				)
				response.HelperText = fmt.Sprintf("helper model response mapping failed: %v", helperMapErr)
			} else {
				slog.Default().Debug("inference.local_ollama.completed",
					"provider", "local_ollama",
					"model", input.HelperModel,
					"phase", "helper",
					"payload_bytes", helperPayloadBytes,
					"elapsed_ms", helperElapsed.Milliseconds(),
					"guardrail_invoked", guardrailInvoked,
					"helper_invoked", true,
					"parse_ok", true,
				)
				response.HelperOutput = helperOutput
				response.HelperText = helperText
				response.HelperTag = input.HelperModel
			}
		}
	}

	response.LatencyMS = time.Since(start).Milliseconds()
	return response, nil
}

func buildOllamaPhaseRootContext(
	parent context.Context,
	primaryTimeout time.Duration,
	guardrailTimeout time.Duration,
	helperTimeout time.Duration,
	guardrailInvoked bool,
	helperInvoked bool,
) (context.Context, context.CancelFunc, bool) {
	required := primaryTimeout
	if required <= 0 {
		required = 15 * time.Second
	}
	if guardrailInvoked {
		required += guardrailTimeout
	}
	if helperInvoked {
		required += helperTimeout
	}
	// Small buffer for parse/persistence work around model calls.
	required += 3 * time.Second

	noop := func() {}
	if required <= 0 {
		return parent, noop, false
	}
	if err := parent.Err(); err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			return parent, noop, false
		}
		detached, cancel := context.WithTimeout(context.WithoutCancel(parent), required)
		return detached, cancel, true
	}
	if deadline, ok := parent.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 && remaining >= required {
			return parent, noop, false
		}
		detached, cancel := context.WithTimeout(context.WithoutCancel(parent), required)
		return detached, cancel, true
	}
	return parent, noop, false
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model     string          `json:"model"`
	Messages  []ollamaMessage `json:"messages"`
	Format    any             `json:"format,omitempty"`
	Stream    bool            `json:"stream"`
	Think     *bool           `json:"think,omitempty"`
	KeepAlive string          `json:"keep_alive,omitempty"`
	Options   map[string]any  `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Error string `json:"error"`
}

func (c *HTTPInferenceClient) callOllamaChat(
	ctx context.Context,
	baseURL string,
	phase string,
	timeout time.Duration,
	payload ollamaChatRequest,
) (string, int, time.Duration, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", 0, 0, fmt.Errorf("marshal ollama chat request: %w", err)
	}
	payloadBytes := len(body)

	callCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		callCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	start := time.Now()
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", payloadBytes, time.Since(start), fmt.Errorf("build ollama chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", payloadBytes, time.Since(start), fmt.Errorf("execute ollama %s request: %w", phase, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", payloadBytes, time.Since(start), fmt.Errorf("read ollama response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		return "", payloadBytes, time.Since(start), fmt.Errorf("ollama %s request returned status %d", phase, resp.StatusCode)
	}

	var decoded ollamaChatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return "", payloadBytes, time.Since(start), fmt.Errorf("decode ollama chat response: %w", err)
	}

	if strings.TrimSpace(decoded.Error) != "" {
		return "", payloadBytes, time.Since(start), fmt.Errorf("ollama returned error: %s", decoded.Error)
	}
	if strings.TrimSpace(decoded.Message.Content) == "" {
		return "", payloadBytes, time.Since(start), fmt.Errorf("ollama returned empty message content")
	}

	return decoded.Message.Content, payloadBytes, time.Since(start), nil
}

func buildPrimaryMessages(input InferenceRequest) []ollamaMessage {
	messages := make([]ollamaMessage, 0, 2)
	if strings.TrimSpace(input.SystemPrompt) != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: strings.TrimSpace(input.SystemPrompt)})
	}

	user := strings.TrimSpace(input.Prompt)
	if len(input.InputJSON) != 0 && string(input.InputJSON) != "{}" {
		if user != "" {
			user += "\n\n"
		}
		user += "Context JSON:\n" + string(input.InputJSON)
	}
	if user == "" {
		user = "Provide a JSON analysis response."
	}

	messages = append(messages, ollamaMessage{Role: "user", Content: user})
	return messages
}

func buildOllamaOptions(input InferenceRequest) map[string]any {
	options := map[string]any{}
	if input.MaxTokens > 0 {
		options["num_predict"] = input.MaxTokens
	}
	if input.Temperature >= 0 {
		options["temperature"] = input.Temperature
	}
	if len(options) == 0 {
		return nil
	}
	return options
}

func localOllamaFormat(expectJSON bool) any {
	if expectJSON {
		return "json"
	}
	return nil
}

func mapOllamaContent(content string, expectJSON bool) (json.RawMessage, string, error) {
	trimmed := strings.TrimSpace(stripThinkWrappers(content))
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty content")
	}

	candidate := stripJSONFence(trimmed)
	if expectJSON {
		if json.Valid([]byte(candidate)) {
			return json.RawMessage(candidate), candidate, nil
		}
		if score, ok := parseScoreFromContent(candidate); ok {
			body, _ := json.Marshal(map[string]any{"score": score})
			return body, candidate, nil
		}
		return nil, candidate, fmt.Errorf("content is not valid json")
	}

	if json.Valid([]byte(candidate)) {
		return json.RawMessage(candidate), candidate, nil
	}

	body, _ := json.Marshal(map[string]string{"text": trimmed})
	return body, trimmed, nil
}

func stripJSONFence(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func stripThinkWrappers(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.ReplaceAll(trimmed, "<think></think>", "")
	trimmed = strings.ReplaceAll(trimmed, "<think>\n</think>", "")
	trimmed = strings.ReplaceAll(trimmed, "</think>", "")
	trimmed = strings.ReplaceAll(trimmed, "<think>", "")
	return strings.TrimSpace(trimmed)
}

func parseScoreFromContent(value string) (float64, bool) {
	token, ok := parseScoreTokenContent(value)
	if !ok {
		return 0, false
	}
	score, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return 0, false
	}
	return score, true
}

func parseScoreTokenContent(value string) (string, bool) {
	lower := strings.ToLower(value)
	start := strings.Index(lower, "<score>")
	end := strings.Index(lower, "</score>")
	if start == -1 || end == -1 || end <= start+7 {
		return "", false
	}
	token := strings.TrimSpace(value[start+7 : end])
	if token == "" {
		return "", false
	}
	return strings.ToLower(token), true
}

func classifyGuardrailFailure(err error) string {
	if err == nil {
		return string(GuardrailStatusUnavailable)
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
		return string(GuardrailStatusTimeout)
	}
	return string(GuardrailStatusUnavailable)
}

func parseGuardrailStatusAndScore(output json.RawMessage) (string, *float64) {
	if len(output) == 0 || string(output) == "{}" {
		return string(GuardrailStatusUnavailable), nil
	}
	var payload map[string]any
	if err := json.Unmarshal(output, &payload); err != nil {
		return string(GuardrailStatusFlagged), nil
	}

	if scoreValue, ok := payload["score"]; ok {
		switch typed := scoreValue.(type) {
		case float64:
			if typed >= 0.7 {
				return string(GuardrailStatusFlagged), &typed
			}
			return string(GuardrailStatusPassed), &typed
		case string:
			normalized := strings.ToLower(strings.TrimSpace(typed))
			if isGuardrailYesToken(normalized) {
				score := 1.0
				return string(GuardrailStatusFlagged), &score
			}
			if isGuardrailNoToken(normalized) {
				score := 0.0
				return string(GuardrailStatusPassed), &score
			}
			score, err := strconv.ParseFloat(normalized, 64)
			if err == nil {
				if score >= 0.7 {
					return string(GuardrailStatusFlagged), &score
				}
				return string(GuardrailStatusPassed), &score
			}
		}
	}

	decision := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", payload["decision"])))
	switch decision {
	case "allow", "pass", "passed", "ok", "approve", "approved":
		return string(GuardrailStatusPassed), nil
	case "flag", "flagged", "deny", "reject", "blocked":
		return string(GuardrailStatusFlagged), nil
	default:
		return string(GuardrailStatusFlagged), nil
	}
}

func guardrailScoreOrZero(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func parseLocalGuardrailContent(content string) (json.RawMessage, string, string, *float64, error) {
	trimmed := strings.TrimSpace(stripThinkWrappers(content))
	if trimmed == "" {
		return nil, "", "", nil, fmt.Errorf("empty guardrail content")
	}

	candidate := stripJSONFence(trimmed)
	if json.Valid([]byte(candidate)) {
		output := json.RawMessage(candidate)
		status, score := parseGuardrailStatusAndScore(output)
		return output, candidate, status, score, nil
	}

	if token, ok := parseScoreTokenContent(candidate); ok {
		return mapGuardrailScoreToken(candidate, token)
	}

	plainToken := strings.ToLower(strings.TrimSpace(candidate))
	if isGuardrailYesToken(plainToken) || isGuardrailNoToken(plainToken) {
		return mapGuardrailScoreToken(candidate, plainToken)
	}

	if numeric, err := strconv.ParseFloat(strings.TrimSpace(candidate), 64); err == nil {
		return mapGuardrailNumericScore(candidate, numeric)
	}

	return nil, candidate, "", nil, fmt.Errorf("guardrail content is not parseable score output")
}

func mapGuardrailScoreToken(rawText string, token string) (json.RawMessage, string, string, *float64, error) {
	normalized := strings.ToLower(strings.TrimSpace(token))
	if numeric, err := strconv.ParseFloat(normalized, 64); err == nil {
		return mapGuardrailNumericScore(rawText, numeric)
	}
	switch {
	case isGuardrailYesToken(normalized):
		score := 1.0
		body, _ := json.Marshal(map[string]any{
			"score":  "yes",
			"status": string(GuardrailStatusFlagged),
		})
		return body, strings.TrimSpace(rawText), string(GuardrailStatusFlagged), &score, nil
	case isGuardrailNoToken(normalized):
		score := 0.0
		body, _ := json.Marshal(map[string]any{
			"score":  "no",
			"status": string(GuardrailStatusPassed),
		})
		return body, strings.TrimSpace(rawText), string(GuardrailStatusPassed), &score, nil
	default:
		return nil, strings.TrimSpace(rawText), "", nil, fmt.Errorf("unsupported guardrail score token %q", token)
	}
}

func mapGuardrailNumericScore(rawText string, score float64) (json.RawMessage, string, string, *float64, error) {
	status := string(GuardrailStatusPassed)
	if score >= 0.7 {
		status = string(GuardrailStatusFlagged)
	}
	body, _ := json.Marshal(map[string]any{
		"score":  score,
		"status": status,
	})
	return body, strings.TrimSpace(rawText), status, &score, nil
}

func isGuardrailYesToken(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yes", "y", "true", "flag", "flagged", "deny", "blocked", "reject", "fail":
		return true
	default:
		return false
	}
}

func isGuardrailNoToken(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "no", "n", "false", "allow", "pass", "passed", "ok", "approve", "approved":
		return true
	default:
		return false
	}
}

func decodeRawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return string(raw)
	}
	return out
}

func (c *HTTPInferenceClient) resolveTimeout(seconds int) time.Duration {
	if seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if c.defaultTimeout > 0 {
		return c.defaultTimeout
	}
	return 45 * time.Second
}

func buildCompactGuardrailInput(input InferenceRequest, primaryOutput json.RawMessage) string {
	lines := []string{
		"criterion: flag output when ACH/NACHA or fraud-risk concerns need reviewer escalation.",
		"goal: decide if reviewer follow-up is required.",
	}

	if request := strings.TrimSpace(truncateText(input.Prompt, 220)); request != "" {
		lines = append(lines, "request_summary: "+request)
	}
	if primary := compactJSONString(primaryOutput, 6, 220); primary != "" {
		lines = append(lines, "primary_summary: "+primary)
	}
	if context := compactJSONString(input.InputJSON, 5, 180); context != "" {
		lines = append(lines, "context_summary: "+context)
	}
	lines = append(lines, "question: should this output be flagged? Return only <score> yes </score> or <score> no </score>.")

	return strings.Join(lines, "\n")
}

func compactJSONString(raw json.RawMessage, maxKeys int, maxChars int) string {
	summary := compactJSONSummary(raw, maxKeys, maxChars)
	if len(summary) == 0 {
		return ""
	}
	body, _ := json.Marshal(summary)
	return truncateText(string(body), maxChars)
}

func compactJSONSummary(raw json.RawMessage, maxKeys int, maxChars int) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 || string(raw) == "{}" {
		return out
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		out["raw_preview"] = truncateText(string(raw), maxChars)
		return out
	}

	switch typed := decoded.(type) {
	case map[string]any:
		count := 0
		for key, value := range typed {
			if maxKeys > 0 && count >= maxKeys {
				break
			}
			out[key] = truncateAny(value, maxChars)
			count++
		}
		out["_key_count"] = len(typed)
	case []any:
		out["_array_len"] = len(typed)
		if len(typed) > 0 {
			out["_sample"] = truncateAny(typed[0], maxChars)
		}
	default:
		out["value"] = truncateAny(typed, maxChars)
	}

	return out
}

func truncateAny(value any, maxChars int) any {
	switch typed := value.(type) {
	case string:
		return truncateText(typed, maxChars)
	case map[string]any, []any:
		body, err := json.Marshal(typed)
		if err != nil {
			return truncateText(fmt.Sprintf("%v", value), maxChars)
		}
		return truncateText(string(body), maxChars)
	default:
		return value
	}
}

func truncateText(value string, maxChars int) string {
	value = strings.TrimSpace(value)
	if maxChars <= 0 || len(value) <= maxChars {
		return value
	}
	return value[:maxChars] + "..."
}

func (c *HTTPInferenceClient) resolveBaseURL(override string) string {
	override = strings.TrimRight(strings.TrimSpace(override), "/")
	if override != "" {
		return override
	}
	return c.baseURL
}
