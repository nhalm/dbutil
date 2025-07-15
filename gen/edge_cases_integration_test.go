package gen

import (
	"context"
	"testing"
)

func TestEdgeCases_DatabaseConnectionErrors(t *testing.T) {
	// Test error handling when database connection fails
	introspector := NewIntrospector(nil, "public")
	ctx := context.Background()

	// These should handle nil database gracefully or panic
	t.Run("nil_database_GetTables", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected with nil database connection
				t.Logf("Expected panic with nil database: %v", r)
			}
		}()

		_, err := introspector.GetTables(ctx)
		if err == nil {
			t.Error("Expected error with nil database connection")
		}
	})

	t.Run("nil_database_getTableNames", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected with nil database connection
				t.Logf("Expected panic with nil database: %v", r)
			}
		}()

		_, err := introspector.getTableNames(ctx)
		if err == nil {
			t.Error("Expected error with nil database connection")
		}
	})

	t.Run("nil_database_getTableDetails", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected with nil database connection
				t.Logf("Expected panic with nil database: %v", r)
			}
		}()

		_, err := introspector.getTableDetails(ctx, "test_table")
		if err == nil {
			t.Error("Expected error with nil database connection")
		}
	})
}
