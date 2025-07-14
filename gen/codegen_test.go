package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test data for code generation tests
func getTestTable() Table {
	return Table{
		Name:   "users",
		Schema: "public",
		Columns: []Column{
			{
				Name:         "id",
				Type:         "uuid",
				GoType:       "uuid.UUID",
				IsNullable:   false,
				DefaultValue: "",
				IsArray:      false,
			},
			{
				Name:         "name",
				Type:         "text",
				GoType:       "string",
				IsNullable:   false,
				DefaultValue: "",
				IsArray:      false,
			},
			{
				Name:         "email",
				Type:         "text",
				GoType:       "string",
				IsNullable:   false,
				DefaultValue: "",
				IsArray:      false,
			},
			{
				Name:         "is_active",
				Type:         "boolean",
				GoType:       "pgtype.Bool",
				IsNullable:   true,
				DefaultValue: "true",
				IsArray:      false,
			},
			{
				Name:         "created_at",
				Type:         "timestamptz",
				GoType:       "pgtype.Timestamptz",
				IsNullable:   true,
				DefaultValue: "now()",
				IsArray:      false,
			},
			{
				Name:         "metadata",
				Type:         "jsonb",
				GoType:       "*json.RawMessage",
				IsNullable:   true,
				DefaultValue: "",
				IsArray:      false,
			},
		},
		PrimaryKey: []string{"id"},
		Indexes:    []Index{},
	}
}

func getTestConfig() *Config {
	return &Config{
		OutputDir:   "./test-output",
		PackageName: "repositories",
		Verbose:     false,
	}
}

func TestNewCodeGenerator(t *testing.T) {
	config := getTestConfig()
	cg := NewCodeGenerator(config)

	if cg == nil {
		t.Fatal("NewCodeGenerator returned nil")
	}

	if cg.config != config {
		t.Error("Config not set correctly")
	}

	if cg.typeMapper == nil {
		t.Error("TypeMapper not initialized")
	}
}

func TestCodeGenerator_generateStruct(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())
	table := getTestTable()

	code, err := cg.generateStruct(table)
	if err != nil {
		t.Fatalf("generateStruct failed: %v", err)
	}

	// Check that the struct is generated correctly
	expectedParts := []string{
		"type Users struct {",
		"Id uuid.UUID",
		"Name string",
		"Email string",
		"IsActive pgtype.Bool",
		"CreatedAt pgtype.Timestamptz",
		"Metadata *json.RawMessage",
		"GetID() uuid.UUID",
		"return u.Id",
	}

	for _, part := range expectedParts {
		if !strings.Contains(code, part) {
			t.Errorf("Generated struct missing expected part: %s\nGenerated code:\n%s", part, code)
		}
	}

	// Check struct tags
	if !strings.Contains(code, `json:"id" db:"id"`) {
		t.Error("Missing correct struct tags for id field")
	}

	if !strings.Contains(code, `json:"is_active" db:"is_active"`) {
		t.Error("Missing correct struct tags for is_active field")
	}
}

func TestCodeGenerator_generateRepository(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())
	table := getTestTable()

	code, err := cg.generateRepository(table)
	if err != nil {
		t.Fatalf("generateRepository failed: %v", err)
	}

	// Check that the repository is generated correctly
	expectedParts := []string{
		"type UsersRepository struct {",
		"conn *pgxpool.Pool",
		"func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository",
		"return &UsersRepository{",
		"conn: conn,",
	}

	for _, part := range expectedParts {
		if !strings.Contains(code, part) {
			t.Errorf("Generated repository missing expected part: %s\nGenerated code:\n%s", part, code)
		}
	}
}

func TestCodeGenerator_prepareCRUDTemplateData(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())
	table := getTestTable()

	data, err := cg.prepareCRUDTemplateData(table)
	if err != nil {
		t.Fatalf("prepareCRUDTemplateData failed: %v", err)
	}

	// Check basic template data
	if data["StructName"] != "Users" {
		t.Errorf("Expected StructName 'Users', got %v", data["StructName"])
	}

	if data["RepositoryName"] != "UsersRepository" {
		t.Errorf("Expected RepositoryName 'UsersRepository', got %v", data["RepositoryName"])
	}

	if data["TableName"] != "users" {
		t.Errorf("Expected TableName 'users', got %v", data["TableName"])
	}

	if data["IDColumn"] != "id" {
		t.Errorf("Expected IDColumn 'id', got %v", data["IDColumn"])
	}

	// Check select columns
	selectColumns := data["SelectColumns"].(string)
	expectedColumns := []string{"id", "name", "email", "is_active", "created_at", "metadata"}
	for _, col := range expectedColumns {
		if !strings.Contains(selectColumns, col) {
			t.Errorf("SelectColumns missing column: %s", col)
		}
	}

	// Check create fields (should exclude ID and columns with defaults)
	createFields := data["CreateFields"].([]map[string]string)
	expectedCreateFields := []string{"Name", "Email", "Metadata"}
	if len(createFields) != len(expectedCreateFields) {
		t.Errorf("Expected %d create fields, got %d", len(expectedCreateFields), len(createFields))
	}

	// Check update fields (should include all non-ID columns)
	updateFields := data["UpdateFields"].([]map[string]string)
	expectedUpdateFields := []string{"Name", "Email", "IsActive", "CreatedAt", "Metadata"}
	if len(updateFields) != len(expectedUpdateFields) {
		t.Errorf("Expected %d update fields, got %d", len(expectedUpdateFields), len(updateFields))
	}

	// Check parameter placeholders are sequential
	insertPlaceholders := data["InsertPlaceholders"].(string)
	expectedPlaceholders := "$1, $2, $3"
	if insertPlaceholders != expectedPlaceholders {
		t.Errorf("Expected placeholders '%s', got '%s'", expectedPlaceholders, insertPlaceholders)
	}
}

func TestCodeGenerator_generateCRUDOperations(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())
	table := getTestTable()

	code, err := cg.generateCRUDOperations(table)
	if err != nil {
		t.Fatalf("generateCRUDOperations failed: %v", err)
	}

	// Check that all CRUD operations are present
	expectedOperations := []string{
		"GetByID(ctx context.Context, id uuid.UUID) (*Users, error)",
		"Create(ctx context.Context, params CreateUsersParams) (*Users, error)",
		"Update(ctx context.Context, id uuid.UUID, params UpdateUsersParams) (*Users, error)",
		"Delete(ctx context.Context, id uuid.UUID) error",
		"List(ctx context.Context) ([]Users, error)",
	}

	for _, op := range expectedOperations {
		if !strings.Contains(code, op) {
			t.Errorf("Generated CRUD operations missing: %s", op)
		}
	}

	// Check parameter structs
	if !strings.Contains(code, "type CreateUsersParams struct") {
		t.Error("Missing CreateUsersParams struct")
	}

	if !strings.Contains(code, "type UpdateUsersParams struct") {
		t.Error("Missing UpdateUsersParams struct")
	}

	// Check SQL queries
	expectedQueries := []string{
		"SELECT id, name, email, is_active, created_at, metadata",
		"INSERT INTO users",
		"UPDATE users",
		"DELETE FROM users",
		"ORDER BY id ASC",
	}

	for _, query := range expectedQueries {
		if !strings.Contains(code, query) {
			t.Errorf("Generated CRUD operations missing SQL: %s", query)
		}
	}
}

func TestCodeGenerator_generateTableCode(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())
	table := getTestTable()

	code, err := cg.generateTableCode(table)
	if err != nil {
		t.Fatalf("generateTableCode failed: %v", err)
	}

	// Check header
	if !strings.Contains(code, "// Code generated by dbutil-gen. DO NOT EDIT.") {
		t.Error("Missing generated code header")
	}

	if !strings.Contains(code, "// Source: table users") {
		t.Error("Missing source table comment")
	}

	// Check package declaration
	if !strings.Contains(code, "package repositories") {
		t.Error("Missing package declaration")
	}

	// Check imports
	expectedImports := []string{
		"context",
		"github.com/google/uuid",
		"github.com/jackc/pgx/v5/pgtype",
		"github.com/jackc/pgx/v5/pgxpool",
	}

	for _, imp := range expectedImports {
		if !strings.Contains(code, imp) {
			t.Errorf("Missing import: %s", imp)
		}
	}

	// Check that all major components are present
	majorComponents := []string{
		"type Users struct",
		"type UsersRepository struct",
		"func NewUsersRepository",
		"func (r *UsersRepository) GetByID",
		"func (r *UsersRepository) Create",
		"func (r *UsersRepository) Update",
		"func (r *UsersRepository) Delete",
		"func (r *UsersRepository) List",
	}

	for _, component := range majorComponents {
		if !strings.Contains(code, component) {
			t.Errorf("Missing component: %s", component)
		}
	}
}

func TestCodeGenerator_writeCodeToFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &Config{
		OutputDir:   tempDir,
		PackageName: "test",
		Verbose:     false,
	}

	cg := NewCodeGenerator(config)

	// Test code to write
	testCode := `package test

import "fmt"

type TestStruct struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

func (t *TestStruct) String() string {
	return fmt.Sprintf("TestStruct{ID: %d, Name: %s}", t.ID, t.Name)
}
`

	filename := filepath.Join(tempDir, "test_generated.go")

	// Write the file
	err := cg.writeCodeToFile(filename, testCode)
	if err != nil {
		t.Fatalf("writeCodeToFile failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Read the file back
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Check that content is properly formatted
	contentStr := string(content)
	if !strings.Contains(contentStr, "package test") {
		t.Error("File content missing package declaration")
	}

	if !strings.Contains(contentStr, "type TestStruct struct") {
		t.Error("File content missing struct definition")
	}

	// Check that the code is properly formatted (should have proper indentation)
	lines := strings.Split(contentStr, "\n")
	for i, line := range lines {
		if strings.Contains(line, "ID   int") {
			// Check that struct fields are properly indented
			if !strings.HasPrefix(line, "\t") {
				t.Errorf("Line %d not properly indented: %s", i+1, line)
			}
		}
	}
}

func TestCodeGenerator_writeCodeToFile_InvalidCode(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		OutputDir:   tempDir,
		PackageName: "test",
		Verbose:     false,
	}

	cg := NewCodeGenerator(config)

	// Invalid Go code that won't format
	invalidCode := `package test
	func invalid syntax here {
		this is not valid go code
	`

	filename := filepath.Join(tempDir, "invalid.go")

	// Should fail with formatting error
	err := cg.writeCodeToFile(filename, invalidCode)
	if err == nil {
		t.Fatal("Expected error for invalid code, got nil")
	}

	if !strings.Contains(err.Error(), "failed to format generated code") {
		t.Errorf("Expected formatting error, got: %v", err)
	}
}

func TestCodeGenerator_combineImports(t *testing.T) {
	cg := NewCodeGenerator(getTestConfig())

	list1 := []string{"context", "fmt"}
	list2 := []string{"fmt", "github.com/jackc/pgx/v5/pgtype", "context"}
	list3 := []string{"github.com/google/uuid"}

	combined := cg.combineImports(list1, list2, list3)

	// Check that duplicates are removed
	expected := []string{"context", "fmt", "github.com/jackc/pgx/v5/pgtype", "github.com/google/uuid"}
	if len(combined) != len(expected) {
		t.Errorf("Expected %d imports, got %d", len(expected), len(combined))
	}

	// Check that all expected imports are present
	for _, exp := range expected {
		found := false
		for _, imp := range combined {
			if imp == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected import: %s", exp)
		}
	}
}

func TestCodeGenerator_GenerateTableRepository_Integration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &Config{
		OutputDir:   tempDir,
		PackageName: "repositories",
		Verbose:     false,
	}

	cg := NewCodeGenerator(config)
	table := getTestTable()

	// Generate the repository
	err := cg.GenerateTableRepository(table)
	if err != nil {
		t.Fatalf("GenerateTableRepository failed: %v", err)
	}

	// Check that file was created with correct name
	expectedFilename := filepath.Join(tempDir, "users_generated.go")
	if _, err := os.Stat(expectedFilename); os.IsNotExist(err) {
		t.Fatal("Generated file does not exist")
	}

	// Read and validate the generated file
	content, err := os.ReadFile(expectedFilename)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Validate key components are present
	requiredComponents := []string{
		"package repositories",
		"type Users struct",
		"type UsersRepository struct",
		"func NewUsersRepository",
		"func (r *UsersRepository) GetByID",
		"func (r *UsersRepository) Create",
		"func (r *UsersRepository) Update",
		"func (r *UsersRepository) Delete",
		"func (r *UsersRepository) List",
		"func (u Users) GetID() uuid.UUID",
	}

	for _, component := range requiredComponents {
		if !strings.Contains(contentStr, component) {
			t.Errorf("Generated file missing component: %s", component)
		}
	}
}
