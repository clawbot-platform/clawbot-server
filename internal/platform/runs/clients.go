package runs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

	primaryContent, primaryPayloadBytes, primaryElapsed, err := c.callOllamaChat(ctx, baseURL, "primary", primaryTimeout, ollamaChatRequest{
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
		PrimaryOutput: primaryOutput,
		PrimaryText:   primaryText,
		ModelTag:      input.PrimaryModel,
	}

	if guardrailInvoked {
		guardrailContent, guardrailPayloadBytes, guardrailElapsed, guardrailErr := c.callOllamaChat(ctx, baseURL, "guardrail", guardrailTimeout, ollamaChatRequest{
			Model: input.GuardrailModel,
			Messages: []ollamaMessage{
				{Role: "system", Content: "Evaluate policy risk. Return compact JSON with fields decision, risk_level, rationale."},
				{Role: "user", Content: buildCompactGuardrailInput(input, primaryOutput)},
			},
			Format:  localOllamaFormat(true),
			Stream:  false,
			Options: buildOllamaOptions(input),
		})
		if guardrailErr != nil {
			return InferenceResponse{}, guardrailErr
		}

		guardrailOutput, guardrailText, guardrailMapErr := mapOllamaContent(guardrailContent, true)
		if guardrailMapErr != nil {
			slog.Default().Debug("inference.local_ollama.parse_failed",
				"provider", "local_ollama",
				"model", input.GuardrailModel,
				"phase", "guardrail",
				"payload_bytes", guardrailPayloadBytes,
				"elapsed_ms", guardrailElapsed.Milliseconds(),
				"parse_ok", false,
			)
			return InferenceResponse{}, fmt.Errorf("map guardrail ollama response: %w", guardrailMapErr)
		}
		slog.Default().Debug("inference.local_ollama.completed",
			"provider", "local_ollama",
			"model", input.GuardrailModel,
			"phase", "guardrail",
			"payload_bytes", guardrailPayloadBytes,
			"elapsed_ms", guardrailElapsed.Milliseconds(),
			"guardrail_invoked", true,
			"helper_invoked", helperInvoked,
			"parse_ok", true,
		)

		response.GuardrailOutput = guardrailOutput
		response.Guardrail = guardrailOutput
		response.GuardrailText = guardrailText
		response.GuardrailTag = input.GuardrailModel
	}

	if helperInvoked {
		helperPayload := map[string]any{
			"request_summary": truncateText(input.Prompt, 260),
			"context_summary": compactJSONSummary(input.InputJSON, 8, 260),
			"primary_summary": compactJSONSummary(primaryOutput, 8, 260),
		}
		helperBody, _ := json.Marshal(helperPayload)

		helperContent, helperPayloadBytes, helperElapsed, helperErr := c.callOllamaChat(ctx, baseURL, "helper", helperTimeout, ollamaChatRequest{
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

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Format   any             `json:"format,omitempty"`
	Stream   bool            `json:"stream"`
	Options  map[string]any  `json:"options,omitempty"`
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
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty content")
	}

	candidate := stripJSONFence(trimmed)
	if expectJSON {
		if json.Valid([]byte(candidate)) {
			return json.RawMessage(candidate), candidate, nil
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
	payload := map[string]any{
		"request_summary": truncateText(input.Prompt, 320),
		"system_summary":  truncateText(input.SystemPrompt, 260),
		"context_summary": compactJSONSummary(input.InputJSON, 8, 220),
		"primary_summary": compactJSONSummary(primaryOutput, 10, 220),
		"risk_criteria": []string{
			"nacha_compliance",
			"fraud_signal_quality",
			"policy_violation_risk",
			"hallucination_risk",
		},
	}
	body, _ := json.Marshal(payload)
	return string(body)
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
