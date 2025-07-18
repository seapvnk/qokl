package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/server"
	"github.com/seapvnk/qokl/tasks"
)

func setupTestTask(t *testing.T) (http.Handler, *tasks.Listener) {
	srv := server.New("./")
	listener := tasks.New("./")
	return srv.Router, listener
}

// Checks if a simple get can be performed
func TestPeformTaskAndSaveCache(t *testing.T) {
	core.OpenStore()
	router, listener := setupTestTask(t)
	go listener.Run()
	defer listener.Close()
	defer core.CloseStore()

	// send a request to dispatch a task that will store the cache
	payload := `{"value": "it works"}`
	req := httptest.NewRequest("POST", "/api/test-cache", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.Code)
	}

	expected := `{"message":"ok"`
	if !strings.Contains(resp.Body.String(), expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, resp.Body.String())
	}

	time.Sleep(150 * time.Millisecond)

	// expects to see cache value in get endpoint
	req = httptest.NewRequest("GET", "/api/test-cache", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.Code)
	}

	expected = `{"data":"it works"`
	if !strings.Contains(resp.Body.String(), expected) {
		t.Errorf("Expected response to contain %q, got %q", expected, resp.Body.String())
	}
}
