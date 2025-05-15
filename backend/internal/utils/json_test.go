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
