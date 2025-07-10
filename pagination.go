package dbutil

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// PaginationParams holds parameters for cursor-based pagination
type PaginationParams struct {
	// Cursor is an opaque string representing the position to start from.
	// If nil, pagination starts from the beginning.
	Cursor *string
	// Limit is the maximum number of items to return.
	// Must be positive.
	Limit int
}

// PaginationResult holds the result of a paginated query
type PaginationResult[T any] struct {
	// Items contains the paginated results
	Items []T
	// NextCursor is an opaque cursor for the next page.
	// If nil, there are no more pages.
	NextCursor *string
	// HasMore indicates whether there are more items beyond this page
	HasMore bool
}

// PaginationError represents errors that occur during pagination
type PaginationError struct {
	Operation string
	Reason    string
	Err       error
}

func (e *PaginationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("pagination %s failed: %s: %v", e.Operation, e.Reason, e.Err)
	}
	return fmt.Sprintf("pagination %s failed: %s", e.Operation, e.Reason)
}

func (e *PaginationError) Unwrap() error {
	return e.Err
}

// EncodeCursor creates an opaque cursor from a UUID v7
func EncodeCursor(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	// Base64 encode the UUID bytes for an opaque cursor
	return base64.RawURLEncoding.EncodeToString(id[:])
}

// DecodeCursor decodes an opaque cursor back to a UUID v7
func DecodeCursor(cursor string) (uuid.UUID, error) {
	if cursor == "" {
		return uuid.Nil, nil
	}

	// Decode the base64 cursor
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return uuid.Nil, &PaginationError{
			Operation: "decode",
			Reason:    "invalid cursor format",
			Err:       err,
		}
	}

	// Ensure we have exactly 16 bytes for a UUID
	if len(data) != 16 {
		return uuid.Nil, &PaginationError{
			Operation: "decode",
			Reason:    "invalid cursor length",
			Err:       fmt.Errorf("expected 16 bytes, got %d", len(data)),
		}
	}

	// Convert bytes to UUID
	var id uuid.UUID
	copy(id[:], data)

	// Basic validation - check if it looks like a valid UUID
	if id == uuid.Nil {
		return uuid.Nil, &PaginationError{
			Operation: "decode",
			Reason:    "cursor represents nil UUID",
			Err:       nil,
		}
	}

	return id, nil
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(params PaginationParams) error {
	if params.Limit <= 0 {
		return &PaginationError{
			Operation: "validate",
			Reason:    "limit must be positive",
			Err:       nil,
		}
	}

	// If cursor is provided, validate it
	if params.Cursor != nil && *params.Cursor != "" {
		_, err := DecodeCursor(*params.Cursor)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewPaginationError creates a new pagination error
func NewPaginationError(operation, reason string, err error) error {
	return &PaginationError{
		Operation: operation,
		Reason:    reason,
		Err:       err,
	}
}

// IsPaginationError checks if an error is a pagination error
func IsPaginationError(err error) bool {
	var paginationErr *PaginationError
	return errors.As(err, &paginationErr)
}

// Paginate executes a paginated query and returns the complete result
// The queryFunc should handle the cursor logic and return a complete PaginationResult
func Paginate[T any](ctx context.Context, params PaginationParams, queryFunc func(ctx context.Context, cursor *uuid.UUID, limit int) (*PaginationResult[T], error)) (*PaginationResult[T], error) {
	if err := ValidatePaginationParams(params); err != nil {
		return nil, err
	}

	// Decode cursor if provided
	var cursorID *uuid.UUID
	if params.Cursor != nil && *params.Cursor != "" {
		id, err := DecodeCursor(*params.Cursor)
		if err != nil {
			return nil, err
		}
		cursorID = &id
	}

	// Execute the query function
	result, err := queryFunc(ctx, cursorID, params.Limit)
	if err != nil {
		return nil, NewPaginationError("query", "failed to execute query", err)
	}

	return result, nil
}
