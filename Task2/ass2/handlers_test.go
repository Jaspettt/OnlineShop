package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHandler(t *testing.T) {

	req, err := http.NewRequest("POST", "/register", strings.NewReader("username=testuser&email=test@example.com&password=123456"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	registerHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Registration successful! Please check your email to confirm your account."
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestConfirmHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "/confirm?token=valid_token", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	confirmHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	expectedRedirectURL := "/login?confirmation=success"
	if location := rr.Header().Get("Location"); location != expectedRedirectURL {
		t.Errorf("handler returned unexpected redirect location: got %v want %v", location, expectedRedirectURL)
	}
}
