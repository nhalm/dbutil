# PRD: Database Code Generator for dbutil

## Overview
A database-first code generator that connects to existing PostgreSQL databases, introspects the schema, and generates type-safe Go repositories with built-in pagination support using pgx.

## Problem Statement
Currently, developers using dbutil must:
1. Manually write CRUD operations for each table
2. Manually implement pagination for each query
3. Maintain struct definitions that match database schema
4. Write repetitive boilerplate code for basic database operations

This leads to:
- Increased development time
- Potential for bugs and inconsistencies
- Manual maintenance when schema changes
- Repetitive code across projects

## Goals
### Primary Goals
1. **Eliminate boilerplate**: Generate CRUD + pagination automatically
2. **Database-first**: Work with existing database schemas, not code-first models
3. **Type safety**: Generate fully typed Go code using pgx
4. **Built-in pagination**: Every list operation includes cursor-based pagination
5. **Zero dependencies**: Only require pgx and dbutil

### Secondary Goals
1. **Incremental adoption**: Work alongside existing hand-written code
2. **Customizable**: Allow overrides and custom methods
3. **Performance**: Generate efficient queries
4. **Maintainable**: Clean, readable generated code

## Target Users
1. **Go developers** using PostgreSQL
2. **Teams** wanting database-first development
3. **Projects** with existing databases
4. **Developers** frustrated with ORM limitations

## Solution Overview
A CLI tool `dbutil-gen` that:
1. Connects to PostgreSQL database
2. Introspects schema (tables, columns, types, constraints)
3. Analyzes custom SQL queries for complex operations
4. Generates Go repositories with CRUD + pagination + custom queries
5. Uses existing dbutil pagination utilities
6. Integrates with `go generate` workflow

## Detailed Requirements

### Functional Requirements

#### FR1: Database Connection
- **MUST** connect to PostgreSQL using connection string
- **MUST** support connection via environment variables
- **MUST** handle connection failures gracefully
- **SHOULD** support connection pooling for introspection

#### FR2: Schema Introspection
- **MUST** read table definitions from `information_schema`
- **MUST** detect column names, types, and nullability
- **MUST** identify primary keys and validate they are UUID type
- **MUST** detect foreign key relationships
- **MUST** reject tables with non-UUID primary keys with helpful error message
- **SHOULD** support custom schemas (not just `public`)
- **SHOULD** allow table filtering (include/exclude patterns)

#### FR3: Type Mapping
- **MUST** map PostgreSQL types to Go types:
  - `uuid` → `uuid.UUID` (primary keys must be UUID v7)
  - `text/varchar` → `string`
  - `integer` → `int32`
  - `bigint` → `int64`
  - `boolean` → `bool`
  - `timestamp/timestamptz` → `time.Time`
  - `jsonb` → `json.RawMessage`
- **MUST** handle nullable columns with `pgtype` equivalents
- **MUST** validate that primary key columns are UUID type
- **SHOULD** allow custom type mappings via configuration

#### FR4: Code Generation
- **MUST** generate for each table:
  - Struct representing table row
  - Repository struct with connection
  - Constructor function
  - `GetID()` method for pagination interface
- **MUST** generate CRUD operations:
  - `GetByID(ctx, id) (*T, error)`
  - `Create(ctx, params) (*T, error)`
  - `Update(ctx, id, params) (*T, error)`
  - `Delete(ctx, id) error`
- **MUST** generate pagination:
  - `List(ctx, params) (*PaginationResult[T], error)`
- **SHOULD** generate common queries:
  - `GetByUniqueColumn` for unique constraints
  - `ListByForeignKey` for foreign key relationships

#### FR5: Pagination Integration
- **MUST** use existing `dbutil.PaginationParams` and `PaginationResult[T]`
- **MUST** use existing `dbutil.Paginate()` function
- **MUST** implement `HasID` interface for all generated structs
- **MUST** assume all primary keys are UUID v7 (time-sortable)
- **MUST** generate pagination queries that sort by `id ASC` for consistent ordering
- **MUST** reject tables without UUID primary keys with clear error message

#### FR6: Query-Based Generation
- **MUST** parse SQL files with custom queries
- **MUST** support sqlc-style query annotations:
  - `-- name: QueryName :many` (returns slice)
  - `-- name: QueryName :one` (returns single item)
  - `-- name: QueryName :paginated` (returns paginated result)
  - `-- name: QueryName :exec` (returns affected rows)
- **MUST** analyze SELECT columns to determine output struct types
- **MUST** execute EXPLAIN on queries to validate syntax and get column types
- **MUST** generate structs for query result sets
- **MUST** generate functions that execute the queries with proper types
- **SHOULD** support query parameters with proper Go type mapping
- **SHOULD** detect and handle JOIN operations
- **SHOULD** support complex WHERE clauses and aggregations

#### FR7: File Output
- **MUST** generate clean, formatted Go code
- **MUST** include proper package declarations
- **MUST** include necessary imports
- **MUST** add generation comments (`// Code generated by dbutil-gen. DO NOT EDIT.`)
- **MUST** use `*_generated.go` naming pattern for all generated files
- **MUST** completely overwrite generated files on each run
- **MUST** never modify files without `_generated.go` suffix
- **SHOULD** organize output into separate files per table/query group
- **SHOULD** support custom output directory
- **SHOULD** provide clear separation between generated and custom code

### Non-Functional Requirements

#### NFR1: Performance
- **MUST** complete generation in under 30 seconds for 100 tables
- **SHOULD** use connection pooling for schema queries
- **SHOULD** parallelize table introspection

#### NFR2: Reliability
- **MUST** handle database connection failures
- **MUST** validate generated code compiles
- **MUST** provide clear error messages
- **SHOULD** include retry logic for transient failures

#### NFR3: Usability
- **MUST** integrate with `go generate`
- **MUST** provide helpful CLI help text
- **SHOULD** support configuration files
- **SHOULD** provide progress indicators

#### NFR4: Maintainability
- **MUST** generate readable, idiomatic Go code
- **MUST** follow Go naming conventions
- **SHOULD** include documentation comments
- **SHOULD** format code with `gofmt`

## Technical Design

### CLI Interface
```bash
# Basic usage (tables only)
dbutil-gen --dsn="postgres://user:pass@host/db" --output="./repositories"

# With custom queries
dbutil-gen --dsn="..." --queries="./queries" --output="./repos"

# Tables + queries combined
dbutil-gen --dsn="..." --queries="./queries" --tables --output="./repos"

# With filtering
dbutil-gen --dsn="..." --include="users,posts" --exclude="migrations"

# With custom schema
dbutil-gen --dsn="..." --schema="app" --output="./repos"

# Via environment variable
DATABASE_URL="postgres://..." dbutil-gen --queries="./sql" --output="./repos"
```

### Configuration File Support
```yaml
# dbutil-gen.yaml
database:
  dsn: "postgres://localhost/mydb"
  schema: "public"

output:
  directory: "./repositories"
  package: "repos"

# Table-based generation
tables:
  enabled: true
  include: ["users", "posts", "comments"]
  exclude: ["migrations", "*_temp"]

# Query-based generation
queries:
  enabled: true
  directory: "./queries"
  files: ["*.sql"]

types:
  mappings:
    "text": "string"
    "uuid": "github.com/google/uuid.UUID"
```

### Generated Code Structure
```
repositories/
├── user_repository_generated.go    # Generated from users table (overwritten each run)
├── post_repository_generated.go    # Generated from posts table (overwritten each run)
├── comment_repository_generated.go # Generated from comments table (overwritten each run)
├── user_queries_generated.go       # Generated from queries/users.sql (overwritten each run)
├── analytics_queries_generated.go  # Generated from queries/analytics.sql (overwritten each run)
├── types_generated.go              # Common types/interfaces (overwritten each run)
├── user_repository.go              # Custom user code (never touched by generator)
├── post_repository.go              # Custom user code (never touched by generator)
└── analytics_queries.go            # Custom user code (never touched by generator)
```

### Integration with go generate
```go
//go:generate dbutil-gen --config=dbutil-gen.yaml

// Or inline
//go:generate dbutil-gen --dsn="postgres://localhost/mydb" --queries="./sql" --output="./repos"
```

## Query-Based Generation Examples

### Input: SQL Query Files
```sql
-- queries/users.sql

-- name: GetUserWithProfile :many
SELECT 
    u.id,
    u.name,
    u.email,
    u.created_at,
    p.bio,
    p.avatar_url
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
WHERE u.active = true
ORDER BY u.created_at DESC;

-- name: GetUserByEmail :one
SELECT id, name, email, created_at
FROM users 
WHERE email = $1 AND active = true;

-- name: GetActiveUsersPaginated :paginated
SELECT id, name, email, created_at
FROM users 
WHERE active = true
  AND ($1::uuid IS NULL OR id > $1)  -- UUID v7 cursor for time-ordered pagination
ORDER BY id ASC                      -- UUID v7 sorts chronologically
LIMIT $2;

-- name: GetUserStats :one
SELECT 
    COUNT(*) as total_users,
    COUNT(CASE WHEN active THEN 1 END) as active_users,
    AVG(EXTRACT(EPOCH FROM (NOW() - created_at))) as avg_age_seconds
FROM users;

-- name: UpdateUserLastLogin :exec
UPDATE users 
SET last_login = NOW() 
WHERE id = $1;
```

### Generated Output: Go Code

#### File: `user_queries_generated.go`
```go
// Code generated by dbutil-gen. DO NOT EDIT.
// Source: queries/users.sql
package repositories

import (
    "context"
    "time"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/nhalm/dbutil"
)

// Generated structs for query results
type UserWithProfile struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    Name      string     `json:"name" db:"name"`
    Email     string     `json:"email" db:"email"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    Bio       *string    `json:"bio" db:"bio"`
    AvatarURL *string    `json:"avatar_url" db:"avatar_url"`
}

type UserStats struct {
    TotalUsers     int64   `json:"total_users" db:"total_users"`
    ActiveUsers    int64   `json:"active_users" db:"active_users"`
    AvgAgeSeconds  float64 `json:"avg_age_seconds" db:"avg_age_seconds"`
}

// UserQueries contains all custom user-related queries
type UserQueries struct {
    conn *pgxpool.Pool
}

func NewUserQueries(conn *pgxpool.Pool) *UserQueries {
    return &UserQueries{conn: conn}
}

// GetUserWithProfile executes the GetUserWithProfile query
func (q *UserQueries) GetUserWithProfile(ctx context.Context) ([]UserWithProfile, error) {
    const query = `
        SELECT u.id, u.name, u.email, u.created_at, p.bio, p.avatar_url
        FROM users u
        LEFT JOIN profiles p ON u.id = p.user_id
        WHERE u.active = true
        ORDER BY u.created_at DESC
    `
    
    rows, err := q.conn.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []UserWithProfile
    for rows.Next() {
        var item UserWithProfile
        err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.CreatedAt, &item.Bio, &item.AvatarURL)
        if err != nil {
            return nil, err
        }
        results = append(results, item)
    }
    
    return results, rows.Err()
}

// GetUserByEmail executes the GetUserByEmail query
func (q *UserQueries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    const query = `
        SELECT id, name, email, created_at
        FROM users 
        WHERE email = $1 AND active = true
    `
    
    var user User
    err := q.conn.QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// GetActiveUsersPaginated executes the GetActiveUsersPaginated query with pagination
func (q *UserQueries) GetActiveUsersPaginated(ctx context.Context, params dbutil.PaginationParams) (*dbutil.PaginationResult[User], error) {
    return dbutil.Paginate(ctx, params, func(ctx context.Context, cursor *uuid.UUID, limit int32) ([]User, error) {
        const query = `
            SELECT id, name, email, created_at
            FROM users 
            WHERE active = true
              AND ($1::uuid IS NULL OR id > $1)
            ORDER BY id ASC
            LIMIT $2
        `
        
        rows, err := q.conn.Query(ctx, query, cursor, limit)
        if err != nil {
            return nil, err
        }
        defer rows.Close()
        
        var results []User
        for rows.Next() {
            var user User
            err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
            if err != nil {
                return nil, err
            }
            results = append(results, user)
        }
        
        return results, rows.Err()
    })
}

// GetUserStats executes the GetUserStats query
func (q *UserQueries) GetUserStats(ctx context.Context) (*UserStats, error) {
    const query = `
        SELECT 
            COUNT(*) as total_users,
            COUNT(CASE WHEN active THEN 1 END) as active_users,
            AVG(EXTRACT(EPOCH FROM (NOW() - created_at))) as avg_age_seconds
        FROM users
    `
    
    var stats UserStats
    err := q.conn.QueryRow(ctx, query).Scan(&stats.TotalUsers, &stats.ActiveUsers, &stats.AvgAgeSeconds)
    if err != nil {
        return nil, err
    }
    
    return &stats, nil
}

// UpdateUserLastLogin executes the UpdateUserLastLogin query
func (q *UserQueries) UpdateUserLastLogin(ctx context.Context, userID uuid.UUID) error {
    const query = `
        UPDATE users 
        SET last_login = NOW() 
        WHERE id = $1
    `
    
    _, err := q.conn.Exec(ctx, query, userID)
    return err
}
```

### Custom Code Integration

Users can extend generated code by creating companion files without the `_generated.go` suffix:

#### File: `user_queries.go` (Custom code - never overwritten)
```go
package repositories

import (
    "context"
    "github.com/google/uuid"
)

// Custom methods that extend the generated UserQueries
func (q *UserQueries) GetUserWithProfileAndPermissions(ctx context.Context, userID uuid.UUID) (*UserWithProfile, error) {
    // Custom business logic that combines generated methods
    user, err := q.GetUserWithProfile(ctx)
    if err != nil {
        return nil, err
    }
    
    // Additional custom logic here...
    return user, nil
}

// Custom validation logic
func (q *UserQueries) ValidateUserEmail(email string) error {
    // Custom validation logic
    return nil
}
```

#### File: `user_repository.go` (Custom extensions)
```go
package repositories

import (
    "context"
    "github.com/google/uuid"
)

// Custom methods that extend the generated UserRepository
func (r *UserRepository) GetUserWithCachedProfile(ctx context.Context, userID uuid.UUID) (*User, error) {
    // Custom caching logic that uses generated methods
    user, err := r.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Custom caching implementation...
    return user, nil
}
```

### File Management Rules
1. **Generated files** (`*_generated.go`): 
   - Completely overwritten on each run
   - Never edit manually
   - Always have generation header comment

2. **Custom files** (no `_generated.go` suffix):
   - Never touched by generator
   - Can extend generated types and methods
   - Full control over implementation

3. **Composition pattern**:
   - Generated code provides base functionality
   - Custom code extends and composes generated methods
   - Clean separation of concerns

## Success Metrics
1. **Adoption**: 50+ GitHub stars within 6 months
2. **Usage**: 10+ projects using the generator
3. **Performance**: <10 seconds generation time for typical projects
4. **Quality**: 95%+ test coverage for generator code
5. **Developer Experience**: Positive feedback from early adopters

## Risks & Mitigation

### Risk 1: Complex Schema Support
**Risk**: Advanced PostgreSQL features (views, functions, custom types) not supported
**Mitigation**: Start with basic tables, add advanced features incrementally

### Risk 2: Generated Code Quality
**Risk**: Generated code is hard to read or debug
**Mitigation**: Focus on clean, idiomatic Go code generation from day one

### Risk 3: Performance with Large Schemas
**Risk**: Slow generation for databases with hundreds of tables
**Mitigation**: Implement parallel processing and connection pooling

### Risk 4: Maintenance Burden
**Risk**: Supporting edge cases becomes complex
**Mitigation**: Keep initial scope focused, add features based on user feedback

### Risk 5: UUID v7 Requirement
**Risk**: Developers may have existing tables with non-UUID primary keys
**Mitigation**: Provide clear migration guidance and tooling to convert existing tables to UUID v7

## Implementation Phases

### Phase 1: MVP (4-6 weeks)
- Basic CLI tool
- PostgreSQL connection and introspection
- Simple type mapping
- Basic CRUD generation (table-based)
- Pagination integration
- Single table output

### Phase 2: Query-Based Generation (4-5 weeks)
- SQL file parsing and annotation support
- Query analysis and column type detection
- Struct generation from query results
- Function generation for custom queries
- Pagination support for `:paginated` queries
- Multiple query file support

### Phase 3: Polish (2-3 weeks)
- Configuration file support
- Error handling improvements
- Documentation and examples
- go generate integration
- Combined table + query generation

### Phase 4: Advanced Features (3-4 weeks)
- Foreign key relationship detection
- Custom type mappings
- Include/exclude filtering
- Performance optimizations
- Advanced PostgreSQL type support
- Complex query parameter handling

### Phase 5: Ecosystem (2-3 weeks)
- Integration tests with real databases
- Example projects
- Documentation website
- Community feedback incorporation

## UUID v7 Requirement

### Why UUID v7 Only?
1. **Time-ordered**: UUID v7 provides chronological ordering, perfect for pagination
2. **Consistent performance**: Cursor-based pagination with UUID v7 has predictable performance
3. **Simplicity**: Single ID type eliminates complex type handling and edge cases
4. **Modern standard**: UUID v7 is the recommended UUID version for new applications
5. **Database-friendly**: Avoids UUID v4 fragmentation issues in B-tree indexes

### Migration Path for Existing Tables
```sql
-- Example migration from serial ID to UUID v7
ALTER TABLE users ADD COLUMN new_id UUID DEFAULT gen_random_uuid();
UPDATE users SET new_id = gen_random_uuid();
ALTER TABLE users DROP CONSTRAINT users_pkey;
ALTER TABLE users DROP COLUMN id;
ALTER TABLE users RENAME COLUMN new_id TO id;
ALTER TABLE users ADD PRIMARY KEY (id);
```

### Benefits for Pagination
- **Efficient cursors**: UUID v7 values can be directly compared with `>`
- **Time-ordered results**: Natural chronological ordering without separate timestamp
- **Consistent behavior**: All tables paginate the same way
- **Opaque cursors**: Base64-encoded UUID provides security

## Open Questions
1. Should we support other databases beyond PostgreSQL?
2. How should we handle database migrations and schema evolution?
3. Should generated repositories support transactions?
4. How do we handle custom business logic in generated code?
5. Should we generate interfaces for repositories to aid testing?
6. Should we provide tooling to help migrate existing tables to UUID v7?

## Dependencies
- **pgx/v5**: Database driver and connection pooling
- **google/uuid**: UUID type support
- **existing dbutil**: Pagination utilities
- **Go 1.21+**: Generics support required
- **SQL parser**: Custom implementation (no external dependencies like sqlc)
- **AST analysis**: For query parameter and column detection

## Success Criteria
- [ ] Can generate working repositories for any PostgreSQL table
- [ ] Generated code compiles without errors
- [ ] Pagination works out of the box
- [ ] Performance is acceptable for real-world schemas
- [ ] Developer experience is significantly better than manual coding
- [ ] Documentation is comprehensive and clear 