package controllers

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"forum/utils"
)

func TestPostHandler_getAllCategories(t *testing.T) {
	// Setup test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	defer testDB.Close()

	// Store original db and restore after test
	oldDB := utils.GlobalDB
	utils.GlobalDB = testDB
	defer func() { utils.GlobalDB = oldDB }()

	// Create test table
	_, err = testDB.Exec(`
        CREATE TABLE categories (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL
        )
    `)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tests := []struct {
		name    string
		setup   func(*testing.T, *sql.DB)
		want    []utils.Category
		wantErr bool
	}{
		{
			name:    "Empty categories",
			setup:   func(t *testing.T, db *sql.DB) {},
			want:    []utils.Category{},
			wantErr: false,
		},
		{
			name: "Single category",
			setup: func(t *testing.T, db *sql.DB) {
				_, err := db.Exec("INSERT INTO categories (id, name) VALUES (1, 'Technology')")
				if err != nil {
					t.Fatalf("Failed to insert test data: %v", err)
				}
			},
			want:    []utils.Category{{ID: 1, Name: "Technology"}},
			wantErr: false,
		},
		{
			name: "Multiple categories",
			setup: func(t *testing.T, db *sql.DB) {
				_, err := db.Exec(`
                    INSERT INTO categories (id, name) VALUES 
                    (1, 'Technology'),
                    (2, 'Sports'),
                    (3, 'Politics')
                `)
				if err != nil {
					t.Fatalf("Failed to insert test data: %v", err)
				}
			},
			want: []utils.Category{
				{ID: 1, Name: "Technology"},
				{ID: 2, Name: "Sports"},
				{ID: 3, Name: "Politics"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear table before each test
			testDB.Exec("DELETE FROM categories")

			// Setup test data
			tt.setup(t, testDB)

			ph := &PostHandler{}
			got, err := ph.getAllCategories()

			if (err != nil) != tt.wantErr {
				t.Errorf("getAllCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("getAllCategories() got %v items, want %v items", len(got), len(tt.want))
				return
			}

			for i, category := range got {
				if category.ID != tt.want[i].ID || category.Name != tt.want[i].Name {
					t.Errorf("getAllCategories()[%d] = %v, want %v", i, category, tt.want[i])
				}
			}
		})
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Just now",
			input:    now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "One minute ago",
			input:    now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "Multiple minutes ago",
			input:    now.Add(-45 * time.Minute),
			expected: "45 minutes ago",
		},
		{
			name:     "One hour ago",
			input:    now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "Multiple hours ago",
			input:    now.Add(-5 * time.Hour),
			expected: "5 hours ago",
		},
		{
			name:     "Yesterday",
			input:    now.Add(-25 * time.Hour),
			expected: "yesterday",
		},

		{
			name:     "Multiple days ago",
			input:    now.Add(-5 * 24 * time.Hour),
			expected: "5 days ago",
		},
		{
			name:     "One week ago",
			input:    now.Add(-7 * 24 * time.Hour),
			expected: "1 week ago",
		},
		{
			name:     "Multiple weeks ago",
			input:    now.Add(-3 * 7 * 24 * time.Hour),
			expected: "3 weeks ago",
		},
		{
			name:     "More than a month ago",
			input:    now.Add(-45 * 24 * time.Hour),
			expected: now.Add(-45 * 24 * time.Hour).Format("Jan 2, 2006"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatTimeAgo(tc.input)
			if result != tc.expected {
				t.Errorf("FormatTimeAgo(%v) = %s; want %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCheckAuthStatus(t *testing.T) {
	// Setup database for testing
	db, err := utils.InitialiseDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a PostHandler instance
	postHandler := &PostHandler{}

	// Create a test user and session
	userID := "test_user_123"
	sessionToken, err := utils.CreateSession(db, userID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	testCases := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedResult bool
	}{
		{
			name: "Valid Session",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: sessionToken,
				})
				return req
			},
			expectedResult: true,
		},
		{
			name: "No Session Cookie",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				return req
			},
			expectedResult: false,
		},
		{
			name: "Invalid Session Token",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: "invalid_token",
				})
				return req
			},
			expectedResult: false,
		},
	}

	// Set the global database for session validation
	utils.GlobalDB = db

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.setupRequest()
			result := postHandler.checkAuthStatus(req)

			if result != tc.expectedResult {
				t.Errorf("checkAuthStatus() = %v; want %v", result, tc.expectedResult)
			}
		})
	}

	// Clean up: delete the test session
	_, err = db.Exec("DELETE FROM sessions WHERE id = ?", sessionToken)
	if err != nil {
		t.Fatalf("Failed to clean up test session: %v", err)
	}
}

func TestGetCategoryIDByName(t *testing.T) {
	// Setup database for testing
	db, err := utils.InitialiseDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set the global database for the function
	utils.GlobalDB = db

	// First, delete any existing test categories to ensure clean state
	_, err = db.Exec(`
		DELETE FROM categories 
		WHERE name IN ('Technology', 'Sports', 'Music')
	`)
	if err != nil {
		t.Fatalf("Failed to clean up existing categories: %v", err)
	}

	// Insert test categories
	_, err = db.Exec(`
		INSERT INTO categories (name) VALUES 
		('Technology'), 
		('Sports'), 
		('Music')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test categories: %v", err)
	}

	// Dynamically retrieve the IDs of the inserted categories
	var techID, sportsID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = 'Technology'").Scan(&techID)
	if err != nil {
		t.Fatalf("Failed to retrieve Technology category ID: %v", err)
	}

	err = db.QueryRow("SELECT id FROM categories WHERE name = 'Sports'").Scan(&sportsID)
	if err != nil {
		t.Fatalf("Failed to retrieve Sports category ID: %v", err)
	}

	testCases := []struct {
		name         string
		categoryName string
		expectedID   int
		expectError  bool
	}{
		{
			name:         "Existing Category - Technology",
			categoryName: "Technology",
			expectedID:   techID,
			expectError:  false,
		},
		{
			name:         "Existing Category - Sports",
			categoryName: "Sports",
			expectedID:   sportsID,
			expectError:  false,
		},
		{
			name:         "Non-Existent Category",
			categoryName: "NonExistentCategory",
			expectedID:   0,
			expectError:  true,
		},
		{
			name:         "Empty Category Name",
			categoryName: "",
			expectedID:   0,
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			categoryID, err := getCategoryIDByName(tc.categoryName)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for category '%s', but got none", tc.categoryName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for category '%s': %v", tc.categoryName, err)
				}

				if categoryID != tc.expectedID {
					t.Errorf("getCategoryIDByName(%s) = %d; want %d",
						tc.categoryName, categoryID, tc.expectedID)
				}
			}
		})
	}

	// Clean up: remove test categories
	_, err = db.Exec(`
		DELETE FROM categories 
		WHERE name IN ('Technology', 'Sports', 'Music')
	`)
	if err != nil {
		t.Fatalf("Failed to clean up test categories: %v", err)
	}
}
