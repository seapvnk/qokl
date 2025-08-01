package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seapvnk/qokl/server"
)

func setupTestServer(t *testing.T) http.Handler {
	srv := server.New("./")
	return srv.Router
}

// Checks if a simple get can be performed
func TestApiCanGet(t *testing.T) {
	router := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/hello-world", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.Code)
	}

	expected := `{"message":"hello, world!"}`
	if !strings.Contains(resp.Body.String(), expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, resp.Body.String())
	}
}

// Checks if can use imports
func TestApiCanUseLib(t *testing.T) {
	router := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/with-package", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.Code)
	}

	expected := `{"add":4,"mul":4}`
	if !strings.Contains(resp.Body.String(), expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, resp.Body.String())
	}
}

// Checks if its possible to retrieve a url param
func TestApiCanGetParam(t *testing.T) {
	router := setupTestServer(t)

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Say hello to pedro",
			url:      "/api/hello/pedro",
			expected: `{"message":"hello, pedro!"}`,
		},
		{
			name:     "Say hello to pedro and repeat his age",
			url:      "/api/hello/pedro/23",
			expected: `{"message":"hello, pedro, you are 23!"}`,
		},
		{
			name:     "Say hi to pedro",
			url:      "/api/hello/pedro/say-hi",
			expected: `{"message":"hi, pedro!"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Expected status 200 OK, got %d", resp.Code)
			}

			if !strings.Contains(resp.Body.String(), tt.expected) {
				t.Errorf("Expected response to contain %q, got %q", tt.expected, resp.Body.String())
			}
		})
	}
}

func TestApiPostCanUseBody(t *testing.T) {
	router := setupTestServer(t)

	payload := `{"email": "myemail@mail.com", "password": "mypasswd"}`
	req := httptest.NewRequest("POST", "/api/hello", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.Code)
	}

	expected := `{"message":"your email ismyemail@mail.com, and your password is mypasswd"`
	if !strings.Contains(resp.Body.String(), expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, resp.Body.String())
	}
}

func TestApiGetCanUseHeader(t *testing.T) {
	router := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/hello", nil)
	req.Header.Set("Authorization", "ok")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.Code)
	}

	expected := `{"message":"your token is: ok"}`
	if !strings.Contains(resp.Body.String(), expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, resp.Body.String())
	}
}
