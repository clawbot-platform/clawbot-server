package runs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPInferenceClientExecuteLocalOllamaPrimary(t *testing.T) {
	t.Helper()

	var (
		chatCalls int
		lastBody  map[string]any
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("expected /api/chat path, got %s", r.URL.Path)
		}
		chatCalls++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer r.Body.Close()
		if err := json.Unmarshal(body, &lastBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"ibm/granite3.3:8b","message":{"role":"assistant","content":"{\"summary\":\"ok\"}"}}`))
	}))
	defer srv.Close()

	client := NewHTTPInferenceClient(srv.URL, 2*time.Second)
	response, err := client.Execute(context.Background(), InferenceRequest{
		Provider:         "local_ollama",
		Prompt:           "analyze this run",
		SystemPrompt:     "act as ACH analyst",
		InputJSON:        json.RawMessage(`{"batch_id":"B1"}`),
		PrimaryModel:     "ibm/granite3.3:8b",
		Temperature:      0.2,
		MaxTokens:        512,
		ExpectJSON:       true,
		EnableGuardrails: false,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if chatCalls != 1 {
		t.Fatalf("expected one ollama chat call, got %d", chatCalls)
	}
	if got := lastBody["model"]; got != "ibm/granite3.3:8b" {
		t.Fatalf("expected primary model in request, got %#v", got)
	}
	if got := lastBody["stream"]; got != false {
		t.Fatalf("expected stream=false, got %#v", got)
	}
	if got := lastBody["format"]; got != "json" {
		t.Fatalf("expected format=json, got %#v", got)
	}

	messages, ok := lastBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %#v", lastBody["messages"])
	}
	first, _ := messages[0].(map[string]any)
	second, _ := messages[1].(map[string]any)
	if first["role"] != "system" || first["content"] != "act as ACH analyst" {
		t.Fatalf("unexpected system message %#v", first)
	}
	if second["role"] != "user" {
		t.Fatalf("unexpected user message %#v", second)
	}
	if !strings.Contains(second["content"].(string), "Context JSON:\n{\"batch_id\":\"B1\"}") {
		t.Fatalf("expected context json in user content, got %q", second["content"])
	}

	if string(response.PrimaryOutput) != `{"summary":"ok"}` {
		t.Fatalf("unexpected primary output %s", response.PrimaryOutput)
	}
	if response.ModelTag != "ibm/granite3.3:8b" {
		t.Fatalf("unexpected model tag %q", response.ModelTag)
	}
}

func TestHTTPInferenceClientExecuteLocalOllamaGuardrailAndHelper(t *testing.T) {
	t.Helper()

	callsByModel := map[string]int{}
	guardrailUserMessage := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("expected /api/chat path, got %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer r.Body.Close()

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		model, _ := payload["model"].(string)
		callsByModel[model]++

		w.Header().Set("Content-Type", "application/json")
		switch model {
		case "ibm/granite3.3:8b":
			_, _ = w.Write([]byte(`{"model":"ibm/granite3.3:8b","message":{"role":"assistant","content":"{\"summary\":\"primary\"}"}}`))
		case "ibm/granite3.3-guardian:8b":
			if msgs, ok := payload["messages"].([]any); ok && len(msgs) > 1 {
				if msg, ok := msgs[1].(map[string]any); ok {
					guardrailUserMessage, _ = msg["content"].(string)
				}
			}
			_, _ = w.Write([]byte(`{"model":"ibm/granite3.3-guardian:8b","message":{"role":"assistant","content":"{\"decision\":\"allow\"}"}}`))
		case "granite4:3b":
			_, _ = w.Write([]byte(`{"model":"granite4:3b","message":{"role":"assistant","content":"{\"helper\":\"hints\"}"}}`))
		default:
			t.Fatalf("unexpected model %q", model)
		}
	}))
	defer srv.Close()

	client := NewHTTPInferenceClient("http://127.0.0.1:1", 2*time.Second)
	response, err := client.Execute(context.Background(), InferenceRequest{
		Provider:          "local_ollama",
		BaseURL:           srv.URL,
		Prompt:            "score this cycle",
		PrimaryModel:      "ibm/granite3.3:8b",
		GuardrailModel:    "ibm/granite3.3-guardian:8b",
		HelperModel:       "granite4:3b",
		ExpectJSON:        true,
		EnableGuardrails:  true,
		EnableHelperModel: true,
		HelperRequested:   true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if callsByModel["ibm/granite3.3:8b"] != 1 || callsByModel["ibm/granite3.3-guardian:8b"] != 1 || callsByModel["granite4:3b"] != 1 {
		t.Fatalf("unexpected call distribution %#v", callsByModel)
	}
	if string(response.PrimaryOutput) != `{"summary":"primary"}` {
		t.Fatalf("unexpected primary output %s", response.PrimaryOutput)
	}
	if string(response.GuardrailOutput) != `{"decision":"allow"}` {
		t.Fatalf("unexpected guardrail output %s", response.GuardrailOutput)
	}
	if string(response.HelperOutput) != `{"helper":"hints"}` {
		t.Fatalf("unexpected helper output %s", response.HelperOutput)
	}
	if !strings.Contains(guardrailUserMessage, "risk_criteria") {
		t.Fatalf("expected compact guardrail payload to include risk criteria, got %q", guardrailUserMessage)
	}
	if strings.Contains(guardrailUserMessage, "\"input_json\"") {
		t.Fatalf("expected compact guardrail payload without full input_json nesting, got %q", guardrailUserMessage)
	}
}

func TestHTTPInferenceClientExecuteLocalOllamaHelperFailureIsNonBlocking(t *testing.T) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("expected /api/chat path, got %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer r.Body.Close()

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		model, _ := payload["model"].(string)

		if model == "granite4:3b" {
			http.Error(w, "helper unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{\"summary\":\"ok\"}"}}`))
	}))
	defer srv.Close()

	client := NewHTTPInferenceClient(srv.URL, 2*time.Second)
	response, err := client.Execute(context.Background(), InferenceRequest{
		Provider:          "local_ollama",
		Prompt:            "score this run",
		PrimaryModel:      "ibm/granite3.3:8b",
		HelperModel:       "granite4:3b",
		ExpectJSON:        true,
		EnableHelperModel: true,
		HelperRequested:   true,
	})
	if err != nil {
		t.Fatalf("Execute() should not fail when helper fails, got %v", err)
	}
	if !strings.Contains(response.HelperText, "helper model call failed") {
		t.Fatalf("expected helper failure text, got %q", response.HelperText)
	}
	if string(response.PrimaryOutput) != `{"summary":"ok"}` {
		t.Fatalf("unexpected primary output %s", response.PrimaryOutput)
	}
}

func TestHTTPInferenceClientExecuteGatewayFallback(t *testing.T) {
	t.Helper()

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/inference/execute" {
			t.Fatalf("expected gateway execute path, got %s", r.URL.Path)
		}
		called = true

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"primary_output":{"route":"gateway"},"guardrail":{"decision":"pass"}}}`))
	}))
	defer srv.Close()

	client := NewHTTPInferenceClient(srv.URL, 2*time.Second)
	response, err := client.Execute(context.Background(), InferenceRequest{
		Provider:     "gateway",
		Prompt:       "route through gateway",
		PrimaryModel: "ibm/granite3.3:8b",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("expected gateway path to be called")
	}
	if string(response.PrimaryOutput) != `{"route":"gateway"}` {
		t.Fatalf("unexpected gateway primary output %s", response.PrimaryOutput)
	}
	if string(response.GuardrailOutput) != `{"decision":"pass"}` {
		t.Fatalf("expected guardrail alias to be normalized, got %s", response.GuardrailOutput)
	}
}

func TestHTTPInferenceClientExecuteLocalOllamaUsesPhaseTimeouts(t *testing.T) {
	t.Helper()

	callsByModel := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer r.Body.Close()

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		model, _ := payload["model"].(string)
		callsByModel[model]++

		switch model {
		case "ibm/granite3.3:8b":
			time.Sleep(30 * time.Millisecond)
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{\"summary\":\"primary\"}"}}`))
		case "ibm/granite3.3-guardian:8b":
			time.Sleep(90 * time.Millisecond)
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"{\"decision\":\"allow\"}"}}`))
		default:
			t.Fatalf("unexpected model %q", model)
		}
	}))
	defer srv.Close()

	client := NewHTTPInferenceClient(srv.URL, 5*time.Second)
	_, err := client.Execute(context.Background(), InferenceRequest{
		Provider:                "local_ollama",
		Prompt:                  "test timeout split",
		PrimaryModel:            "ibm/granite3.3:8b",
		GuardrailModel:          "ibm/granite3.3-guardian:8b",
		ExpectJSON:              true,
		EnableGuardrails:        true,
		PrimaryTimeoutSeconds:   1,
		GuardrailTimeoutSeconds: 0, // intentionally too short via fallback set below
	})
	if err != nil {
		t.Fatalf("expected first call to use default timeout and succeed, got %v", err)
	}

	_, err = client.Execute(context.Background(), InferenceRequest{
		Provider:                "local_ollama",
		Prompt:                  "test timeout split",
		PrimaryModel:            "ibm/granite3.3:8b",
		GuardrailModel:          "ibm/granite3.3-guardian:8b",
		ExpectJSON:              true,
		EnableGuardrails:        true,
		PrimaryTimeoutSeconds:   1,
		GuardrailTimeoutSeconds: 0,
		HelperTimeoutSeconds:    0,
	})
	if err != nil {
		t.Fatalf("expected second call to also succeed with default timeout, got %v", err)
	}

	clientTight := NewHTTPInferenceClient(srv.URL, 40*time.Millisecond)
	_, err = clientTight.Execute(context.Background(), InferenceRequest{
		Provider:                "local_ollama",
		Prompt:                  "test timeout split",
		PrimaryModel:            "ibm/granite3.3:8b",
		GuardrailModel:          "ibm/granite3.3-guardian:8b",
		ExpectJSON:              true,
		EnableGuardrails:        true,
		PrimaryTimeoutSeconds:   1,
		GuardrailTimeoutSeconds: 0,
	})
	if err == nil {
		t.Fatal("expected guardrail phase timeout error when fallback timeout is tight")
	}
	if callsByModel["ibm/granite3.3:8b"] < 3 {
		t.Fatalf("expected primary calls to complete before guardrail timeout, got %#v", callsByModel)
	}
}
