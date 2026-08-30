package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConnectionMiddleware(t *testing.T) {
	// Create middleware instance
	middleware := NewConnectionMiddleware(true)

	// Create a test handler that takes some time
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate some processing time
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("test response"))
		if err != nil {
			t.Errorf("Write operation failed %e", err)
		}
	})

	// Wrap the test handler with the middleware
	wrappedHandler := middleware.Middleware(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	// Execute the request
	start := time.Now()

	wrappedHandler.ServeHTTP(recorder, req)

	elapsed := time.Since(start)

	// Verify the response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	if recorder.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", recorder.Body.String())
	}

	// Verify that some time has passed (at least the sleep duration)
	if elapsed < 10*time.Millisecond {
		t.Errorf("Expected elapsed time to be at least 10ms, got %v", elapsed)
	}
}

func TestConnectionMiddleware_StatusCodeCapture(t *testing.T) {
	middleware := NewConnectionMiddleware(true)

	// Test different status codes
	testCases := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusInternalServerError,
		http.StatusNotFound,
	}

	for _, expectedStatus := range testCases {
		t.Run(http.StatusText(expectedStatus), func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(expectedStatus)
			})

			wrappedHandler := middleware.Middleware(testHandler)
			req := httptest.NewRequest("POST", "/test", nil)
			recorder := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(recorder, req)

			if recorder.Code != expectedStatus {
				t.Errorf("Expected status code %d, got %d", expectedStatus, recorder.Code)
			}
		})
	}
}
