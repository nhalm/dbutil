package gen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQueryParser_ParseQueries(t *testing.T) {
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

	// Check specific queries
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
			t.Errorf("Query %s: expected type %s, got %s", name, expectedType, query.Type)
		}

		if query.SQL == "" {
			t.Errorf("Query %s: SQL is empty", name)
		}

		if query.SourceFile == "" {
			t.Errorf("Query %s: SourceFile is empty", name)
		}
	}
}

func TestQueryParser_ParseAnnotation(t *testing.T) {
	parser := NewQueryParser("")

	tests := []struct {
		name     string
		line     string
		expected *QueryAnnotation
	}{
		{
			name:     "basic annotation",
			line:     "-- name: GetUser :one",
			expected: &QueryAnnotation{Name: "GetUser", Type: QueryTypeOne},
		},
		{
			name:     "annotation with extra spaces",
			line:     "--   name:   GetUser   :one   ",
			expected: &QueryAnnotation{Name: "GetUser", Type: QueryTypeOne},
		},
		{
			name:     "annotation with semicolon",
			line:     "-- name: GetUser :one;",
			expected: &QueryAnnotation{Name: "GetUser", Type: QueryTypeOne},
		},
		{
			name:     "many type",
			line:     "-- name: ListUsers :many",
			expected: &QueryAnnotation{Name: "ListUsers", Type: QueryTypeMany},
		},
		{
			name:     "exec type",
			line:     "-- name: CreateUser :exec",
			expected: &QueryAnnotation{Name: "CreateUser", Type: QueryTypeExec},
		},
		{
			name:     "paginated type",
			line:     "-- name: GetUsersPaginated :paginated",
			expected: &QueryAnnotation{Name: "GetUsersPaginated", Type: QueryTypePaginated},
		},
		{
			name:     "underscore in name",
			line:     "-- name: get_user_by_email :one",
			expected: &QueryAnnotation{Name: "get_user_by_email", Type: QueryTypeOne},
		},
		{
			name:     "invalid format",
			line:     "-- name GetUser :one",
			expected: nil,
		},
		{
			name:     "invalid type",
			line:     "-- name: GetUser :invalid",
			expected: nil,
		},
		{
			name:     "regular comment",
			line:     "-- This is a regular comment",
			expected: nil,
		},
		{
			name:     "empty line",
			line:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseAnnotation(tt.line)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected %+v, got nil", tt.expected)
				return
			}

			if result.Name != tt.expected.Name {
				t.Errorf("Expected name %s, got %s", tt.expected.Name, result.Name)
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Expected type %s, got %s", tt.expected.Type, result.Type)
			}
		})
	}
}

func TestQueryParser_ParseQueryType(t *testing.T) {
	parser := NewQueryParser("")

	tests := []struct {
		name     string
		input    string
		expected QueryType
		hasError bool
	}{
		{"one", "one", QueryTypeOne, false},
		{"many", "many", QueryTypeMany, false},
		{"exec", "exec", QueryTypeExec, false},
		{"paginated", "paginated", QueryTypePaginated, false},
		{"ONE uppercase", "ONE", QueryTypeOne, false},
		{"Many mixed case", "Many", QueryTypeMany, false},
		{"invalid type", "invalid", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseQueryType(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQueryParser_ValidateQuery(t *testing.T) {
	parser := NewQueryParser("")

	tests := []struct {
		name     string
		query    Query
		hasError bool
	}{
		{
			name: "valid select one query",
			query: Query{
				Name: "GetUser",
				Type: QueryTypeOne,
				SQL:  "SELECT id, name FROM users WHERE id = $1",
			},
			hasError: false,
		},
		{
			name: "valid select many query",
			query: Query{
				Name: "ListUsers",
				Type: QueryTypeMany,
				SQL:  "SELECT id, name FROM users ORDER BY name",
			},
			hasError: false,
		},
		{
			name: "valid exec query",
			query: Query{
				Name: "CreateUser",
				Type: QueryTypeExec,
				SQL:  "INSERT INTO users (name) VALUES ($1)",
			},
			hasError: false,
		},
		{
			name: "valid paginated query",
			query: Query{
				Name: "GetUsersPaginated",
				Type: QueryTypePaginated,
				SQL:  "SELECT id, name FROM users ORDER BY id LIMIT $1",
			},
			hasError: false,
		},
		{
			name: "valid CTE query",
			query: Query{
				Name: "GetUsersWithCTE",
				Type: QueryTypeMany,
				SQL:  "WITH active_users AS (SELECT id FROM users WHERE active = true) SELECT * FROM active_users",
			},
			hasError: false,
		},
		{
			name: "empty name",
			query: Query{
				Name: "",
				Type: QueryTypeOne,
				SQL:  "SELECT id FROM users",
			},
			hasError: true,
		},
		{
			name: "empty SQL",
			query: Query{
				Name: "GetUser",
				Type: QueryTypeOne,
				SQL:  "",
			},
			hasError: true,
		},
		{
			name: "empty type",
			query: Query{
				Name: "GetUser",
				Type: "",
				SQL:  "SELECT id FROM users",
			},
			hasError: true,
		},
		{
			name: "invalid Go identifier",
			query: Query{
				Name: "123GetUser",
				Type: QueryTypeOne,
				SQL:  "SELECT id FROM users",
			},
			hasError: true,
		},
		{
			name: "select with exec type",
			query: Query{
				Name: "GetUser",
				Type: QueryTypeExec,
				SQL:  "SELECT id FROM users",
			},
			hasError: true,
		},
		{
			name: "CTE with exec type",
			query: Query{
				Name: "GetUser",
				Type: QueryTypeExec,
				SQL:  "WITH active_users AS (SELECT id FROM users WHERE active = true) SELECT * FROM active_users",
			},
			hasError: true,
		},
		{
			name: "insert with one type",
			query: Query{
				Name: "CreateUser",
				Type: QueryTypeOne,
				SQL:  "INSERT INTO users (name) VALUES ($1)",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateQuery(tt.query)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestQueryParser_IsValidGoIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid identifier", "GetUser", true},
		{"underscore prefix", "_GetUser", true},
		{"with numbers", "GetUser123", true},
		{"with underscores", "get_user_by_email", true},
		{"single letter", "a", true},
		{"single underscore", "_", true},
		{"empty string", "", false},
		{"starts with number", "123GetUser", false},
		{"with spaces", "Get User", false},
		{"with hyphens", "get-user", false},
		{"with dots", "get.user", false},
		{"with special chars", "get@user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGoIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("isValidGoIdentifier(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQueryParser_FindSQLFiles(t *testing.T) {
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

func TestQueryParser_ParseFile_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectCount int
	}{
		{
			name:        "empty file",
			content:     "",
			expectError: false,
			expectCount: 0,
		},
		{
			name: "only comments",
			content: `-- This is a comment
-- Another comment`,
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "annotation without SQL",
			content:     "-- name: GetUser :one",
			expectError: true,
			expectCount: 0,
		},
		{
			name: "multiple queries",
			content: `-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users;

-- name: CreateUser :exec
INSERT INTO users (name) VALUES ($1);`,
			expectError: false,
			expectCount: 3,
		},
		{
			name: "query with complex SQL",
			content: `-- name: GetUserWithPosts :many
SELECT 
    u.id,
    u.name,
    p.title,
    p.content
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.active = true
ORDER BY u.name, p.created_at DESC;`,
			expectError: false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filename := filepath.Join(tempDir, "test.sql")
			err := os.WriteFile(filename, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse file
			parser := NewQueryParser("")
			queries, err := parser.parseFile(filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(queries) != tt.expectCount {
				t.Errorf("Expected %d queries, got %d", tt.expectCount, len(queries))
			}

			// Clean up
			os.Remove(filename)
		})
	}
}

func TestQueryParser_ErrorHandling(t *testing.T) {
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
		parser := NewQueryParser("")
		_, err := parser.ParseQueries()
		if err == nil {
			t.Errorf("Expected error for empty queries directory")
		}
	})
}
