package gen

import (
	"os"
	"path/filepath"
	"testing"
)

// TestQueryParser_Integration tests the parser with real-world SQL files
func TestQueryParser_Integration(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create realistic SQL files
	usersSQL := `-- User management queries

-- name: GetUserByID :one
SELECT id, name, email, created_at, updated_at
FROM users 
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, name, email, created_at, updated_at
FROM users 
WHERE email = $1 AND active = true;

-- name: ListActiveUsers :many
SELECT id, name, email, created_at
FROM users 
WHERE active = true
ORDER BY name ASC;

-- name: GetUsersWithPosts :many
SELECT 
    u.id,
    u.name,
    u.email,
    COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.active = true
GROUP BY u.id, u.name, u.email
ORDER BY post_count DESC;

-- name: GetUsersPaginated :paginated
SELECT id, name, email, created_at
FROM users 
WHERE active = true
  AND ($1::uuid IS NULL OR id > $1)
ORDER BY id ASC
LIMIT $2;

-- name: CreateUser :exec
INSERT INTO users (name, email, active)
VALUES ($1, $2, true);

-- name: UpdateUserEmail :exec
UPDATE users 
SET email = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;`

	postsSQL := `-- Post management queries

-- name: GetPostByID :one
SELECT id, user_id, title, content, published, created_at, updated_at
FROM posts 
WHERE id = $1;

-- name: GetPostsByUser :many
SELECT id, title, content, published, created_at
FROM posts 
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetPublishedPosts :many
SELECT 
    p.id,
    p.title,
    p.content,
    p.created_at,
    u.name as author_name
FROM posts p
JOIN users u ON p.user_id = u.id
WHERE p.published = true
ORDER BY p.created_at DESC;

-- name: GetPostsPaginated :paginated
SELECT id, title, content, published, created_at
FROM posts 
WHERE published = true
  AND ($1::uuid IS NULL OR id > $1)
ORDER BY id ASC
LIMIT $2;

-- name: CreatePost :exec
INSERT INTO posts (user_id, title, content, published)
VALUES ($1, $2, $3, $4);

-- name: UpdatePost :exec
UPDATE posts 
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;`

	// Write test files
	err := os.WriteFile(filepath.Join(tempDir, "users.sql"), []byte(usersSQL), 0644)
	if err != nil {
		t.Fatalf("Failed to write users.sql: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "posts.sql"), []byte(postsSQL), 0644)
	if err != nil {
		t.Fatalf("Failed to write posts.sql: %v", err)
	}

	// Parse queries
	parser := NewQueryParser(tempDir)
	queries, err := parser.ParseQueries()
	if err != nil {
		t.Fatalf("ParseQueries() failed: %v", err)
	}

	// Verify we found the expected number of queries
	expectedCount := 15 // 8 from users.sql + 7 from posts.sql
	if len(queries) != expectedCount {
		t.Errorf("Expected %d queries, got %d", expectedCount, len(queries))
		// Debug: print all found queries
		for i, q := range queries {
			t.Logf("Query %d: %s (%s)", i+1, q.Name, q.Type)
		}
	}

	// Verify all queries are valid
	for _, query := range queries {
		if err := parser.ValidateQuery(query); err != nil {
			t.Errorf("Query %s validation failed: %v", query.Name, err)
		}
	}

	// Verify specific queries exist with correct types
	expectedQueries := map[string]QueryType{
		"GetUserByID":       QueryTypeOne,
		"GetUserByEmail":    QueryTypeOne,
		"ListActiveUsers":   QueryTypeMany,
		"GetUsersWithPosts": QueryTypeMany,
		"GetUsersPaginated": QueryTypePaginated,
		"CreateUser":        QueryTypeExec,
		"UpdateUserEmail":   QueryTypeExec,
		"DeleteUser":        QueryTypeExec,
		"GetPostByID":       QueryTypeOne,
		"GetPostsByUser":    QueryTypeMany,
		"GetPublishedPosts": QueryTypeMany,
		"GetPostsPaginated": QueryTypePaginated,
		"CreatePost":        QueryTypeExec,
		"UpdatePost":        QueryTypeExec,
		"DeletePost":        QueryTypeExec,
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

	// Test specific query content
	getUserByID := queryMap["GetUserByID"]
	if getUserByID.Type != QueryTypeOne {
		t.Errorf("GetUserByID should be :one type")
	}

	getUsersPaginated := queryMap["GetUsersPaginated"]
	if getUsersPaginated.Type != QueryTypePaginated {
		t.Errorf("GetUsersPaginated should be :paginated type")
	}

	createUser := queryMap["CreateUser"]
	if createUser.Type != QueryTypeExec {
		t.Errorf("CreateUser should be :exec type")
	}

	t.Logf("Successfully parsed %d queries from %d files", len(queries), 2)
}

// TestQueryParser_RealWorldScenarios tests various real-world edge cases
func TestQueryParser_RealWorldScenarios(t *testing.T) {
	tempDir := t.TempDir()

	// Test complex query with multiple JOINs and subqueries
	complexSQL := `-- Complex analytics queries

-- name: GetUserAnalytics :many
SELECT 
    u.id,
    u.name,
    u.email,
    COUNT(DISTINCT p.id) as post_count,
    COUNT(DISTINCT c.id) as comment_count,
    AVG(p.created_at) as avg_post_date,
    (
        SELECT COUNT(*) 
        FROM posts p2 
        WHERE p2.user_id = u.id 
          AND p2.published = true
    ) as published_posts
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
LEFT JOIN comments c ON p.id = c.post_id
WHERE u.active = true
  AND u.created_at > $1
GROUP BY u.id, u.name, u.email
HAVING COUNT(DISTINCT p.id) > $2
ORDER BY post_count DESC, comment_count DESC
LIMIT $3;

-- name: BulkUpdateUserStatus :exec
UPDATE users 
SET active = $2, updated_at = NOW()
WHERE id = ANY($1::uuid[]);

-- name: GetTopPostsWithMetrics :paginated
WITH post_metrics AS (
    SELECT 
        p.id,
        p.title,
        p.created_at,
        COUNT(c.id) as comment_count,
        COUNT(DISTINCT c.user_id) as unique_commenters
    FROM posts p
    LEFT JOIN comments c ON p.id = c.post_id
    WHERE p.published = true
    GROUP BY p.id, p.title, p.created_at
)
SELECT 
    pm.id,
    pm.title,
    pm.created_at,
    pm.comment_count,
    pm.unique_commenters,
    (pm.comment_count * 2 + pm.unique_commenters * 3) as engagement_score
FROM post_metrics pm
WHERE ($1::uuid IS NULL OR pm.id > $1)
ORDER BY engagement_score DESC, pm.id ASC
LIMIT $2;`

	err := os.WriteFile(filepath.Join(tempDir, "analytics.sql"), []byte(complexSQL), 0644)
	if err != nil {
		t.Fatalf("Failed to write analytics.sql: %v", err)
	}

	// Parse queries
	parser := NewQueryParser(tempDir)
	queries, err := parser.ParseQueries()
	if err != nil {
		t.Fatalf("ParseQueries() failed: %v", err)
	}

	// Should find 3 queries
	if len(queries) != 3 {
		t.Errorf("Expected 3 queries, got %d", len(queries))
	}

	// Verify all queries are valid
	for _, query := range queries {
		if err := parser.ValidateQuery(query); err != nil {
			t.Errorf("Query %s validation failed: %v", query.Name, err)
		}
	}

	// Verify query types
	expectedTypes := map[string]QueryType{
		"GetUserAnalytics":       QueryTypeMany,
		"BulkUpdateUserStatus":   QueryTypeExec,
		"GetTopPostsWithMetrics": QueryTypePaginated,
	}

	queryMap := make(map[string]Query)
	for _, query := range queries {
		queryMap[query.Name] = query
	}

	for name, expectedType := range expectedTypes {
		query, exists := queryMap[name]
		if !exists {
			t.Errorf("Query %s not found", name)
			continue
		}

		if query.Type != expectedType {
			t.Errorf("Query %s: expected type %s, got %s", name, expectedType, query.Type)
		}
	}

	t.Logf("Successfully parsed complex queries with CTEs, subqueries, and array parameters")
}
