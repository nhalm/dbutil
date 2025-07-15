package gen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQueryParser_ParseQueries_FileIO(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create test SQL files
	testFiles := map[string]string{
		"users.sql": `-- name: GetUser :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email FROM users ORDER BY name;

-- name: CreateUser :exec
INSERT INTO users (name, email) VALUES ($1, $2);`,

		"posts.sql": `-- name: GetPostsByUser :many
SELECT id, title, content FROM posts WHERE user_id = $1;

-- name: GetPostsPaginated :paginated
SELECT id, title, content FROM posts ORDER BY id ASC LIMIT $1;`,
	}

	// Write test files
	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Test parsing
	parser := NewQueryParser(tempDir)
	queries, err := parser.ParseQueries()
	if err != nil {
		t.Fatalf("ParseQueries() failed: %v", err)
	}

	// Verify results
	if len(queries) != 5 {
		t.Errorf("Expected 5 queries, got %d", len(queries))
	}

	// Check specific queries (order may vary)
	expectedQueries := map[string]QueryType{
		"GetUser":           QueryTypeOne,
		"ListUsers":         QueryTypeMany,
		"CreateUser":        QueryTypeExec,
		"GetPostsByUser":    QueryTypeMany,
		"GetPostsPaginated": QueryTypePaginated,
	}

	queryMap := make(map[string]Query)
	for _, query := range queries {
		queryMap[query.Name] = query
	}

	for name, expectedType := range expectedQueries {
		query, exists := queryMap[name]
		if !exists {
			t.Errorf("Query %s not found", name)
			continue
		}

		if query.Type != expectedType {
			t.Errorf("Query %s: expected type %v, got %v", name, expectedType, query.Type)
		}

		if query.SQL == "" {
			t.Errorf("Query %s: SQL should not be empty", name)
		}
	}
}

func TestQueryParser_FindSQLFiles_FileIO(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create subdirectories
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files
	testFiles := []string{
		"users.sql",
		"posts.sql",
		"admin.SQL",  // uppercase extension
		"readme.txt", // non-SQL file
		filepath.Join("subdir", "nested.sql"),
	}

	for _, filename := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		err := os.WriteFile(fullPath, []byte("-- test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Test finding SQL files
	parser := NewQueryParser(tempDir)
	sqlFiles, err := parser.findSQLFiles()
	if err != nil {
		t.Fatalf("findSQLFiles() failed: %v", err)
	}

	// Should find 4 SQL files (users.sql, posts.sql, admin.SQL, subdir/nested.sql)
	expectedCount := 4
	if len(sqlFiles) != expectedCount {
		t.Errorf("Expected %d SQL files, got %d", expectedCount, len(sqlFiles))
	}

	// Check that non-SQL files are excluded
	for _, file := range sqlFiles {
		if filepath.Base(file) == "readme.txt" {
			t.Errorf("Non-SQL file should not be included: %s", file)
		}
	}
}

func TestQueryParser_ParseFile_EdgeCases_FileIO(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectCount int
	}{
		{
			name:        "empty_file",
			content:     "",
			expectError: false,
			expectCount: 0,
		},
		{
			name: "only_comments",
			content: `-- This is a comment
-- Another comment
-- Yet another comment`,
			expectError: false,
			expectCount: 0,
		},
		{
			name: "annotation_without_SQL",
			content: `-- name: GetUser :one
-- This is just a comment`,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "multiple_queries",
			content: `-- name: GetUser :one
SELECT id, name FROM users WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (name, email) VALUES ($1, $2);`,
			expectError: false,
			expectCount: 2,
		},
		{
			name: "query_with_complex_SQL",
			content: `-- name: GetUserWithPosts :many
SELECT u.id, u.name, p.title
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.created_at > $1
ORDER BY u.name, p.created_at DESC;`,
			expectError: false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			filename := filepath.Join(tempDir, "test.sql")
			err := os.WriteFile(filename, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse the file
			parser := NewQueryParser(tempDir)
			queries, err := parser.parseFile(filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
				if len(queries) != tt.expectCount {
					t.Errorf("Expected %d queries, got %d", tt.expectCount, len(queries))
				}
			}

			// Clean up
			os.Remove(filename)
		})
	}
}

func TestQueryParser_ErrorHandling_FileIO(t *testing.T) {
	t.Run("nonexistent directory", func(t *testing.T) {
		parser := NewQueryParser("/nonexistent/directory")
		_, err := parser.ParseQueries()
		if err == nil {
			t.Errorf("Expected error for nonexistent directory")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		parser := NewQueryParser(tempDir)
		_, err := parser.ParseQueries()
		if err == nil {
			t.Errorf("Expected error for empty directory")
		}
	})

	t.Run("empty queries directory", func(t *testing.T) {
		tempDir := t.TempDir()
		parser := NewQueryParser(tempDir)
		_, err := parser.ParseQueries()
		if err == nil {
			t.Errorf("Expected error for directory with no SQL files")
		}
	})
}
