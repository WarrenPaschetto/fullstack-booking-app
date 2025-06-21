package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSHandlesPreflight(t *testing.T) {
	allowed := []string{"http://localhost:3000"}

	nextCalled := false
	dummyNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := CORS(allowed)(dummyNext)

	req := httptest.NewRequest(http.MethodOptions, "/some-path", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for preflight, got %d", resp.StatusCode)
	}

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("wrong Access-Control-Allow-Origin: got %q, want %q", origin, "http://localhost:3000")
	}

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	wantMethods := "GET, POST, PUT, DELETE, OPTIONS"
	if methods != wantMethods {
		t.Errorf("wrong Access-Control-Allow-Methods: got %q, want %q", methods, wantMethods)
	}

	headers := resp.Header.Get("Access-Control-Allow-Headers")
	wantHeaders := "Content-Type, Authorization"
	if headers != wantHeaders {
		t.Errorf("wrong Access-Control-Allow-Headers: got %q, want %q", headers, wantHeaders)
	}

	if nextCalled {
		t.Error("expected next handler NOT to be called for OPTIONS, but it was")
	}
}

func TestCORSPassesThroughNonOptions(t *testing.T) {
	allowed := []string{"http://localhost:3000"}

	nextCalled := false
	dummyNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("next ran"))
	})

	handler := CORS(allowed)(dummyNext)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("expected status 418 from next, got %d", resp.StatusCode)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called for non-OPTIONS, but it was not")
	}

	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("wrong Access-Control-Allow-Origin: got %q, want %q", origin, "http://localhost:3000")
	}

	methods := resp.Header.Get("Access-Control-Allow-Methods")
	wantMethods := "GET, POST, PUT, DELETE, OPTIONS"
	if methods != wantMethods {
		t.Errorf("wrong Access-Control-Allow-Methods: got %q, want %q", methods, wantMethods)
	}

	headers := resp.Header.Get("Access-Control-Allow-Headers")
	wantHeaders := "Content-Type, Authorization"
	if headers != wantHeaders {
		t.Errorf("wrong Access-Control-Allow-Headers: got %q, want %q", headers, wantHeaders)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "next ran") {
		t.Errorf("expected body to include %q, got %q", "next ran", body)
	}
}
