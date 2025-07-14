package gen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCodeGeneration_EndToEnd tests the complete code generation pipeline
func TestCodeGeneration_EndToEnd(t *testing.T) {
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
		Include:     []string{"users"}, // Only generate for users table
		Verbose:     false,
	}

	// Create and run generator
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

	// Create a go.mod file for the generated code
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

	// Test that the generated code compiles
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated code failed to compile: %v\nOutput: %s", err, string(output))
	}

	t.Logf("Generated code compiled successfully")
}

// TestGeneratedCode_CompilationOnly tests that generated code compiles for all valid tables
func TestGeneratedCode_CompilationOnly(t *testing.T) {
	// Get test database connection
	pool := getTestDB(t)
	defer pool.Close()

	// Create temporary directory for generated code
	tempDir := t.TempDir()

	// Configure generator for all valid tables (excluding problematic ones)
	config := &Config{
		DSN:         os.Getenv("TEST_DATABASE_URL"),
		Schema:      "public",
		OutputDir:   tempDir,
		PackageName: "testgen",
		Tables:      true,
		Exclude:     []string{"composite_pk_table", "invalid_pk_table"},
		Verbose:     false,
	}

	// Create and run generator
	generator := New(config)
	ctx := context.Background()

	err := generator.Generate(ctx)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Create a go.mod file for the generated code
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

	// Test that all generated code compiles
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Generated code failed to compile: %v\nOutput: %s", err, string(output))
	}

	// Count generated files
	files, err := filepath.Glob(filepath.Join(tempDir, "*_generated.go"))
	if err != nil {
		t.Fatalf("Failed to list generated files: %v", err)
	}

	expectedFileCount := 8 // Based on valid tables in test schema
	if len(files) != expectedFileCount {
		t.Errorf("Expected %d generated files, got %d", expectedFileCount, len(files))
	}

	t.Logf("Successfully compiled %d generated repository files", len(files))
}

// TestGeneratedCode_StructValidation tests that generated structs have correct structure
func TestGeneratedCode_StructValidation(t *testing.T) {
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

	// Read the generated file
	generatedFile := filepath.Join(tempDir, "users_generated.go")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Validate struct has all expected fields with correct types
	expectedFields := map[string]string{
		"Id":                "uuid.UUID",
		"Name":              "string",
		"Email":             "string",
		"PasswordHash":      "string",
		"IsActive":          "pgtype.Bool",
		"CreatedAt":         "pgtype.Timestamptz",
		"UpdatedAt":         "pgtype.Timestamptz",
		"LastLogin":         "pgtype.Timestamptz",
		"Metadata":          "pgtype.JSON",
		"Age":               "pgtype.Int4",
		"Balance":           "pgtype.Float8",
		"ProfilePictureUrl": "pgtype.Text",
	}

	for field, expectedType := range expectedFields {
		fieldDeclaration := field + " " + expectedType
		if !strings.Contains(contentStr, fieldDeclaration) {
			t.Errorf("Generated struct missing field: %s %s", field, expectedType)
		}
	}

	// Validate GetID method exists (should use value receiver, not pointer)
	if !strings.Contains(contentStr, "func (u Users) GetID() uuid.UUID") {
		t.Error("Generated struct missing GetID method")
	}

	// Validate repository methods exist
	expectedMethods := []string{
		"func NewUsersRepository",
		"func (r *UsersRepository) GetByID",
		"func (r *UsersRepository) Create",
		"func (r *UsersRepository) Update",
		"func (r *UsersRepository) Delete",
		"func (r *UsersRepository) List",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, method) {
			t.Errorf("Generated repository missing method: %s", method)
		}
	}
}
