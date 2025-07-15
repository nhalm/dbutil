package gen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCodeGeneration_EndToEnd tests the complete code generation pipeline
func TestCodeGeneration_EndToEnd(t *testing.T) {
	// Get test database connection
	pool := getTestDB(t)
	defer pool.Close()

	// Create temporary directory for generated code
	tempDir := t.TempDir()

	// Configure generator with explicit functions
	config := &Config{
		DSN:         "postgres://dbutil:dbutil_test_password@localhost:5432/dbutil_test",
		Schema:      "public",
		OutputDir:   tempDir,
		PackageName: "testgen",
		Tables:      true,
		Include:     []string{"users"}, // Only generate for users table
		TableConfigs: map[string]TableConfig{
			"users": {
				Functions: []string{"create", "get", "update", "delete", "list", "paginate"},
			},
		},
		Verbose: false,
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

	// Run go mod tidy to generate go.sum
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	tidyCmd.Env = append(os.Environ(), "GO111MODULE=on")

	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, string(tidyOutput))
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

	// Configure generator for all valid tables with explicit functions
	config := &Config{
		DSN:         "postgres://dbutil:dbutil_test_password@localhost:5432/dbutil_test",
		Schema:      "public",
		OutputDir:   tempDir,
		PackageName: "testgen",
		Tables:      true,
		Include:     []string{"users", "profiles", "posts", "comments", "categories", "post_categories", "files", "data_types_test"},
		TableConfigs: map[string]TableConfig{
			"users":           {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"profiles":        {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"posts":           {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"comments":        {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"categories":      {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"post_categories": {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"files":           {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
			"data_types_test": {Functions: []string{"create", "get", "update", "delete", "list", "paginate"}},
		},
		Verbose: false,
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

	// Run go mod tidy to generate go.sum
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	tidyCmd.Env = append(os.Environ(), "GO111MODULE=on")

	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, string(tidyOutput))
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
