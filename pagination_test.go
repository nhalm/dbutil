package dbutil

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestEncodeCursor(t *testing.T) {
	// Test with valid UUID v7
	id := NewUUID()
	cursor := EncodeCursor(id)

	if cursor == "" {
		t.Errorf("Expected non-empty cursor for valid UUID")
	}

	// Test with nil UUID
	nilCursor := EncodeCursor(uuid.Nil)
	if nilCursor != "" {
		t.Errorf("Expected empty cursor for nil UUID, got %s", nilCursor)
	}
}

func TestDecodeCursor(t *testing.T) {
	// Test with valid cursor
	originalID := NewUUID()
	cursor := EncodeCursor(originalID)

	decodedID, err := DecodeCursor(cursor)
	if err != nil {
		t.Errorf("Unexpected error decoding cursor: %v", err)
	}

	if decodedID != originalID {
		t.Errorf("Expected %v, got %v", originalID, decodedID)
	}

	// Test with empty cursor
	decodedNil, err := DecodeCursor("")
	if err != nil {
		t.Errorf("Unexpected error decoding empty cursor: %v", err)
	}

	if decodedNil != uuid.Nil {
		t.Errorf("Expected nil UUID for empty cursor, got %v", decodedNil)
	}

	// Test with invalid cursor
	_, err = DecodeCursor("invalid-cursor")
	if err == nil {
		t.Errorf("Expected error for invalid cursor")
	}

	var paginationErr *PaginationError
	if !errors.As(err, &paginationErr) {
		t.Errorf("Expected PaginationError, got %T", err)
	}

	// Test with cursor of wrong length
	_, err = DecodeCursor("dGVzdA") // "test" in base64
	if err == nil {
		t.Errorf("Expected error for cursor with wrong length")
	}
}

func TestValidatePaginationParams(t *testing.T) {
	// Test valid parameters
	validParams := PaginationParams{
		Cursor: nil,
		Limit:  20,
	}

	err := ValidatePaginationParams(validParams)
	if err != nil {
		t.Errorf("Unexpected error for valid params: %v", err)
	}

	// Test valid parameters with cursor
	cursor := EncodeCursor(NewUUID())
	validParamsWithCursor := PaginationParams{
		Cursor: &cursor,
		Limit:  50,
	}

	err = ValidatePaginationParams(validParamsWithCursor)
	if err != nil {
		t.Errorf("Unexpected error for valid params with cursor: %v", err)
	}

	// Test zero limit
	zeroLimitParams := PaginationParams{
		Cursor: nil,
		Limit:  0,
	}

	err = ValidatePaginationParams(zeroLimitParams)
	if err == nil {
		t.Errorf("Expected error for zero limit")
	}

	// Test negative limit
	negativeLimitParams := PaginationParams{
		Cursor: nil,
		Limit:  -1,
	}

	err = ValidatePaginationParams(negativeLimitParams)
	if err == nil {
		t.Errorf("Expected error for negative limit")
	}

	// Test with large limit (should be allowed now)
	largeLimitParams := PaginationParams{
		Cursor: nil,
		Limit:  10000,
	}

	err = ValidatePaginationParams(largeLimitParams)
	if err != nil {
		t.Errorf("Unexpected error for large limit: %v", err)
	}

	// Test invalid cursor
	invalidCursor := "invalid-cursor"
	invalidCursorParams := PaginationParams{
		Cursor: &invalidCursor,
		Limit:  20,
	}

	err = ValidatePaginationParams(invalidCursorParams)
	if err == nil {
		t.Errorf("Expected error for invalid cursor")
	}
}

func TestPaginationError(t *testing.T) {
	// Test error without wrapped error
	err := &PaginationError{
		Operation: "test",
		Reason:    "test reason",
		Err:       nil,
	}

	expected := "pagination test failed: test reason"
	if err.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, err.Error())
	}

	// Test error with wrapped error
	wrappedErr := errors.New("wrapped error")
	errWithWrapped := &PaginationError{
		Operation: "test",
		Reason:    "test reason",
		Err:       wrappedErr,
	}

	expectedWithWrapped := "pagination test failed: test reason: wrapped error"
	if errWithWrapped.Error() != expectedWithWrapped {
		t.Errorf("Expected %s, got %s", expectedWithWrapped, errWithWrapped.Error())
	}

	// Test unwrap
	if errWithWrapped.Unwrap() != wrappedErr {
		t.Errorf("Expected wrapped error, got %v", errWithWrapped.Unwrap())
	}
}

func TestNewPaginationError(t *testing.T) {
	wrappedErr := errors.New("wrapped")
	err := NewPaginationError("test", "reason", wrappedErr)

	var paginationErr *PaginationError
	if !errors.As(err, &paginationErr) {
		t.Errorf("Expected PaginationError, got %T", err)
	}

	if paginationErr.Operation != "test" {
		t.Errorf("Expected operation 'test', got %s", paginationErr.Operation)
	}

	if paginationErr.Reason != "reason" {
		t.Errorf("Expected reason 'reason', got %s", paginationErr.Reason)
	}

	if paginationErr.Err != wrappedErr {
		t.Errorf("Expected wrapped error, got %v", paginationErr.Err)
	}
}

func TestIsPaginationError(t *testing.T) {
	// Test with pagination error
	paginationErr := &PaginationError{
		Operation: "test",
		Reason:    "test",
		Err:       nil,
	}

	if !IsPaginationError(paginationErr) {
		t.Errorf("Expected true for PaginationError")
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	if IsPaginationError(regularErr) {
		t.Errorf("Expected false for regular error")
	}

	// Test with nil
	if IsPaginationError(nil) {
		t.Errorf("Expected false for nil error")
	}
}

func TestPaginationResult(t *testing.T) {
	// Test creating a pagination result
	items := []string{"item1", "item2", "item3"}
	nextCursor := EncodeCursor(NewUUID())

	result := PaginationResult[string]{
		Items:      items,
		NextCursor: &nextCursor,
		HasMore:    true,
	}

	if len(result.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result.Items))
	}

	if result.NextCursor == nil {
		t.Errorf("Expected non-nil NextCursor")
	}

	if !result.HasMore {
		t.Errorf("Expected HasMore to be true")
	}

	// Test result with no more pages
	resultNoMore := PaginationResult[string]{
		Items:      items,
		NextCursor: nil,
		HasMore:    false,
	}

	if resultNoMore.NextCursor != nil {
		t.Errorf("Expected nil NextCursor")
	}

	if resultNoMore.HasMore {
		t.Errorf("Expected HasMore to be false")
	}
}

type TestItem struct {
	ID   uuid.UUID
	Name string
}

func (t TestItem) GetID() uuid.UUID {
	return t.ID
}

func TestPaginate(t *testing.T) {
	ctx := context.Background()

	// Create test data
	testItems := []TestItem{
		{ID: NewUUID(), Name: "Item 1"},
		{ID: NewUUID(), Name: "Item 2"},
		{ID: NewUUID(), Name: "Item 3"},
		{ID: NewUUID(), Name: "Item 4"},
		{ID: NewUUID(), Name: "Item 5"},
	}

	// Test first page (no cursor)
	params := PaginationParams{
		Cursor: nil,
		Limit:  3,
	}

	result, err := Paginate(ctx, params, func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]TestItem, error) {
		var startIndex int

		if cursor == nil {
			startIndex = 0
		} else {
			// Find the starting point after the cursor
			startIndex = -1
			for i, item := range testItems {
				if item.ID == *cursor {
					startIndex = i + 1
					break
				}
			}
			if startIndex == -1 || startIndex >= len(testItems) {
				return []TestItem{}, nil
			}
		}

		// Get items starting from startIndex, up to limit
		var items []TestItem
		for i := startIndex; i < len(testItems) && len(items) < int(limit); i++ {
			items = append(items, testItems[i])
		}

		return items, nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result.Items))
	}

	if !result.HasMore {
		t.Errorf("Expected HasMore to be true")
	}

	if result.NextCursor == nil {
		t.Errorf("Expected non-nil NextCursor")
	}

	// Test next page with cursor
	nextParams := PaginationParams{
		Cursor: result.NextCursor,
		Limit:  3,
	}

	nextResult, err := Paginate(ctx, nextParams, func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]TestItem, error) {
		// Same query logic as above
		var startIndex int

		if cursor == nil {
			startIndex = 0
		} else {
			startIndex = -1
			for i, item := range testItems {
				if item.ID == *cursor {
					startIndex = i + 1
					break
				}
			}
			if startIndex == -1 || startIndex >= len(testItems) {
				return []TestItem{}, nil
			}
		}

		var items []TestItem
		for i := startIndex; i < len(testItems) && len(items) < int(limit); i++ {
			items = append(items, testItems[i])
		}

		return items, nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(nextResult.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(nextResult.Items))
	}

	if nextResult.HasMore {
		t.Errorf("Expected HasMore to be false")
	}

	if nextResult.NextCursor != nil {
		t.Errorf("Expected nil NextCursor")
	}
}

func TestPaginateWithError(t *testing.T) {
	ctx := context.Background()

	params := PaginationParams{
		Cursor: nil,
		Limit:  10,
	}

	// Test query function error
	_, err := Paginate(ctx, params, func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]TestItem, error) {
		return nil, errors.New("query error")
	})

	if err == nil {
		t.Errorf("Expected error from query function")
	}

	var paginationErr *PaginationError
	if !errors.As(err, &paginationErr) {
		t.Errorf("Expected PaginationError, got %T", err)
	}
}

func TestPaginateInvalidParams(t *testing.T) {
	ctx := context.Background()

	params := PaginationParams{
		Cursor: nil,
		Limit:  0, // Invalid
	}

	_, err := Paginate(ctx, params, func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]TestItem, error) {
		return []TestItem{}, nil
	})

	if err == nil {
		t.Errorf("Expected error for invalid parameters")
	}
}
