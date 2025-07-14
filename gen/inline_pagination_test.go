package gen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlinePagination_TemplateGeneration(t *testing.T) {
	// Create temporary directory for test output
	tempDir := t.TempDir()

	config := &Config{
		OutputDir:   tempDir,
		PackageName: "repositories",
		Verbose:     false,
	}

	cg := NewCodeGenerator(config)
	table := getTestTable()

	// Test shared pagination types generation
	err := cg.GenerateSharedPaginationTypes()
	if err != nil {
		t.Fatalf("GenerateSharedPaginationTypes failed: %v", err)
	}

	// Read the generated pagination file
	paginationFile := cg.config.GetOutputPath("pagination.go")
	paginationContent, err := os.ReadFile(paginationFile)
	if err != nil {
		t.Fatalf("Failed to read pagination file: %v", err)
	}
	paginationTypes := string(paginationContent)

	// Check that all required components are present in shared pagination file
	expectedComponents := []string{
		"type PaginationParams struct",
		"type PaginationResult[T any] struct",
		"func encodeCursor(id uuid.UUID) string",
		"func decodeCursor(cursor string) (uuid.UUID, error)",
		"func validatePaginationParams(params PaginationParams) error",
		"Items []T `json:\"items\"`",
		"HasMore bool `json:\"has_more\"`",
		"NextCursor string `json:\"next_cursor,omitempty\"`",
		"base64.URLEncoding.EncodeToString(id[:])",
		"base64.URLEncoding.DecodeString(cursor)",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(paginationTypes, component) {
			t.Errorf("Pagination types missing component: %s", component)
		}
	}

	// Test repository generation with shared types
	repositoryCode, err := cg.generateRepositoryWithSharedTypes(table)
	if err != nil {
		t.Fatalf("generateRepositoryWithSharedTypes failed: %v", err)
	}

	expectedListComponents := []string{
		"func (r *UsersRepository) ListPaginated(ctx context.Context, params PaginationParams) (*PaginationResult[Users], error)",
		"validatePaginationParams(params)",
		"decodeCursor(params.Cursor)",
		"encodeCursor(lastItem.GetID())",
		"WHERE ($1::uuid IS NULL OR id > $1)",
		"ORDER BY id ASC",
		"LIMIT $2",
		"hasMore := len(items) > limit",
		"items = items[:limit]",
	}

	for _, component := range expectedListComponents {
		if !strings.Contains(repositoryCode, component) {
			t.Errorf("Repository code missing component: %s", component)
		}
	}
}

func TestInlinePagination_EndToEnd(t *testing.T) {
	// Get test database connection
	pool := getTestDB(t)
	defer pool.Close()

	// Create temporary directory for generated code
	tempDir := t.TempDir()

	// Configure generator for users table only
	config := &Config{
		DSN:         os.Getenv("TEST_DATABASE_URL"),
		Schema:      "public",
		OutputDir:   tempDir,
		PackageName: "testgen",
		Tables:      true,
		Include:     []string{"users"},
		Verbose:     false,
	}

	// Generate code
	generator := New(config)
	ctx := context.Background()

	err := generator.Generate(ctx)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check that file was generated
	generatedFile := filepath.Join(tempDir, "users_generated.go")
	if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
		t.Fatal("Generated file does not exist")
	}

	// Read the generated file
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify shared pagination components
	sharedPaginationComponents := []string{
		// Zero external dependencies
		"package testgen",
		"import (",
		"\"context\"",
		"\"fmt\"",
		"\"github.com/google/uuid\"",
		"\"github.com/jackc/pgx/v5/pgtype\"",
		"\"github.com/jackc/pgx/v5/pgxpool\"",
		")",

		// Both list methods
		"func (r *UsersRepository) List(ctx context.Context) ([]Users, error)",
		"func (r *UsersRepository) ListPaginated(ctx context.Context, params PaginationParams) (*PaginationResult[Users], error)",

		// GetID method with value receiver
		"func (u Users) GetID() uuid.UUID",

		// Pagination logic
		"validatePaginationParams(params)",
		"decodeCursor(params.Cursor)",
		"encodeCursor(lastItem.GetID())",
		"WHERE ($1::uuid IS NULL OR id > $1)",
		"ORDER BY id ASC",
		"LIMIT $2",
		"hasMore := len(items) > limit",
		"items = items[:limit]",
	}

	for _, component := range sharedPaginationComponents {
		if !strings.Contains(contentStr, component) {
			t.Errorf("Generated file missing shared pagination component: %s", component)
		}
	}

	// Verify NO external dependencies
	externalDependencies := []string{
		"github.com/nhalm/dbutil",
		"dbutil.Paginate",
		"dbutil.PaginationParams",
		"dbutil.PaginationResult",
		"gen.Paginate",
		"gen.PaginationParams",
		"gen.PaginationResult",
	}

	for _, dep := range externalDependencies {
		if strings.Contains(contentStr, dep) {
			t.Errorf("Generated file contains external dependency: %s", dep)
		}
	}

	// Test that the generated code compiles
	goModContent := `module testgen

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.5
)
`

	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test compilation
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated code failed to compile: %v\nOutput: %s", err, string(output))
	}

	t.Logf("Generated inline pagination code compiled successfully")
}

func TestInlinePagination_DualListMethods(t *testing.T) {
	// Get test database connection
	pool := getTestDB(t)
	defer pool.Close()

	// Create temporary directory for generated code
	tempDir := t.TempDir()

	// Configure generator
	config := &Config{
		DSN:         os.Getenv("TEST_DATABASE_URL"),
		Schema:      "public",
		OutputDir:   tempDir,
		PackageName: "testgen",
		Tables:      true,
		Include:     []string{"users"},
		Verbose:     false,
	}

	// Generate code
	generator := New(config)
	ctx := context.Background()

	err := generator.Generate(ctx)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Read the generated file
	generatedFile := filepath.Join(tempDir, "users_generated.go")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Test that both list methods exist with correct signatures
	listMethod := "func (r *UsersRepository) List(ctx context.Context) ([]Users, error)"
	if !strings.Contains(contentStr, listMethod) {
		t.Error("Missing simple List method")
	}

	paginatedListMethod := "func (r *UsersRepository) ListPaginated(ctx context.Context, params PaginationParams) (*PaginationResult[Users], error)"
	if !strings.Contains(contentStr, paginatedListMethod) {
		t.Error("Missing paginated ListPaginated method")
	}

	// Test that simple list method doesn't have pagination logic
	simpleListStart := strings.Index(contentStr, "func (r *UsersRepository) List(ctx context.Context) ([]Users, error)")
	paginatedListStart := strings.Index(contentStr, "func (r *UsersRepository) ListPaginated(ctx context.Context, params PaginationParams) (*PaginationResult, error)")

	if simpleListStart == -1 || paginatedListStart == -1 {
		t.Fatal("Could not find both list methods")
	}

	simpleListCode := contentStr[simpleListStart:paginatedListStart]

	// Simple list should NOT have pagination logic
	paginationOnlyComponents := []string{
		"validatePaginationParams",
		"decodeCursor",
		"encodeCursor",
		"hasMore",
		"NextCursor",
		"PaginationResult",
		"WHERE ($1::uuid IS NULL OR id > $1)",
		"LIMIT $2",
	}

	for _, component := range paginationOnlyComponents {
		if strings.Contains(simpleListCode, component) {
			t.Errorf("Simple List method contains pagination logic: %s", component)
		}
	}

	// Simple list should have basic query
	if !strings.Contains(simpleListCode, "ORDER BY id ASC") {
		t.Error("Simple List method missing ORDER BY clause")
	}

	if strings.Contains(simpleListCode, "LIMIT") {
		t.Error("Simple List method should not have LIMIT clause")
	}
}

func TestInlinePagination_CursorLogic(t *testing.T) {
	// Create temporary directory for test output
	tempDir := t.TempDir()

	config := &Config{
		OutputDir:   tempDir,
		PackageName: "repositories",
		Verbose:     false,
	}

	cg := NewCodeGenerator(config)

	// Generate shared pagination types
	err := cg.GenerateSharedPaginationTypes()
	if err != nil {
		t.Fatalf("GenerateSharedPaginationTypes failed: %v", err)
	}

	// Read the generated pagination file
	paginationFile := cg.config.GetOutputPath("pagination.go")
	paginationContent, err := os.ReadFile(paginationFile)
	if err != nil {
		t.Fatalf("Failed to read pagination file: %v", err)
	}
	paginationTypes := string(paginationContent)

	// Test cursor encoding logic
	if !strings.Contains(paginationTypes, "base64.URLEncoding.EncodeToString(id[:])") {
		t.Error("Missing cursor encoding logic")
	}

	// Test cursor decoding logic
	expectedDecodingComponents := []string{
		"base64.URLEncoding.DecodeString(cursor)",
		"if len(cursorBytes) != 16",
		"copy(id[:], cursorBytes)",
		"return uuid.Nil, fmt.Errorf(\"empty cursor\")",
		"return uuid.Nil, fmt.Errorf(\"invalid cursor format: %w\", err)",
		"return uuid.Nil, fmt.Errorf(\"invalid cursor length: expected 16 bytes, got %d\", len(cursorBytes))",
	}

	for _, component := range expectedDecodingComponents {
		if !strings.Contains(paginationTypes, component) {
			t.Errorf("Missing cursor decoding component: %s", component)
		}
	}

	// Test parameter validation logic
	expectedValidationComponents := []string{
		"if params.Limit < 0",
		"if params.Limit > 100",
		"if params.Cursor != \"\"",
		"decodeCursor(params.Cursor)",
		"return fmt.Errorf(\"limit cannot be negative\")",
		"return fmt.Errorf(\"limit cannot exceed 100\")",
		"return fmt.Errorf(\"invalid cursor: %w\", err)",
	}

	for _, component := range expectedValidationComponents {
		if !strings.Contains(paginationTypes, component) {
			t.Errorf("Missing parameter validation component: %s", component)
		}
	}
}

func TestInlinePagination_ZeroDependencies(t *testing.T) {
	// Get test database connection
	pool := getTestDB(t)
	defer pool.Close()

	// Create temporary directory for generated code
	tempDir := t.TempDir()

	// Configure generator for multiple tables
	config := &Config{
		DSN:         os.Getenv("TEST_DATABASE_URL"),
		Schema:      "public",
		OutputDir:   tempDir,
		PackageName: "testgen",
		Tables:      true,
		Exclude:     []string{"composite_pk_table", "invalid_pk_table"},
		Verbose:     false,
	}

	// Generate code
	generator := New(config)
	ctx := context.Background()

	err := generator.Generate(ctx)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Check all generated files
	files, err := filepath.Glob(filepath.Join(tempDir, "*_generated.go"))
	if err != nil {
		t.Fatalf("Failed to list generated files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No generated files found")
	}

	// Check each generated file for zero dependencies
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", file, err)
		}

		contentStr := string(content)

		// Verify NO external pagination dependencies
		externalDependencies := []string{
			"github.com/nhalm/dbutil",
			"dbutil.Paginate",
			"dbutil.PaginationParams",
			"dbutil.PaginationResult",
			"gen.Paginate",
			"gen.PaginationParams",
			"gen.PaginationResult",
		}

		for _, dep := range externalDependencies {
			if strings.Contains(contentStr, dep) {
				t.Errorf("File %s contains external dependency: %s", filepath.Base(file), dep)
			}
		}

		// Verify required shared pagination components exist (should NOT be in individual files)
		sharedComponents := []string{
			"type PaginationParams struct",
			"type PaginationResult struct",
			"func encodeCursor(id uuid.UUID) string",
			"func decodeCursor(cursor string) (uuid.UUID, error)",
			"func validatePaginationParams(params PaginationParams) error",
		}

		// These should NOT be in individual repository files (they're in pagination.go)
		for _, component := range sharedComponents {
			if strings.Contains(contentStr, component) {
				t.Errorf("File %s contains shared component that should be in pagination.go: %s", filepath.Base(file), component)
			}
		}
	}

	// Verify pagination.go exists and contains shared components
	paginationFile := filepath.Join(tempDir, "pagination.go")
	paginationContent, err := os.ReadFile(paginationFile)
	if err != nil {
		t.Fatalf("Failed to read pagination.go: %v", err)
	}

	paginationStr := string(paginationContent)
	requiredSharedComponents := []string{
		"type PaginationParams struct",
		"type PaginationResult[T any] struct",
		"func encodeCursor(id uuid.UUID) string",
		"func decodeCursor(cursor string) (uuid.UUID, error)",
		"func validatePaginationParams(params PaginationParams) error",
	}

	for _, component := range requiredSharedComponents {
		if !strings.Contains(paginationStr, component) {
			t.Errorf("pagination.go missing required shared component: %s", component)
		}
	}

	// Test that all generated code compiles together
	goModContent := `module testgen

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.5
)
`

	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test compilation
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated code failed to compile: %v\nOutput: %s", err, string(output))
	}

	t.Logf("Successfully compiled %d generated repository files with zero external dependencies", len(files))
}

func TestInlinePagination_GetIDMethod(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())
	table := getTestTable()

	// Generate struct code
	structCode, err := cg.generateStruct(table)
	if err != nil {
		t.Fatalf("generateStruct failed: %v", err)
	}

	// Test that GetID method uses value receiver, not pointer receiver
	expectedGetIDSignature := "func (u Users) GetID() uuid.UUID"
	if !strings.Contains(structCode, expectedGetIDSignature) {
		t.Errorf("GetID method should use value receiver, not pointer receiver")
	}

	// Test that GetID method returns the correct field
	if !strings.Contains(structCode, "return u.Id") {
		t.Error("GetID method should return u.Id")
	}

	// Ensure we don't have the old pointer receiver version
	oldPointerSignature := "func (u *Users) GetID() uuid.UUID"
	if strings.Contains(structCode, oldPointerSignature) {
		t.Error("GetID method should not use pointer receiver")
	}
}
