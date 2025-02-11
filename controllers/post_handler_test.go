package controllers

import (
	"database/sql"
	"testing"

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
