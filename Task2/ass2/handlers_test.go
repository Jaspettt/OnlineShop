package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	// Create a request with mock user data
	req, err := http.NewRequest("POST", "/register", strings.NewReader("username=testuser&email=test@example.com&password=123456"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the register handler function directly with the mock request and response recorder
	registerHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "Registration successful! Please check your email to confirm your account."
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestConfirmHandler(t *testing.T) {
	// Create a request with a valid token
	req, err := http.NewRequest("GET", "/confirm?token=valid_token", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the confirm handler function directly with the mock request and response recorder
	confirmHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Check the redirect URL
	expectedRedirectURL := "/login?confirmation=success"
	if location := rr.Header().Get("Location"); location != expectedRedirectURL {
		t.Errorf("handler returned unexpected redirect location: got %v want %v", location, expectedRedirectURL)
	}
}
