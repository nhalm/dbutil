package dbutil

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

// HasID interface constraint for pagination - all entities must have a UUID ID
type HasID interface {
	GetID() uuid.UUID
}

// PaginationParams holds parameters for cursor-based pagination
type PaginationParams struct {
	// Cursor is the base64-encoded UUID to start pagination from
	// If empty, starts from the beginning
	Cursor string `json:"cursor,omitempty"`

	// Limit is the maximum number of items to return
	// Must be between 1 and 100, defaults to 20
	Limit int `json:"limit,omitempty"`
}

// PaginationResult holds the result of a paginated query
type PaginationResult[T HasID] struct {
	// Items is the list of items returned
	Items []T `json:"items"`

	// HasMore indicates if there are more items available
	HasMore bool `json:"has_more"`

	// NextCursor is the cursor for the next page
	// Only set if HasMore is true
	NextCursor string `json:"next_cursor,omitempty"`

	// Total is the total count of items (optional, expensive to calculate)
	Total *int `json:"total,omitempty"`
}

// QueryFunc is a function that executes a paginated query
// It receives a cursor (nil for first page) and limit, and returns items
type QueryFunc[T HasID] func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]T, error)

// Paginate executes a paginated query using the provided query function
// This is the main pagination utility that handles cursor management internally
func Paginate[T HasID](ctx context.Context, params PaginationParams, queryFunc QueryFunc[T]) (*PaginationResult[T], error) {
	// Validate and set default limit
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Parse cursor if provided
	var cursor *uuid.UUID
	if params.Cursor != "" {
		cursorUUID, err := DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}
		cursor = &cursorUUID
	}

	// Execute query with limit + 1 to check if there are more items
	items, err := queryFunc(ctx, cursor, int32(limit+1))
	if err != nil {
		return nil, fmt.Errorf("pagination query failed: %w", err)
	}

	// Check if there are more items
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit] // Remove the extra item
	}

	// Generate next cursor if there are more items
	var nextCursor string
	if hasMore && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor = encodeCursor(lastItem.GetID())
	}

	return &PaginationResult[T]{
		Items:      items,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// encodeCursor encodes a UUID as a base64 cursor
func encodeCursor(id uuid.UUID) string {
	return base64.URLEncoding.EncodeToString(id[:])
}

// DecodeCursor decodes a base64 cursor back to a UUID
func DecodeCursor(cursor string) (uuid.UUID, error) {
	if cursor == "" {
		return uuid.Nil, fmt.Errorf("empty cursor")
	}

	cursorBytes, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	if len(cursorBytes) != 16 {
		return uuid.Nil, fmt.Errorf("invalid cursor length: expected 16 bytes, got %d", len(cursorBytes))
	}

	var id uuid.UUID
	copy(id[:], cursorBytes)
	return id, nil
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(params PaginationParams) error {
	if params.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if params.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	if params.Cursor != "" {
		_, err := DecodeCursor(params.Cursor)
		if err != nil {
			return fmt.Errorf("invalid cursor: %w", err)
		}
	}

	return nil
}
