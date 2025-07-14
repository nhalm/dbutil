# Development Plan: Database Code Generator (dbutil-gen)

## Overview
This document breaks down the PRD_DATABASE_GENERATOR.md into discrete, contextual steps that agents can work on independently. Each step builds upon previous work and agents should update the PRD with their progress.

## Important Instructions for All Agents

**⚠️ CRITICAL: Each agent works on ONE step only, then hands off to a fresh agent**

### Before Starting Your Step:
1. **Read PRD_DATABASE_GENERATOR.md** to understand the full context and what previous steps accomplished
2. **Check the "Implementation Status" section** in the PRD to see what has been completed
3. **Review the codebase** to understand the current state
4. **Focus ONLY on your assigned step** - do not work on future steps

### During Your Step:
1. **Stay focused** on your step's deliverables only
2. **Test thoroughly** - the next agent depends on your work being solid
3. **Document decisions** as you make them
4. **Ask questions** if requirements are unclear

### Before Ending Your Session:
1. **Complete ALL deliverables** for your step - do not leave anything unfinished
2. **Verify ALL success criteria** are met for your step
3. **Update PRD_DATABASE_GENERATOR.md** with detailed implementation notes
4. **Create a comprehensive handoff** for the next agent
5. **Test everything works** - the next agent should be able to build on your work immediately

### Handoff Requirements:
- **What you built** (detailed summary)
- **Why you built it that way** (design decisions)
- **What works** (tested and verified)
- **What the next agent needs to know** (context, gotchas, next steps)
- **How to verify your work** (commands to run, tests to check)

---

## Fresh Agent Onboarding Template

**When starting a new step, use this template:**

```
I'm a fresh agent starting Step X: [Step Name] from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step X deliverables
- Set up tracking for this step's work
```

**When ending your step, use this template:**

```
I'm completing Step X: [Step Name]

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent
```

---

## Step 1: Test Infrastructure Setup
**Agent Focus:** DevOps/Testing Infrastructure

### Context
Set up the foundational testing infrastructure that all subsequent development will depend on. This enables TDD approach for the complex code generator.

### Tasks
1. Create `docker-compose.yml` with PostgreSQL 15 container
2. Create `test/sql/init.sql` with comprehensive test schema (see PRD for exact schema)
3. Create `Makefile` with all specified targets (build, test, integration-test, etc.)
4. Add `/bin/` to `.gitignore`
5. Verify the test infrastructure works end-to-end

### Deliverables
- `docker-compose.yml` - PostgreSQL test container
- `test/sql/init.sql` - Test database schema with all data types
- `Makefile` - Build and test targets
- Updated `.gitignore` - Exclude `/bin/` directory
- `README.md` - Instructions for running tests

### Success Criteria
- `make test-setup` successfully starts PostgreSQL container
- `make clean` tears down docker-compose.yml, clears ./bin directory and any other temporary artifacts
- Database initializes with test schema and data
- `make build` creates `bin/dbutil-gen` binary
- `make test` runs (even if no tests exist yet)
- `make integration-test` starts PostgreSQL container if not already running, runs (even if no tests exist yet)

### Notes for Next Agent
- Test infrastructure is ready for unit and integration tests
- Database schema covers all PostgreSQL data types and edge cases
- Build system is standardized and reliable

---

## Step 2: Core Type System & Database Introspection
**Agent Focus:** Database/Type Systems

### Context
Build the foundational type system and database introspection that converts PostgreSQL schemas into Go data structures. This is the core of the code generator.

### Tasks
1. Review existing code in `gen/types.go`, `gen/introspect.go`, `gen/types_mapping.go`
2. Write comprehensive unit tests for type mapping (PostgreSQL → Go)
3. Write integration tests for database introspection
4. Fix any bugs found during testing
5. Ensure UUID v7 primary key validation works correctly

### Deliverables
- `gen/types_test.go` - Unit tests for data structures
- `gen/introspect_test.go` - Integration tests for database introspection
- `gen/types_mapping_test.go` - Unit tests for type mapping
- Bug fixes and improvements to existing code
- Documentation of type mapping decisions

### Success Criteria
- All PostgreSQL data types map correctly to Go types
- Database introspection works against test schema
- UUID v7 primary key validation rejects non-UUID tables
- Nullable types use pgtype correctly
- Array types are handled properly
- `make integration-test` passes for introspection

### Notes for Next Agent
- Type system is fully tested and reliable
- Database introspection handles all edge cases
- Foundation is solid for code generation

---

## Step 3: Code Generation Engine
**Agent Focus:** Code Generation/Templates

### Context
Build the code generation engine that creates Go repositories with CRUD operations. This transforms the introspected schema into working Go code.

### Tasks
1. Review existing code in `gen/codegen.go`, `gen/crud_templates.go`
2. Implement file writing functionality (currently stubbed)
3. Write tests for code generation
4. Ensure generated code compiles and is properly formatted
5. Test CRUD operations against test database

### Deliverables
- `gen/codegen_test.go` - Unit tests for code generation
- Complete file writing implementation
- Integration tests that verify generated code compiles
- Integration tests that verify CRUD operations work
- Example generated code documentation

### Success Criteria
- Generated code compiles without errors
- Generated structs have correct field types and tags
- CRUD operations work against test database
- File naming follows `*_generated.go` pattern
- Generated code is properly formatted with gofmt
- `make integration-test` passes for code generation

### Notes for Next Agent
- Code generation produces working, compilable Go code
- CRUD operations are fully functional
- Foundation ready for pagination integration

---

## Step 4: Inline Pagination Integration
**Agent Focus:** Inline Pagination Code Generation

### Context
Implement pagination as inline generated code with zero external dependencies. Generate both `List()` and `ListPaginated()` methods for each table with complete pagination logic embedded directly in the generated repositories.

### Tasks
1. Design inline pagination architecture (no external dependencies)
2. Create dual list method templates (`List()` and `ListPaginated()`)
3. Generate pagination logic inline: cursor encoding/decoding, validation, query logic
4. Generate `PaginationResult` struct inline in each generated file
5. Test inline pagination with UUID v7 time-ordered cursors

### Deliverables
- Updated CRUD templates with both `List()` and `ListPaginated()` methods
- Generated inline pagination logic (cursor handling, validation, queries)
- Generated `PaginationResult` struct in each repository file
- Integration tests for inline pagination functionality
- Example usage of inline paginated repositories

### Success Criteria
- Generated code has zero external pagination dependencies
- Both `List()` and `ListPaginated()` methods generated for each table
- Cursor-based pagination works with UUID v7 ordering
- Pagination handles edge cases (empty results, large datasets)
- Generated code compiles with only pgx dependencies
- `make integration-test` passes for inline pagination

### Notes for Next Agent
- Inline pagination is fully implemented and tested
- Generated repositories are completely self-contained
- Zero external dependencies for pagination functionality
- Basic table generation with pagination is complete

---

## Step 5: CLI Integration & End-to-End Testing
**Agent Focus:** CLI/UX

### Context
Wire up the CLI interface with the code generation engine and create comprehensive end-to-end tests. This makes the tool usable by developers.

### Tasks
1. Review existing `cmd/dbutil-gen/main.go`
2. Implement complete CLI workflow (flags → config → generation → output)
3. Add comprehensive error handling and user feedback
4. Create end-to-end tests that use the CLI
5. Add progress indicators and verbose logging

### Deliverables
- Complete CLI implementation with all flags working
- End-to-end integration tests using the CLI
- Error handling for common failure scenarios
- User-friendly error messages and help text
- CLI documentation and examples

### Success Criteria
- CLI can generate repositories from real databases
- All command-line flags work correctly
- Configuration file support works
- Include/exclude filtering works
- Verbose mode provides helpful output
- Error messages are clear and actionable
- `make build && bin/dbutil-gen --help` works

### Notes for Next Agent
- CLI is fully functional and user-friendly
- Basic table generation workflow is complete
- Ready for query-based generation features

---

## Step 6: Query-Based Code Generation
**Agent Focus:** SQL Parsing/Query Analysis

### Context
Implement the query-based code generation that parses SQL files with sqlc-style annotations and generates corresponding Go functions.

### Tasks
1. Implement SQL file parsing with annotation support
2. Build query analyzer that uses PostgreSQL EXPLAIN to get column types
3. Generate Go functions for custom queries
4. Support all query types (:one, :many, :exec, :paginated)
5. Handle query parameters and result structs

### Deliverables
- `gen/query_parser.go` - SQL file parsing with annotations
- `gen/query_analyzer.go` - Query analysis using PostgreSQL
- Updated code generation for query-based functions
- Integration tests for query-based generation
- Example SQL files with annotations

### Success Criteria
- SQL files are parsed correctly with annotations
- Query analysis determines correct Go types
- Generated query functions compile and work
- All query types (:one, :many, :exec, :paginated) supported
- Query parameters are properly typed
- `make integration-test` passes for query generation

### Notes for Next Agent
- Query-based generation is fully implemented
- Both table and query generation modes work
- Tool is feature-complete per PRD requirements

---

## Step 7: Documentation & Polish
**Agent Focus:** Documentation/UX Polish

### Context
Create comprehensive documentation, examples, and polish the developer experience. This makes the tool ready for public use.

### Tasks
1. Create comprehensive README with examples
2. Add inline documentation to all public APIs
3. Create example projects showing generated code usage
4. Add performance benchmarks and optimization
5. Create migration guide for existing projects

### Deliverables
- Comprehensive README.md with examples
- API documentation for all public functions
- Example projects demonstrating usage
- Performance benchmarks and optimization
- Migration guide for UUID v7 adoption
- Release preparation

### Success Criteria
- Documentation is comprehensive and clear
- Examples work and demonstrate key features
- Performance meets PRD requirements (<30s for 100 tables)
- Tool is ready for public release
- All PRD success criteria are met

### Notes for Next Agent
- Tool is production-ready and well-documented
- Ready for community feedback and iteration
- Foundation set for future enhancements

---

## Cross-Step Requirements

### PRD Maintenance
Each agent must update `PRD_DATABASE_GENERATOR.md` with:
- New "Implementation Status" section entry
- What was built and why
- Any design decisions or changes
- What the next agent needs to know
- Updated success criteria checkboxes

### Testing Requirements
- All code must have unit tests
- Integration tests must pass
- No regressions in existing functionality
- Test coverage should increase with each step

### Code Quality
- Follow Go best practices and conventions
- Use proper error handling
- Add comprehensive documentation
- Ensure code is maintainable and readable

### Communication
- Document any blockers or design questions
- Note any PRD requirements that need clarification
- Suggest improvements for future steps 