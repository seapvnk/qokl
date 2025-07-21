package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/server"
)

func setupTestDB(t *testing.T) http.Handler {
	srv := server.New("./")
	return srv.Router
}

// Checks if insert can be performed
func TestDBCanInsert(t *testing.T) {
	core.OpenDB("./.storage")
	defer os.RemoveAll("./.storage")
	router := setupTestDB(t)

	payload := `(insert %(admin user) name: "Pedro" age: 23)`
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(payload)))

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.Code)
	}

	var body map[string]any
	err := json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		t.Fatalf("Cannot decode: %v", err)
	}

	_, idOk := body["id"]
	if !idOk {
		t.Errorf("id need to be in the json")
	}

	// check name
	name, nameOk := body["name"]
	if !nameOk {
		t.Errorf("name need to be in the json")
	}

	switch nameVal := name.(type) {
	case string:
		if nameVal != "Pedro" {
			t.Errorf("expected name to be Pedro, get %s instedad", nameVal)
		}
	default:
		t.Errorf("expected name to be string, get %T instead", nameVal)
	}

	// check age
	age, ageOk := body["age"]
	if !ageOk {
		t.Errorf("age need to be in the json")
	}

	switch ageVal := age.(type) {
	case float64:
		if ageVal != 23 {
			t.Errorf("expected age to be 23, get %f instedad", ageVal)
		}
	default:
		t.Errorf("expected age to be float64, get %T instead", ageVal)
	}
}

// Check if select works and can filter values
func TestDBQueryFilterReturnsOnlyOneResult(t *testing.T) {
	core.OpenDB("./.storage")
	defer os.RemoveAll("./.storage")
	router := setupTestDB(t)

	payload1 := `(insert %(admin user) name: "Pedro" age: 23)`
	req1 := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(payload1)))
	resp1 := httptest.NewRecorder()
	router.ServeHTTP(resp1, req1)
	if resp1.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on first insert, got %d", resp1.Code)
	}

	payload2 := `(insert %(admin user) name: "Sergio" age: 23)`
	req2 := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(payload2)))
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)
	if resp2.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on second insert, got %d", resp2.Code)
	}

	query := `(select admin: (fn [e] (== "Pedro" (hget e %name))))`
	queryReq := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte(query)))
	queryResp := httptest.NewRecorder()
	router.ServeHTTP(queryResp, queryReq)
	if queryResp.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on query, got %d", queryResp.Code)
	}

	var result []map[string]any
	err := json.NewDecoder(queryResp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Error decoding query response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if name, ok := result[0]["name"].(string); !ok || name != "Pedro" {
		t.Errorf("Expected result name to be Pedro, got %v", result[0]["name"])
	}
}
