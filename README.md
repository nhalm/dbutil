# Database Utilities

A reusable Go package that provides database connection utilities and testing infrastructure for applications using PostgreSQL with pgx and sqlc.

## Overview

This package is designed specifically for **sqlc users** who want:
- ✅ **Reusable database utilities** that work with any sqlc-generated queries
- ✅ **Optimized testing infrastructure** with shared connections for 10-20x faster tests
- ✅ **Type-safe PostgreSQL operations** with comprehensive pgx type helpers
- ✅ **Flexible configuration** including schema paths and connection settings
- ✅ **Structured error handling** with consistent error types
- ✅ **Health checks and metrics** for production monitoring
- ✅ **Retry logic** for transient failures
- ✅ **Read/write splitting** for scaled deployments
- ✅ **Query logging and tracing** for debugging
- ✅ **Connection hooks** for lifecycle event management

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
    
    "github.com/nhalm/dbutil"
    "your-project/internal/repository/sqlc" // Your sqlc-generated package
)

func main() {
    ctx := context.Background()
    
    // Create connection with your sqlc queries
    conn, err := dbutil.NewConnection(ctx, "", sqlc.New)
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

## Key Features

### 🔗 **Works with Any sqlc Package**
Unlike other database utilities, this package doesn't import specific sqlc-generated code. You provide your own sqlc queries:

```go
// Works with any sqlc-generated queries
conn, err := dbutil.NewConnection(ctx, "", myapp.New)
conn, err := dbutil.NewConnection(ctx, "", yourapp.New)
```

### ⚙️ **Flexible Configuration**
```go
config := &dbutil.Config{
    MaxConns:        20,
    MinConns:        5,
    MaxConnLifetime: 1 * time.Hour,
    SearchPath:      "myschema", // Configurable schema path
}
conn, err := dbutil.NewConnectionWithConfig(ctx, "", sqlc.New, config)
```

### 🚀 **High-Performance Testing**
```go
func TestUserOperations(t *testing.T) {
    conn := dbutil.RequireTestDB(t, sqlc.New) // Shared connection
    dbutil.CleanupTestData(conn,
        "DELETE FROM users WHERE email LIKE 'test_%'",
    )
    // ... test logic
}
```

### 🔄 **Transaction Support**
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

### 🏥 **Health Checks & Monitoring**
```go
// Health checks
if conn.IsReady(ctx) {
    log.Println("Database is ready")
}

// Connection pool stats
stats := conn.Stats()
log.Printf("Active connections: %d", stats.TotalConns())

// With metrics collection
conn = conn.WithMetrics(myMetricsCollector)

// With connection hooks
logger := dbutil.NewDefaultLogger(dbutil.LogLevelInfo)
conn, err = dbutil.NewConnectionWithLoggingHooks(ctx, "", sqlc.New, logger)
```

### 🔄 **Retry Logic**
```go
// Automatic retry for transient failures
retryableConn := conn.WithRetry(nil) // Uses defaults
err = retryableConn.WithRetryableTransaction(ctx, func(ctx context.Context, tx *sqlc.Queries) error {
    return tx.CreateUser(ctx, params)
})
```

### 📊 **Read/Write Splitting**
```go
// Separate read and write connections
rwConn, err := dbutil.NewReadWriteConnection(ctx, readDSN, writeDSN, sqlc.New)
readQueries := rwConn.ReadQueries()   // Use for SELECT queries
writeQueries := rwConn.WriteQueries() // Use for INSERT/UPDATE/DELETE
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

## Testing Strategy

This package implements a **shared connection approach** for integration tests with optimized performance.

### Test Architecture

#### 1. **Unit Tests**
- ✅ **No database required** - Fast execution
- Tests business logic, type conversions, and record comparison
- Always run in CI/CD pipelines

#### 2. **Integration Tests**
- 🔧 **Requires test database** - Real PostgreSQL instance
- Uses **shared database connection** initialized once
- Tests complete workflow with real data

#### 3. **Performance Tests**
- 📊 **Benchmark testing** with large datasets
- Measures performance with 1K-10K+ records

## Key Feature: Shared Database Connection

### **Optimized Pattern**
```go
import "github.com/nhalm/dbutil"
import "your-project/internal/repository/sqlc"

func TestSomething(t *testing.T) {
    conn := dbutil.RequireTestDB(t, sqlc.New)     // Reuse shared connection
    dbutil.CleanupTestData(conn,                  // Clean data, not connection
        "DELETE FROM your_schema.entities WHERE id ~ '^test-'",
    )
    // ... test logic
}
```

### Available Test Utilities

#### `RequireTestDB[T Querier](t TestingT, newQueriesFunc func(*pgxpool.Pool) T) *Connection[T]`
- Returns shared test database connection with your sqlc queries
- Skips test if `TEST_DATABASE_URL` not set
- Works with both `*testing.T` and `*testing.B`

#### `GetTestConnection[T Querier](newQueriesFunc func(*pgxpool.Pool) T) *Connection[T]`
- Returns shared connection (or nil if not available)
- Thread-safe initialization with `sync.Once`

#### `CleanupTestData[T Querier](conn *Connection[T], sqlStatements ...string)`
- Executes cleanup SQL statements
- Handles multiple SQL statements in sequence
- Logs warnings for failed cleanup (doesn't fail tests)

## Benefits of Shared Connection

### **🚀 Performance Improvements**
- **Setup Time**: ~50ms per test → ~50ms total for all tests
- **Connection Overhead**: Eliminated repeated connection establishment
- **Test Suite Speed**: 10-20x faster for integration tests

### **🔒 Resource Efficiency**
- **Database Connections**: 1 connection vs N connections
- **Memory Usage**: Reduced connection pool overhead
- **Database Load**: Less connection churn

### **🧹 Clean Test Isolation**
- Each test gets clean data via `CleanupTestData()`
- Shared connection but isolated test data
- No test interference or data leakage

## Usage

### Running Tests

#### **Unit Tests (Always Run)**
```bash
# Fast unit tests - no database required
go test ./...
```

#### **Integration Tests**
```bash
# Set up test database and run integration tests
export TEST_DATABASE_URL="postgres://user:password@localhost:5433/test_db?sslmode=disable"
go test -tags=integration ./... -v
```

### Basic Usage in Your Application

```go
import (
    "context"
    "github.com/nhalm/database"
    "your-project/internal/repository/sqlc"
)

func main() {
    ctx := context.Background()
    
    // Create connection with your sqlc queries
    conn, err := dbutil.NewConnection(ctx, "", sqlc.New)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Use your queries
    queries := conn.Queries()
    users, err := queries.GetAllUsers(ctx)
    // ... handle results
}
```

### More Examples

See [examples.md](examples.md) for comprehensive usage examples including:
- Custom configuration
- Transaction handling
- Error handling with structured types
- Multiple database schemas
- Type conversion helpers
- Integration testing patterns

### Test Database Setup

#### **Requirements**
1. **Separate Test Database**: Never use development/production DB
2. **Applied Migrations**: Same schema as production
3. **Environment Variable**: `TEST_DATABASE_URL` must be set
4. **Clean State**: Tests handle their own data cleanup

#### **Example Setup**
```bash
# Start PostgreSQL in Docker
docker run --name test-postgres -e POSTGRES_PASSWORD=testpass -p 5433:5432 -d postgres:15

# Apply migrations to test database
your-migrate-tool up

# Set environment variable
export TEST_DATABASE_URL="postgres://postgres:testpass@localhost:5433/postgres?sslmode=disable"

# Run integration tests
go test -tags=integration ./... -v
```

## pgx Type Helpers

The `pgx_helpers.go` file provides conversion utilities for pgx types:

### **Text Conversions**
```go
// String pointer to pgtype.Text
pgxText := ToPgxText(&myString)
pgxText := ToPgxText(nil) // Valid: false

// pgtype.Text to string pointer
stringPtr := FromPgxText(pgxText)
stringVal := FromPgxTextToString(pgxText) // Never nil
```

### **Numeric Conversions**
```go
// Float64 pointer to pgtype.Numeric
pgxNum := ToPgxNumericFromFloat64Ptr(&myFloat)

// pgtype.Numeric to float64 pointer
floatPtr := FromPgxNumericPtr(pgxNum)
```

### **UUID, Time, and Other Types**
```go
// UUID conversions
pgxUUID := ToPgxUUID(myUUID)
myUUID := FromPgxUUID(pgxUUID)

// Time conversions
pgxTime := ToPgxTimestamptz(&myTime)
timePtr := FromPgxTimestamptzPtr(pgxTime)

// Integer conversions
pgxInt := ToPgxInt4FromInt(&myInt)
intPtr := FromPgxInt4(pgxInt)

// Boolean conversions
pgxBool := ToPgxBool(&myBool)
boolPtr := FromPgxBool(pgxBool)
```

## Structured Error Types

The `errors.go` file provides generic database error types for consistent error handling:

### **Database Error Types**
- `NotFoundError` - Entity not found in database
- `ValidationError` - Data validation failures  
- `DatabaseError` - Database operation failures

### **Usage Examples**
```go
// Create structured database errors
err := dbutil.NewNotFoundError("User", userID)
err := database.NewValidationError("Email", "create", "address", "invalid format", nil)
err := database.NewDatabaseError("Order", "query", originalErr)

// Error messages are consistent and informative
// "User not found: 123e4567-e89b-12d3-a456-426614174000"
// "validation failed for Email create: address (invalid format)"
```

**Note**: Application-specific errors should be defined in your service layer where domain logic belongs.

## Test Data Management

### **Data Cleanup Strategy**
```go
// Clean test data between tests
func TestSomething(t *testing.T) {
    conn := dbutil.RequireTestDB(t)
    dbutil.CleanupTestData(conn,
        "DELETE FROM your_schema.entities WHERE name LIKE 'test_%'",
        "DELETE FROM your_schema.relations WHERE entity_id IS NULL",
    )
    // ... test logic
}
```

### **Test Data Patterns**
- **Test Entities**: Use consistent prefixes like "test_", "TEST_", etc.
- **Unique Identifiers**: Use UUIDs or patterns that won't conflict with real data
- **Safe Cleanup**: Use patterns that only match test data
- **Isolation**: Ensure test data doesn't interfere with other tests

## Database Development Workflow

### **Schema Changes**
```bash
# Create new migration
your-migrate-tool create migration_name

# Apply migrations to development
your-migrate-tool up

# Update SQL queries in your queries directory

# Regenerate Go code from SQL (if using sqlc)
sqlc generate

# Update converters and models as needed

# Test changes
go test -tags=integration ./...
```

## CI/CD Integration

### **GitHub Actions Example**
```yaml
- name: Run Unit Tests
  run: go test ./...
  
- name: Setup Test Database
  run: |
    docker run --name test-postgres -e POSTGRES_PASSWORD=testpass -p 5433:5432 -d postgres:15
    # Apply migrations here
    
- name: Run Integration Tests  
  env:
    TEST_DATABASE_URL: postgres://postgres:testpass@localhost:5433/postgres?sslmode=disable
  run: go test -tags=integration ./...
  
- name: Check Coverage
  run: go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### **Performance Monitoring**
```bash
# Run benchmarks
go test -tags=integration -bench=. ./... -benchmem
```

## Best Practices

### **✅ Do**
- Use `dbutil.RequireTestDB(t)` for integration tests
- Clean data between tests with `CleanupTestData()`
- Use distinct test data patterns (test_, TEST_, etc.)
- Use structured error types from `errors.go`
- Use pgx helper functions for type conversions
- Set up separate test databases

### **❌ Don't**
- Create new database connections per test
- Use production data patterns in tests
- Run integration tests without `TEST_DATABASE_URL`
- Mix unit and integration test concerns
- Leave test data in database after tests
- Use direct pgx types without helpers

## Error Handling

### **Graceful Test Degradation**
```go
func TestIntegration(t *testing.T) {
    conn := dbutil.RequireTestDB(t) // Skips if no test DB
    // Test will be skipped, not failed, if TEST_DATABASE_URL not set
}
```

### **Structured Error Usage**
```go
// In repository layer
if err == pgx.ErrNoRows {
    return nil, dbutil.NewNotFoundError("User", id)
}

// In service layer
var notFoundErr *database.NotFoundError
if errors.As(err, &notFoundErr) {
    return service.ErrUserNotFound
}
```

## Connection Features

### **Transaction Support**
```go
// High-level transaction wrapper
err := conn.WithTransaction(ctx, func(ctx context.Context, tx *db.Queries) error {
    // All operations automatically rolled back on error
    return tx.CreateSomething(ctx, params)
})

// Manual transaction control
tx, queries, err := conn.BeginTransaction(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// ... use queries ...

return tx.Commit(ctx)
```

### **Connection Pool Configuration**
- **MaxConns**: 10 connections
- **MinConns**: 1 connection  
- **MaxConnLifetime**: 30 minutes
- **Automatic reconnection** on connection loss

## Integration with sqlc

This package is designed to work seamlessly with [sqlc](https://sqlc.dev/):

```go
// Your sqlc-generated queries work directly with this package
conn, err := dbutil.NewConnection(ctx, "", sqlc.New)
queries := conn.Queries()
user, err := queries.GetUser(ctx, userID)
if err != nil {
    return dbutil.NewNotFoundError("User", userID)
}
```

### Why this approach works better:
- ✅ **No coupling**: Your sqlc package doesn't need to know about this library
- ✅ **Reusable**: Works with any sqlc-generated queries
- ✅ **Type-safe**: Full type safety with generics
- ✅ **Flexible**: Easy to test with different query implementations

## Thread Safety

- All connection operations are thread-safe
- Shared test connection uses `sync.Once` for initialization
- Connection pool handles concurrent access automatically
- Test utilities are safe for parallel test execution