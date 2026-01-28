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
	// Initialize the router using the function we extracted
	router := setupRouter()

	// Define test cases
	tests := []struct {
		name           string
		method         string
		target         string
		body           io.Reader
		contentType    string
		expectedStatus int
		expectedText   string // Substring we expect to see in response
	}{
		{
			name:           "Home Page",
			method:         "GET",
			target:         "/",
			body:           nil,
			expectedStatus: http.StatusOK,
			expectedText:   "StackFoundry", // Expect brand name in HTML
		},
		{
			name:           "Privacy Policy",
			method:         "GET",
			target:         "/privacy",
			body:           nil,
			expectedStatus: http.StatusOK,
			expectedText:   "Privacy",
		},
		{
			name:           "NotFound",
			method:         "GET",
			target:         "/made-up-url",
			body:           nil,
			expectedStatus: http.StatusOK, // Your NotFound handler returns 200 with 404 content usually, or you can change to 404
			expectedText:   "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, tt.body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
			rr := httptest.NewRecorder()

			// Call the handler directly
			router.ServeHTTP(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check Body Content
			if !strings.Contains(rr.Body.String(), tt.expectedText) {
				t.Errorf("handler returned unexpected body: got %v want substring %v",
					rr.Body.String(), tt.expectedText)
			}
		})
	}
}

func TestContactFormSubmission(t *testing.T) {
	router := setupRouter()

	// Prepare form data
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("subject", "Test Subject")
	form.Add("message", "This is a test message from main_test.go")

	// Create POST request
	req := httptest.NewRequest("POST", "/api/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	// Execute Request
	router.ServeHTTP(rr, req)

	// 1. Verify 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Contact handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	// 2. Verify Success Component Rendered
	// We expect the "Transmission Received" message from ContactSuccess()
	expected := "Transmission Received"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("Contact handler did not render success message: got body %v", rr.Body.String())
	}
}
