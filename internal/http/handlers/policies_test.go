package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"clawbot-server/internal/platform/policies"
)

type policiesServiceStub struct{}

func (policiesServiceStub) List(context.Context) ([]policies.Policy, error) { return nil, nil }
func (policiesServiceStub) Get(context.Context, string) (policies.Policy, error) {
	return policies.Policy{}, nil
}
func (policiesServiceStub) Create(_ context.Context, input policies.CreateInput, _ string) (policies.Policy, error) {
	return policies.Policy{ID: "policy-1", Name: input.Name}, nil
}
func (policiesServiceStub) Update(_ context.Context, id string, _ policies.UpdateInput, _ string) (policies.Policy, error) {
	return policies.Policy{ID: id, Name: "updated"}, nil
}

func TestPoliciesHandlerCreate(t *testing.T) {
	handler := NewPoliciesHandler(policiesServiceStub{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies", bytes.NewBufferString(`{"name":"baseline-policy"}`))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}
