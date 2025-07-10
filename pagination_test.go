package dbutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// TestEntity implements HasID interface for testing
type TestEntity struct {
	ID   uuid.UUID
	Name string
}

func (e TestEntity) GetID() uuid.UUID {
	return e.ID
}

func TestPaginationParams(t *testing.T) {
	tests := []struct {
		name      string
		params    PaginationParams
		wantError bool
	}{
		{
			name:      "valid params with defaults",
			params:    PaginationParams{},
			wantError: false,
		},
		{
			name: "valid params with limit",
			params: PaginationParams{
				Limit: 50,
			},
			wantError: false,
		},
		{
			name: "valid params with cursor",
			params: PaginationParams{
				Cursor: encodeCursor(uuid.New()),
				Limit:  10,
			},
			wantError: false,
		},
		{
			name: "negative limit",
			params: PaginationParams{
				Limit: -1,
			},
			wantError: true,
		},
		{
			name: "limit too high",
			params: PaginationParams{
				Limit: 101,
			},
			wantError: true,
		},
		{
			name: "invalid cursor",
			params: PaginationParams{
				Cursor: "invalid-cursor",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePaginationParams(tt.params)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCursorEncoding(t *testing.T) {
	// Test with a real UUID
	originalID := uuid.New()

	// Encode the UUID
	cursor := encodeCursor(originalID)
	if cursor == "" {
		t.Error("expected non-empty cursor")
	}

	// Decode the cursor
	decodedID, err := DecodeCursor(cursor)
	if err != nil {
		t.Errorf("failed to decode cursor: %v", err)
	}

	// Verify they match
	if originalID != decodedID {
		t.Errorf("UUID mismatch: original=%s, decoded=%s", originalID, decodedID)
	}
}

func TestDecodeCursor(t *testing.T) {
	tests := []struct {
		name      string
		cursor    string
		wantError bool
	}{
		{
			name:      "empty cursor",
			cursor:    "",
			wantError: true,
		},
		{
			name:      "invalid base64",
			cursor:    "invalid-base64!",
			wantError: true,
		},
		{
			name:      "valid cursor",
			cursor:    encodeCursor(uuid.New()),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCursor(tt.cursor)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPaginate(t *testing.T) {
	ctx := context.Background()

	// Create test data
	testEntities := []TestEntity{
		{ID: uuid.New(), Name: "Entity 1"},
		{ID: uuid.New(), Name: "Entity 2"},
		{ID: uuid.New(), Name: "Entity 3"},
		{ID: uuid.New(), Name: "Entity 4"},
		{ID: uuid.New(), Name: "Entity 5"},
	}

	// Mock query function
	queryFunc := func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]TestEntity, error) {
		startIndex := 0
		if cursor != nil {
			// Find the starting position based on cursor
			for i, entity := range testEntities {
				if entity.ID == *cursor {
					startIndex = i + 1
					break
				}
			}
		}

		endIndex := startIndex + int(limit)
		if endIndex > len(testEntities) {
			endIndex = len(testEntities)
		}

		if startIndex >= len(testEntities) {
			return []TestEntity{}, nil
		}

		return testEntities[startIndex:endIndex], nil
	}

	t.Run("first page", func(t *testing.T) {
		params := PaginationParams{Limit: 2}
		result, err := Paginate(ctx, params, queryFunc)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(result.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(result.Items))
		}

		if !result.HasMore {
			t.Error("expected HasMore to be true")
		}

		if result.NextCursor == "" {
			t.Error("expected NextCursor to be set")
		}
	})

	t.Run("second page", func(t *testing.T) {
		// First get the first page to get the cursor
		firstPageParams := PaginationParams{Limit: 2}
		firstResult, err := Paginate(ctx, firstPageParams, queryFunc)
		if err != nil {
			t.Fatalf("failed to get first page: %v", err)
		}

		// Now get the second page
		secondPageParams := PaginationParams{
			Cursor: firstResult.NextCursor,
			Limit:  2,
		}
		secondResult, err := Paginate(ctx, secondPageParams, queryFunc)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(secondResult.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(secondResult.Items))
		}

		if !secondResult.HasMore {
			t.Error("expected HasMore to be true")
		}
	})

	t.Run("last page", func(t *testing.T) {
		// Get a page that should contain the last item
		params := PaginationParams{
			Limit: 10, // More than we have
		}
		result, err := Paginate(ctx, params, queryFunc)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(result.Items) != 5 {
			t.Errorf("expected 5 items, got %d", len(result.Items))
		}

		if result.HasMore {
			t.Error("expected HasMore to be false")
		}

		if result.NextCursor != "" {
			t.Error("expected NextCursor to be empty")
		}
	})

	t.Run("default limit", func(t *testing.T) {
		params := PaginationParams{} // No limit specified
		result, err := Paginate(ctx, params, queryFunc)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should return all 5 items since default limit is 20
		if len(result.Items) != 5 {
			t.Errorf("expected 5 items, got %d", len(result.Items))
		}
	})

	t.Run("limit exceeds maximum", func(t *testing.T) {
		params := PaginationParams{Limit: 150} // Exceeds max of 100
		result, err := Paginate(ctx, params, queryFunc)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should still work, but limit should be capped at 100
		if len(result.Items) != 5 {
			t.Errorf("expected 5 items, got %d", len(result.Items))
		}
	})
}

func TestPaginateWithError(t *testing.T) {
	ctx := context.Background()

	// Query function that always returns an error
	errorQueryFunc := func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]TestEntity, error) {
		return nil, fmt.Errorf("database error")
	}

	params := PaginationParams{Limit: 10}
	result, err := Paginate(ctx, params, errorQueryFunc)

	if err == nil {
		t.Error("expected error but got none")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}
