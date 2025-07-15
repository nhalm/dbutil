package gen

import (
	"context"
	"testing"
)

// Integration tests - require database connection
func TestQueryAnalyzer_Integration(t *testing.T) {
	// Get test database connection
	pool := getTestDB(t)
	defer pool.Close()

	analyzer := NewQueryAnalyzer(pool)
	ctx := context.Background()

	t.Run("AnalyzeQuery", func(t *testing.T) {
		tests := []struct {
			name           string
			query          Query
			expectedParams int
			expectedCols   int
			shouldError    bool
		}{
			{
				name: "simple SELECT query",
				query: Query{
					Name: "GetUserByID",
					SQL:  "SELECT id, name, email FROM users WHERE id = $1",
					Type: QueryTypeOne,
				},
				expectedParams: 1,
				expectedCols:   3,
				shouldError:    false,
			},
			{
				name: "SELECT query with multiple parameters",
				query: Query{
					Name: "GetUsersByNameAndEmail",
					SQL:  "SELECT id, name, email FROM users WHERE name = $1 AND email = $2",
					Type: QueryTypeMany,
				},
				expectedParams: 2,
				expectedCols:   3,
				shouldError:    false,
			},
			{
				name: "SELECT query with no parameters",
				query: Query{
					Name: "GetAllUsers",
					SQL:  "SELECT id, name, email FROM users",
					Type: QueryTypeMany,
				},
				expectedParams: 0,
				expectedCols:   3,
				shouldError:    false,
			},
			{
				name: "INSERT query",
				query: Query{
					Name: "CreateUser",
					SQL:  "INSERT INTO users (name, email) VALUES ($1, $2)",
					Type: QueryTypeExec,
				},
				expectedParams: 2,
				expectedCols:   0,
				shouldError:    false,
			},
			{
				name: "UPDATE query",
				query: Query{
					Name: "UpdateUser",
					SQL:  "UPDATE users SET name = $1 WHERE id = $2",
					Type: QueryTypeExec,
				},
				expectedParams: 2,
				expectedCols:   0,
				shouldError:    false,
			},
			{
				name: "DELETE query",
				query: Query{
					Name: "DeleteUser",
					SQL:  "DELETE FROM users WHERE id = $1",
					Type: QueryTypeExec,
				},
				expectedParams: 1,
				expectedCols:   0,
				shouldError:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := tt.query
				err := analyzer.AnalyzeQuery(ctx, &query)

				if tt.shouldError {
					if err == nil {
						t.Errorf("Expected error for %s, but got none", tt.name)
					}
					return
				}

				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
					return
				}

				// Check parameters
				if len(query.Parameters) != tt.expectedParams {
					t.Errorf("Expected %d parameters, got %d", tt.expectedParams, len(query.Parameters))
				}

				// Check columns for SELECT queries
				if len(query.Columns) != tt.expectedCols {
					t.Errorf("Expected %d columns, got %d", tt.expectedCols, len(query.Columns))
				}

				// Verify parameter types are set
				for i, param := range query.Parameters {
					if param.Type == "" {
						t.Errorf("Parameter %d type not set", i)
					}
					if param.GoType == "" {
						t.Errorf("Parameter %d GoType not set", i)
					}
				}

				// Verify column types are set for SELECT queries
				for i, col := range query.Columns {
					if col.Type == "" {
						t.Errorf("Column %d type not set", i)
					}
					if col.GoType == "" {
						t.Errorf("Column %d GoType not set", i)
					}
				}
			})
		}
	})

	t.Run("ComplexQueries", func(t *testing.T) {
		tests := []struct {
			name  string
			query Query
		}{
			{
				name: "CTE query",
				query: Query{
					Name: "GetUserHierarchy",
					SQL: `WITH user_posts AS (
						SELECT user_id, COUNT(*) as post_count 
						FROM posts 
						WHERE created_at > $1 
						GROUP BY user_id
					)
					SELECT u.id, u.name, up.post_count
					FROM users u
					JOIN user_posts up ON u.id = up.user_id`,
					Type: QueryTypeMany,
				},
			},
			{
				name: "subquery",
				query: Query{
					Name: "GetUsersWithRecentPosts",
					SQL: `SELECT u.id, u.name
						FROM users u
						WHERE EXISTS (
							SELECT 1 FROM posts p 
							WHERE p.user_id = u.id 
							AND p.created_at > $1
						)`,
					Type: QueryTypeMany,
				},
			},
			{
				name: "window function",
				query: Query{
					Name: "GetUsersWithRanking",
					SQL: `SELECT id, name, email,
						ROW_NUMBER() OVER (ORDER BY created_at DESC) as rank
						FROM users
						WHERE created_at > $1`,
					Type: QueryTypeMany,
				},
			},
			{
				name: "multiple joins",
				query: Query{
					Name: "GetUserPostsWithCategories",
					SQL: `SELECT u.name, p.title, c.name as category
						FROM users u
						JOIN posts p ON u.id = p.user_id
						JOIN post_categories pc ON p.id = pc.post_id
						JOIN categories c ON pc.category_id = c.id
						WHERE u.id = $1`,
					Type: QueryTypeMany,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := tt.query
				err := analyzer.AnalyzeQuery(ctx, &query)
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}

				// Verify basic analysis was performed
				if len(query.Parameters) == 0 && query.SQL != "" {
					// Some queries might not have parameters
				}
			})
		}
	})

	t.Run("TypeMapping", func(t *testing.T) {
		tests := []struct {
			name  string
			query Query
		}{
			{
				name: "basic types",
				query: Query{
					Name: "GetUserDetails",
					SQL:  "SELECT id, name, email, created_at, is_active FROM users WHERE id = $1",
					Type: QueryTypeOne,
				},
			},
			{
				name: "data_types_test table",
				query: Query{
					Name: "GetDataTypesTest",
					SQL:  "SELECT id, text_field, integer_field, boolean_field, timestamp_field, uuid_field FROM data_types_test WHERE id = $1",
					Type: QueryTypeOne,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := tt.query
				err := analyzer.AnalyzeQuery(ctx, &query)
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}

				// Verify column types are mapped
				for _, col := range query.Columns {
					if col.Type == "" {
						t.Errorf("Column %s type not mapped", col.Name)
					}
					if col.GoType == "" {
						t.Errorf("Column %s GoType not mapped", col.Name)
					}
				}
			})
		}
	})

	t.Run("ParameterTypeInference", func(t *testing.T) {
		tests := []struct {
			name  string
			query Query
		}{
			{
				name: "UUID parameter",
				query: Query{
					Name: "GetUserByID",
					SQL:  "SELECT id, name FROM users WHERE id = $1",
					Type: QueryTypeOne,
				},
			},
			{
				name: "multiple parameters",
				query: Query{
					Name: "GetUsersByNameAndEmail",
					SQL:  "SELECT id, name FROM users WHERE name = $1 AND email = $2",
					Type: QueryTypeMany,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := tt.query
				err := analyzer.AnalyzeQuery(ctx, &query)
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}

				// Verify parameter types are inferred
				for _, param := range query.Parameters {
					if param.Type == "" {
						t.Errorf("Parameter %s type not inferred", param.Name)
					}
				}
			})
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		tests := []struct {
			name        string
			query       Query
			expectError bool
		}{
			{
				name: "invalid SQL syntax",
				query: Query{
					Name: "InvalidQuery",
					SQL:  "SELECT FROM users WHERE",
					Type: QueryTypeMany,
				},
				expectError: true,
			},
			{
				name: "missing table",
				query: Query{
					Name: "MissingTable",
					SQL:  "SELECT id FROM nonexistent_table",
					Type: QueryTypeMany,
				},
				expectError: true,
			},
			{
				name: "missing column",
				query: Query{
					Name: "MissingColumn",
					SQL:  "SELECT nonexistent_column FROM users",
					Type: QueryTypeMany,
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := tt.query
				err := analyzer.AnalyzeQuery(ctx, &query)

				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error for %s, but got none", tt.name)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for %s: %v", tt.name, err)
					}
				}
			})
		}
	})
}
