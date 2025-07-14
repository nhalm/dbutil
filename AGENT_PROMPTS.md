# Agent Prompts for dbutil-gen Development

## Overview
This document contains the exact prompts to use when starting fresh agents for each development step. Each agent works on ONE step only, then hands off to the next fresh agent.

---

## Step 1: Test Infrastructure Setup

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 1: Test Infrastructure Setup from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 1 deliverables
- Set up tracking for this step's work

My step focus: Set up foundational testing infrastructure that all subsequent development will depend on (Docker Compose + PostgreSQL + Makefile + Test Schema)
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 1: Test Infrastructure Setup

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- docker-compose.yml with PostgreSQL 15
- test/sql/init.sql with comprehensive test schema
- Makefile with all targets (build, test, integration-test, etc.)
- Updated .gitignore excluding /bin/
- All commands working: make test-setup, make build, make integration-test
```

---

## Step 2: Core Type System & Database Introspection

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 2: Core Type System & Database Introspection from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 2 deliverables
- Set up tracking for this step's work

My step focus: Test and validate the foundational type system and database introspection that converts PostgreSQL schemas into Go data structures. Write comprehensive tests for existing code in gen/types.go, gen/introspect.go, gen/types_mapping.go.
I'm a fresh agent starting Step 2: Core Type System & Database Introspection from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 2 deliverables
- Set up tracking for this step's work

My step focus: Test and validate the foundational type system and database introspection that converts PostgreSQL schemas into Go data structures. Write comprehensive tests for existing code in gen/types.go, gen/introspect.go, gen/types_mapping.go.
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 2: Core Type System & Database Introspection

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- gen/types_test.go with comprehensive unit tests
- gen/introspect_test.go with integration tests
- gen/types_mapping_test.go with type mapping tests
- All PostgreSQL data types mapping correctly to Go types
- UUID v7 primary key validation working
- Integration tests passing against test database
```

---

## Step 3: Code Generation Engine

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 3: Code Generation Engine from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 3 deliverables
- Set up tracking for this step's work

My step focus: Build the code generation engine that creates Go repositories with CRUD operations. Transform introspected schema into working, compilable Go code. Implement file writing functionality that's currently stubbed.
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 3: Code Generation Engine

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- gen/codegen_test.go with unit tests for code generation
- Complete file writing implementation (no more stubs)
- Integration tests verifying generated code compiles
- Integration tests verifying CRUD operations work against test database
- Generated code follows *_generated.go naming pattern
- All generated code properly formatted with gofmt
```

---

## Step 4: Pagination Integration

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 4: Pagination Integration from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 4 deliverables
- Set up tracking for this step's work

My step focus: Implement inline pagination logic directly in generated repositories. Generate both List() and ListPaginated() methods with zero external dependencies using UUID v7 time-ordered cursors.
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 4: Pagination Integration

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- Updated CRUD templates with both List() and ListPaginated() methods
- Generated inline pagination logic (no external dependencies)
- Generated PaginationResult struct in each repository file
- Cursor-based pagination works with UUID v7 ordering
- Integration tests for inline pagination functionality
- Pagination handles edge cases (empty results, large datasets)
- Generated code compiles with zero external pagination dependencies
```

---

## Step 5: CLI Integration & End-to-End Testing

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 5: CLI Integration & End-to-End Testing from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 5 deliverables
- Set up tracking for this step's work

My step focus: Wire up the CLI interface with the code generation engine and create comprehensive end-to-end tests. Make the tool usable by developers with proper error handling and user feedback.
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 5: CLI Integration & End-to-End Testing

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- Complete CLI implementation with all flags working
- End-to-end integration tests using the CLI
- Error handling for common failure scenarios
- User-friendly error messages and help text
- Configuration file support working
- Include/exclude filtering working
- Verbose mode providing helpful output
- CLI can generate repositories from real databases
```

---

## Step 6: Query-Based Code Generation

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 6: Query-Based Code Generation from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 6 deliverables
- Set up tracking for this step's work

My step focus: Implement query-based code generation that parses SQL files with sqlc-style annotations and generates corresponding Go functions. Support :one, :many, :exec, :paginated query types.
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 6: Query-Based Code Generation

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- gen/query_parser.go with SQL file parsing and annotations
- gen/query_analyzer.go with query analysis using PostgreSQL
- Updated code generation for query-based functions
- Integration tests for query-based generation
- Example SQL files with annotations
- All query types (:one, :many, :exec, :paginated) supported
- Query parameters properly typed
- Generated query functions compile and work
```

---

## Step 7: Documentation & Polish

### **Opening Prompt (Copy/Paste to Fresh Agent):**
```
I'm a fresh agent starting Step 7: Documentation & Polish from DEVELOPMENT_PLAN.md

Let me get oriented:
1. Reading PRD_DATABASE_GENERATOR.md Implementation Status
2. Reviewing codebase to understand current state
3. Understanding what previous steps built

Can you help me:
- Verify what the previous agent completed
- Understand the current codebase state
- Create a focused plan for Step 7 deliverables
- Set up tracking for this step's work

My step focus: Create comprehensive documentation, examples, and polish the developer experience. Make the tool ready for public use with performance optimization and migration guides.
```

### **Completion Prompt (When Agent is Ready to Finish):**
```
I'm completing Step 7: Documentation & Polish

Please help me create a comprehensive handoff:
1. Verify all deliverables are complete
2. Update PRD with implementation notes
3. Document what the next agent needs to know
4. Ensure all success criteria are met
5. Test that everything works for the next agent

Expected deliverables to verify:
- Comprehensive README.md with examples
- API documentation for all public functions
- Example projects demonstrating usage
- Performance benchmarks and optimization
- Migration guide for UUID v7 adoption
- Tool ready for public release
- All PRD success criteria met
- Performance meets requirements (<30s for 100 tables)
```

---

## General Guidelines for All Steps

### **What Each Agent Should Do:**
1. **Start** with the opening prompt in a fresh chat
2. **Focus** only on their assigned step
3. **Complete** all deliverables thoroughly
4. **Test** everything works
5. **Document** their work in the PRD
6. **End** with the completion prompt

### **What Each Agent Should NOT Do:**
- Work on future steps
- Leave deliverables incomplete
- Skip testing or documentation
- Make assumptions about previous work without verification

### **Success Criteria:**
- All deliverables for the step are complete
- All tests pass
- Integration with existing code works
- PRD is updated with implementation notes
- Next agent has clear context to continue

### **Verification Commands:**
Each agent should verify their work with:
```bash
make clean          # Start fresh
make test-setup     # Start test infrastructure
make build          # Build the binary
make test           # Run unit tests
make integration-test # Run integration tests
```

---

## Project Completion

When all 7 steps are complete, the dbutil-gen tool should be:
- ✅ Fully functional for table-based code generation
- ✅ Fully functional for query-based code generation
- ✅ Well-tested with comprehensive test coverage
- ✅ Well-documented with examples and guides
- ✅ Ready for public release and community use

The PRD_DATABASE_GENERATOR.md should have all success criteria marked as complete and comprehensive implementation notes from all 7 steps. 