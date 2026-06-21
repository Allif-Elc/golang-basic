package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil) // reset after test

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Error("log should contain HTTP method")
	}
	if !strings.Contains(output, "/api/test") {
		t.Error("log should contain URL path")
	}
	if !strings.Contains(output, "200") {
		t.Error("log should contain status code 200")
	}
}

func TestRequestLogger_Post(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	req.RemoteAddr = "127.0.0.1:5678"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	output := buf.String()
	if !strings.Contains(output, "POST") {
		t.Error("log should contain POST method")
	}
	if !strings.Contains(output, "201") {
		t.Error("log should contain status code 201")
	}
}

func TestProfiler(t *testing.T) {
	handler := Profiler()
	if handler == nil {
		t.Fatal("Profiler() should not return nil")
	}
}
