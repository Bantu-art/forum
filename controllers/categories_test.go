package controllers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forum/utils"
)

// setupTestDB initializes an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	_, err = testDB.Exec(`
		CREATE TABLE categories (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT,
			content TEXT,
			imagepath TEXT,
			post_at DATETIME,
			user_id INTEGER
		);
		CREATE TABLE post_categories (
			post_id INTEGER,
			category_id INTEGER
		);
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT,
			profile_pic TEXT
		);
		CREATE TABLE reaction (
			post_id INTEGER,
			like INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return testDB
}

func TestCategoryHandler_handleCreateCategory(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	utils.GlobalDB = testDB

	ch := NewCategoryHandler()

	form := strings.NewReader("name=Programming")
	req, err := http.NewRequest("POST", "/categories", form)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ch.handleCreateCategory)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
}
func TestCategoryHandler_getAllCategories(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	utils.GlobalDB = testDB

	ch := NewCategoryHandler()

	_, err := testDB.Exec("INSERT INTO categories (id, name) VALUES (1, 'Programming')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	categories, err := ch.getAllCategories()
	if err != nil {
		t.Fatalf("getAllCategories returned error: %v", err)
	}

	if len(categories) != 1 {
		t.Errorf("getAllCategories returned %v categories, want 1", len(categories))
	}
}
