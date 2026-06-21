package utility

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockBuffer struct {
	*bytes.Buffer
	closed bool
}

func TestSendSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]interface{}{"id": 1, "name": "test"}
	SendSuccess(w, http.StatusOK, "success", data)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if result.Status != "success" {
		t.Errorf("status = %q, want %q", result.Status, "success")
	}
	if result.Message != "success" {
		t.Errorf("message = %q, want %q", result.Message, "success")
	}
	if result.Data == nil {
		t.Error("data should not be nil")
	}
}

func TestSendSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	SendSuccess(w, http.StatusOK, "ok", nil)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if result.Data != nil {
		t.Error("data should be nil")
	}
}

func TestSendError(t *testing.T) {
	w := httptest.NewRecorder()
	SendError(w, http.StatusNotFound, "resource not found")

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if result.Status != "error" {
		t.Errorf("status = %q, want %q", result.Status, "error")
	}
	if result.Message != "resource not found" {
		t.Errorf("message = %q, want %q", result.Message, "resource not found")
	}
}

func TestSendError_EmptyMessage(t *testing.T) {
	w := httptest.NewRecorder()
	SendError(w, http.StatusBadRequest, "")

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if result.Message != "" {
		t.Errorf("message should be empty")
	}
}

func TestSendErrorResponse_AppError(t *testing.T) {
	w := httptest.NewRecorder()
	err := NotFoundError("user not found")
	SendErrorResponse(w, err)

	resp := w.Result()
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSendErrorResponse_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	SendErrorResponse(w, nil)

	resp := w.Result()
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestSendErrorResponse_PlainError(t *testing.T) {
	w := httptest.NewRecorder()
	SendErrorResponse(w, http.ErrBodyNotAllowed)

	resp := w.Result()
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestResponse_JSONStructure(t *testing.T) {
	w := httptest.NewRecorder()
	SendSuccess(w, http.StatusOK, "test", []int{1, 2, 3})

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if !bytes.Contains(body, []byte(`"status":"success"`)) {
		t.Error("response should contain status field")
	}
	if !bytes.Contains(body, []byte(`"message":"test"`)) {
		t.Error("response should contain message field")
	}
	if !bytes.Contains(body, []byte(`"data"`)) {
		t.Error("response should contain data field")
	}
}
