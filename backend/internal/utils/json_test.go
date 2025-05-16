package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		msg     string
		err     error
		wantLog []string
	}{
		{
			name:    "Bad request without err",
			code:    http.StatusBadRequest,
			msg:     "bad request",
			err:     nil,
			wantLog: []string{},
		},
		{
			name:    "Bad request with err",
			code:    http.StatusBadRequest,
			msg:     "bad request",
			err:     errors.New("boom"),
			wantLog: []string{"boom"},
		},
		{
			name:    "5xx without err",
			code:    http.StatusInternalServerError,
			msg:     "oops",
			err:     nil,
			wantLog: []string{"Responding with 5XX error: oops"},
		},
		{
			name:    "5xx with err",
			code:    http.StatusInternalServerError,
			msg:     "down",
			err:     errors.New("timeout"),
			wantLog: []string{"timeout", "Responding with 5XX error: down"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			// Capture logs
			var buf bytes.Buffer
			old := log.Writer()
			defer log.SetOutput(old)
			log.SetOutput(&buf)

			RespondWithError(rec, tt.code, tt.msg, tt.err)

			if rec.Code != tt.code {
				t.Errorf("status: got %d, want %d", rec.Code, tt.code)
			}
			var got struct{ Error string }
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if got.Error != tt.msg {
				t.Errorf("body: got %#v, want %#v", got.Error, tt.msg)
			}

			logs := buf.String()
			for _, want := range tt.wantLog {
				if !strings.Contains(logs, want) {
					t.Errorf("logs missing %q in %q", want, logs)
				}
			}
			if len(tt.wantLog) == 0 && buf.Len() != 0 {
				t.Errorf("expected no logs, got %q", logs)
			}

		})
	}
}

type brokenWriter struct {
	Hdr http.Header
}

func (b *brokenWriter) Header() http.Header       { return b.Hdr }
func (b *brokenWriter) WriteHeader(int)           {}
func (b *brokenWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }
func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name            string
		code            int
		payload         interface{}
		useBrokenWriter bool
		key             string
		value           string
		wantLog         []string
	}{
		{
			name:            "Successful JSON response",
			code:            http.StatusOK,
			payload:         map[string]string{"FirstName": "John"},
			useBrokenWriter: false,
			key:             "FirstName",
			value:           "John",
			wantLog:         nil,
		},
		{
			name:            "Error marshalling JSON",
			code:            http.StatusInternalServerError,
			payload:         make(chan int),
			useBrokenWriter: false,
			key:             "FirstName",
			value:           "John",
			wantLog:         []string{"Error marshalling JSON: "},
		},
		{
			name:            "Error writing response",
			code:            http.StatusOK,
			payload:         map[string]string{"FirstName": "John"},
			useBrokenWriter: true,
			key:             "FirstName",
			value:           "John",
			wantLog:         []string{"Error writing response: "},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			var buf bytes.Buffer
			old := log.Writer()
			defer log.SetOutput(old)
			log.SetOutput(&buf)

			if tt.useBrokenWriter {
				bw := &brokenWriter{Hdr: make(http.Header)}
				RespondWithJSON(bw, tt.code, tt.payload)
			} else {
				RespondWithJSON(rec, tt.code, tt.payload)

				if rec.Code != tt.code {
					t.Errorf("status: got %d, want %d", rec.Code, tt.code)
				}
				if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q; want %q", ct, "application/json")
				}
				if _, isChan := tt.payload.(chan int); isChan {
					if rec.Body.Len() != 0 {
						t.Errorf("expected empty body on marshal error, but got %q", rec.Body.String())
					}
				} else {
					var got map[string]string
					if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
						t.Fatalf("invalid JSON: %v", err)
					}
					if got[tt.key] != tt.value {
						t.Errorf("body[%q] = %q; want %q", tt.key, got[tt.key], tt.value)
					}
				}
			}

			logs := buf.String()
			for _, want := range tt.wantLog {
				if !strings.Contains(logs, want) {
					t.Errorf("logs missing %q in %q", want, logs)
				}
			}
			if len(tt.wantLog) == 0 && buf.Len() != 0 {
				t.Errorf("expected no logs, got %q", logs)
			}

		})
	}
}
