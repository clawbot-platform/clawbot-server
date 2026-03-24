package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"clawbot-server/internal/platform/bots"
)

type botsServiceStub struct{}

func (botsServiceStub) List(context.Context) ([]bots.Bot, error) { return nil, nil }
func (botsServiceStub) Get(context.Context, string) (bots.Bot, error) {
	return bots.Bot{}, nil
}
func (botsServiceStub) Create(_ context.Context, input bots.CreateInput, _ string) (bots.Bot, error) {
	return bots.Bot{ID: "bot-1", Name: input.Name, Status: "active"}, nil
}
func (botsServiceStub) Update(_ context.Context, id string, _ bots.UpdateInput, _ string) (bots.Bot, error) {
	return bots.Bot{ID: id, Name: "updated", Status: "inactive"}, nil
}

func TestBotsHandlerCreate(t *testing.T) {
	handler := NewBotsHandler(botsServiceStub{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bots", bytes.NewBufferString(`{"name":"registry","status":"active"}`))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}
