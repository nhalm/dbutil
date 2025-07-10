# Database Utilities

[![Go Version](https://img.shields.io/github/go-mod/go-version/nhalm/dbutil)](https://golang.org/doc/devel/release.html)
[![CI Status](https://github.com/nhalm/dbutil/actions/workflows/ci.yml/badge.svg)](https://github.com/nhalm/dbutil/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhalm/dbutil)](https://goreportcard.com/report/github.com/nhalm/dbutil)
[![Release](https://img.shields.io/github/v/release/nhalm/dbutil)](https://github.com/nhalm/dbutil/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go package that provides database connection utilities, pagination, and testing infrastructure for applications using PostgreSQL with pgx and sqlc.

## Overview

This package is designed specifically for **sqlc users** who want:
- **Reusable database utilities** that work with any sqlc-generated queries
- **Optimized testing infrastructure** with shared connections for faster tests
- **Type-safe PostgreSQL operations** with comprehensive pgx type helpers
- **Built-in pagination** with cursor-based pagination for UUID v7 primary keys
- **Structured error handling** with consistent error types
- **Production-ready features** like health checks, metrics, retry logic, and connection hooks

## Package Structure

- **`github.com/nhalm/dbutil`** - Core utilities including pagination
- **`github.com/nhalm/dbutil/connection`** - Database connection management, testing, and pgx helpers

## Installation

```bash
go get github.com/nhalm/dbutil
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/nhalm/dbutil/connection"
    "your-project/internal/repository/sqlc" // Your sqlc-generated package
)

func main() {
    ctx := context.Background()
    
    // Create connection with your sqlc queries
    conn, err := connection.NewConnection(ctx, "", sqlc.New)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Use your queries directly
    queries := conn.Queries()
    users, err := queries.GetAllUsers(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d users", len(users))
}
```

## Pagination

Built-in cursor-based pagination for UUID v7 primary keys:

```go
import "github.com/nhalm/dbutil"

// Your entity must implement HasID interface
type User struct {
    ID    uuid.UUID `json:"id"`
    Name  string    `json:"name"`
    Email string    `json:"email"`
}

func (u User) GetID() uuid.UUID {
    return u.ID
}

// Paginate query results
params := dbutil.PaginationParams{
    Limit:  20,
    Cursor: "", // Empty for first page
}

result, err := dbutil.Paginate(ctx, params, func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]User, error) {
    // Your sqlc query with cursor support
    return queries.GetUsersPaginated(ctx, sqlc.GetUsersPaginatedParams{
        Cursor: cursor,
        Limit:  limit,
    })
})

if err != nil {
    log.Fatal(err)
}

log.Printf("Found %d users, has more: %t", len(result.Items), result.HasMore)
if result.HasMore {
    log.Printf("Next cursor: %s", result.NextCursor)
}
```

## Configuration

### Environment Variables
The package uses these environment variables with sensible defaults:
- `POSTGRES_HOST` (default: "localhost")
- `POSTGRES_PORT` (default: 5432)
- `POSTGRES_USER` (default: "postgres")
- `POSTGRES_PASSWORD` (default: "")
- `POSTGRES_DB` (default: "postgres")
- `POSTGRES_SSLMODE` (default: "disable")
- `TEST_DATABASE_URL` (for integration tests)

### Custom Configuration
```go
import "github.com/nhalm/dbutil/connection"

config := &connection.Config{
    MaxConns:        20,
    MinConns:        5,
    MaxConnLifetime: 1 * time.Hour,
    SearchPath:      "myschema",
}
conn, err := connection.NewConnectionWithConfig(ctx, "", sqlc.New, config)
```

## Key Features

### **Generic Design**
Works with any sqlc-generated queries without coupling to specific packages:
```go
conn, err := connection.NewConnection(ctx, "", myapp.New)
conn, err := connection.NewConnection(ctx, "", yourapp.New)
```

### **Transaction Support**
```go
err = conn.WithTransaction(ctx, func(ctx context.Context, tx *sqlc.Queries) error {
    // All operations run in transaction, automatically rolled back on error
    user, err := tx.CreateUser(ctx, params)
    if err != nil {
        return err
    }
    return tx.CreateUserProfile(ctx, profileParams)
})
```

### **Health Checks & Monitoring**
```go
if conn.IsReady(ctx) {
    log.Println("Database is ready")
}

stats := conn.Stats()
log.Printf("Active connections: %d", stats.TotalConns())

// With metrics and hooks
conn = conn.WithMetrics(myMetricsCollector)
conn = conn.WithHooks(myHooks)
```

### **Read/Write Splitting**
```go
import "github.com/nhalm/dbutil/connection"

rwConn, err := connection.NewReadWriteConnection(ctx, readDSN, writeDSN, sqlc.New)
readQueries := rwConn.ReadQueries()   // Use for SELECT queries
writeQueries := rwConn.WriteQueries() // Use for INSERT/UPDATE/DELETE
```

### **Retry Logic**
```go
retryableConn := conn.WithRetry(nil) // Uses defaults
err = retryableConn.WithRetryableTransaction(ctx, func(ctx context.Context, tx *sqlc.Queries) error {
    return tx.CreateUser(ctx, params)
})
```

## Testing

This package provides optimized testing utilities with shared connections for faster integration tests:

```go
import "github.com/nhalm/dbutil/connection"

func TestUserOperations(t *testing.T) {
    conn := connection.RequireTestDB(t, sqlc.New)     // Shared connection
    connection.CleanupTestData(conn,                  // Clean data between tests
        "DELETE FROM users WHERE email LIKE 'test_%'",
    )
    
    // Run your test logic
    queries := conn.Queries()
    user, err := queries.CreateUser(ctx, params)
    // ... test assertions
}
```

### Test Database Setup
```bash
# Start test database
docker run --name test-postgres -e POSTGRES_PASSWORD=testpass -p 5433:5432 -d postgres:15

# Set environment variable
export TEST_DATABASE_URL="postgres://postgres:testpass@localhost:5433/postgres?sslmode=disable"

# Run integration tests
go test ./...
```

### Test Utilities
- **`connection.RequireTestDB(t, sqlc.New)`** - Returns shared test connection, skips if no database
- **`connection.CleanupTestData(conn, "DELETE ...")`** - Cleans test data between tests
- **`connection.GetTestConnection(sqlc.New)`** - Returns connection or nil if unavailable

## Type Helpers

Comprehensive pgx type conversion utilities:

```go
import "github.com/nhalm/dbutil/connection"

// String conversions
pgxText := connection.ToPgxText(&myString)
stringPtr := connection.FromPgxText(pgxText)

// Numeric conversions
pgxNum := connection.ToPgxNumericFromFloat64Ptr(&myFloat)
floatPtr := connection.FromPgxNumericPtr(pgxNum)

// Time conversions
pgxTime := connection.ToPgxTimestamptz(&myTime)
timePtr := connection.FromPgxTimestamptzPtr(pgxTime)

// UUID conversions
pgxUUID := connection.ToPgxUUID(myUUID)
myUUID := connection.FromPgxUUID(pgxUUID)
```

## Error Handling

Structured error types for consistent error handling:

```go
import "github.com/nhalm/dbutil/connection"

// Create structured errors
err := connection.NewNotFoundError("User", userID)
err := connection.NewValidationError("Email", "create", "address", "invalid format", nil)
err := connection.NewDatabaseError("Order", "query", originalErr)

// Use with errors.As for type checking
var notFoundErr *connection.NotFoundError
if errors.As(err, &notFoundErr) {
    log.Printf("Entity not found: %s", notFoundErr.Entity)
}
```

## Migration Guide

If you're upgrading from v1, the main changes are:

1. **Import paths**: Connection utilities moved to `github.com/nhalm/dbutil/connection`
2. **New pagination**: Added `dbutil.Paginate()` for cursor-based pagination
3. **Same API**: All existing connection APIs remain the same

### Before (v1):
```go
import "github.com/nhalm/dbutil"

conn, err := dbutil.NewConnection(ctx, "", sqlc.New)
```

### After (v2):
```go
import "github.com/nhalm/dbutil/connection"

conn, err := connection.NewConnection(ctx, "", sqlc.New)
```

## License

MIT License - see [LICENSE](LICENSE) file for details.