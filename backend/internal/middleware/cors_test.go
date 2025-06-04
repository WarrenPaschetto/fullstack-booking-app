package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSHandlesPreflight(t *testing.T) {
	// Create a dummy “next” handler that should never be called for OPTIONS
	nextCalled := false
	dummyNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	// Wrap dummyNext with CORS middleware
	handler := CORS(dummyNext)

	// Build an OPTIONS (preflight) request
	req := httptest.NewRequest(http.MethodOptions, "/some-path", nil)
	// We need an Origin header, since browsers include it in preflight.
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	// 1. For OPTIONS, the middleware should immediately return 200 status:
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for preflight, got %d", resp.StatusCode)
	}

	// 2. It should set Access-Control-Allow-Origin correctly:
	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("wrong Access-Control-Allow-Origin: got %q, want %q", origin, "http://localhost:3000")
	}

	// 3. It should set the allowed methods header:
	methods := resp.Header.Get("Access-Control-Allow-Methods")
	wantMethods := "GET, POST, PUT, DELETE, OPTIONS"
	if methods != wantMethods {
		t.Errorf("wrong Access-Control-Allow-Methods: got %q, want %q", methods, wantMethods)
	}

	// 4. It should set the allowed headers header:
	headers := resp.Header.Get("Access-Control-Allow-Headers")
	wantHeaders := "Content-Type, Authorization"
	if headers != wantHeaders {
		t.Errorf("wrong Access-Control-Allow-Headers: got %q, want %q", headers, wantHeaders)
	}

	// 5. Next handler must NOT have been called:
	if nextCalled {
		t.Error("expected next handler NOT to be called for OPTIONS, but it was")
	}
}

func TestCORSPassesThroughNonOptions(t *testing.T) {
	// Create a dummy “next” handler that simply writes a known response
	nextCalled := false
	dummyNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot) // any non‐200 code to prove this ran
		w.Write([]byte("next ran"))
	})

	// Wrap dummyNext with CORS middleware
	handler := CORS(dummyNext)

	// Build a GET request
	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	// Include Origin to simulate a real browser request
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	// 1. Since this is GET, middleware should call next and preserve its status code (418):
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("expected status 418 from next, got %d", resp.StatusCode)
	}

	// 2. Next handler must have been called:
	if !nextCalled {
		t.Fatal("expected next handler to be called for non-OPTIONS, but it was not")
	}

	// 3. CORS headers must still be set on the response:
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

	// 4. The body from next should be present:
	body := rr.Body.String()
	if !strings.Contains(body, "next ran") {
		t.Errorf("expected body to include %q, got %q", "next ran", body)
	}
}
