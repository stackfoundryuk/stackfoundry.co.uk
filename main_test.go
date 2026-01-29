package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
	// Initialize the router
	router := setupRouter()

	// Define test cases
	tests := []struct {
		name           string
		method         string
		target         string
		body           io.Reader
		contentType    string
		expectedStatus int
		expectedText   string
		expectedCT     string // Check for the Content-Type header fix
	}{
		{
			name:           "Home Page",
			method:         "GET",
			target:         "/",
			body:           nil,
			expectedStatus: http.StatusOK,
			expectedText:   "StackFoundry",
			expectedCT:     "text/html; charset=utf-8",
		},
		{
			name:           "Privacy Policy",
			method:         "GET",
			target:         "/privacy",
			body:           nil,
			expectedStatus: http.StatusOK,
			expectedText:   "Privacy",
			expectedCT:     "text/html; charset=utf-8",
		},
		{
			name:           "NotFound",
			method:         "GET",
			target:         "/made-up-url",
			body:           nil,
			expectedStatus: http.StatusNotFound, // We explicitly set 404 now
			expectedText:   "404",
			expectedCT:     "text/html; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, tt.body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// 1. Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// 2. Check Content-Type Header (The Fix)
			if ct := rr.Header().Get("Content-Type"); ct != tt.expectedCT {
				t.Errorf("handler returned wrong content-type: got %v want %v",
					ct, tt.expectedCT)
			}

			// 3. Check Body Content
			if !strings.Contains(rr.Body.String(), tt.expectedText) {
				t.Errorf("handler returned unexpected body: got %v want substring %v",
					rr.Body.String(), tt.expectedText)
			}
		})
	}
}

func TestContactFormSubmission(t *testing.T) {
	router := setupRouter()

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("subject", "Test Subject")
	form.Add("message", "This is a test message from main_test.go")

	req := httptest.NewRequest("POST", "/api/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 1. Verify 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Contact handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	// 2. Verify Content-Type (HTMX expects HTML partials)
	expectedCT := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expectedCT {
		t.Errorf("Contact handler returned wrong content-type: got %v want %v",
			ct, expectedCT)
	}

	// 3. Verify Success Component Rendered
	expected := "Transmission Received"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("Contact handler did not render success message: got body %v", rr.Body.String())
	}
}
