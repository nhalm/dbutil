package gen

import (
	"strings"
	"testing"
)

func TestEdgeCases_EmptyTables(t *testing.T) {
	// Test handling of empty table structures
	table := Table{
		Name:   "empty_table",
		Schema: "public",
	}

	// Test empty columns
	if len(table.Columns) != 0 {
		t.Errorf("Empty table should have no columns, got %d", len(table.Columns))
	}

	// Test empty primary key
	if len(table.PrimaryKey) != 0 {
		t.Errorf("Empty table should have no primary key, got %v", table.PrimaryKey)
	}

	// Test empty indexes
	if len(table.Indexes) != 0 {
		t.Errorf("Empty table should have no indexes, got %d", len(table.Indexes))
	}

	// Test GetColumn with empty columns
	col := table.GetColumn("nonexistent")
	if col != nil {
		t.Errorf("GetColumn on empty table should return nil, got %v", col)
	}

	// Test GetPrimaryKeyColumn with empty primary key
	pkCol := table.GetPrimaryKeyColumn()
	if pkCol != nil {
		t.Errorf("GetPrimaryKeyColumn on empty table should return nil, got %v", pkCol)
	}
}

func TestEdgeCases_SpecialCharacters(t *testing.T) {
	// Test handling of special characters in names
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "underscore_case",
			input:    "user_profile",
			expected: "UserProfile",
		},
		{
			name:     "multiple_underscores",
			input:    "user_profile_data",
			expected: "UserProfileData",
		},
		{
			name:     "leading_underscore",
			input:    "_private_field",
			expected: "PrivateField",
		},
		{
			name:     "trailing_underscore",
			input:    "field_name_",
			expected: "FieldName",
		},
		{
			name:     "multiple_consecutive_underscores",
			input:    "user__profile",
			expected: "UserProfile",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "single_character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "already_pascal_case",
			input:    "UserProfile",
			expected: "UserProfile",
		},
		{
			name:     "mixed_case",
			input:    "userId",
			expected: "UserId",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			table := Table{Name: tc.input}
			result := table.GoStructName()
			if result != tc.expected {
				t.Errorf("GoStructName(%s) = %s, want %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestEdgeCases_LongNames(t *testing.T) {
	// Test handling of very long table and column names
	longName := strings.Repeat("very_long_name_", 10) + "end"

	table := Table{
		Name:   longName,
		Schema: "public",
	}

	// Test that long names are handled properly
	structName := table.GoStructName()
	if structName == "" {
		t.Error("Long table name should produce non-empty struct name")
	}

	// Test that the result is valid Go identifier (starts with uppercase)
	if len(structName) > 0 && structName[0] < 'A' || structName[0] > 'Z' {
		t.Errorf("Struct name should start with uppercase letter, got %s", structName)
	}

	// Test filename generation
	fileName := table.GoFileName()
	if fileName == "" {
		t.Error("Long table name should produce non-empty filename")
	}

	if !strings.HasSuffix(fileName, "_generated.go") {
		t.Errorf("Filename should end with _generated.go, got %s", fileName)
	}
}

func TestEdgeCases_NullableArrays(t *testing.T) {
	typeMapper := NewTypeMapper(nil)

	// Test nullable array types
	testCases := []struct {
		name         string
		pgType       string
		isNullable   bool
		isArray      bool
		expectedType string
		expectError  bool
	}{
		{
			name:         "nullable_text_array",
			pgType:       "text",
			isNullable:   true,
			isArray:      true,
			expectedType: "[]pgtype.Text",
			expectError:  false,
		},
		{
			name:         "nullable_uuid_array",
			pgType:       "uuid",
			isNullable:   true,
			isArray:      true,
			expectedType: "[]pgtype.UUID",
			expectError:  false,
		},
		{
			name:         "non_nullable_text_array",
			pgType:       "text",
			isNullable:   false,
			isArray:      true,
			expectedType: "[]string",
			expectError:  false,
		},
		{
			name:         "nullable_non_array_text",
			pgType:       "text",
			isNullable:   true,
			isArray:      false,
			expectedType: "pgtype.Text",
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goType, err := typeMapper.MapType(tc.pgType, tc.isNullable, tc.isArray)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
				if goType != tc.expectedType {
					t.Errorf("MapType(%s, %v, %v) = %s, want %s",
						tc.pgType, tc.isNullable, tc.isArray, goType, tc.expectedType)
				}
			}
		})
	}
}

func TestEdgeCases_UnsupportedTypes(t *testing.T) {
	typeMapper := NewTypeMapper(nil)

	// Test unsupported PostgreSQL types
	unsupportedTypes := []string{
		"unknown_type",
		"custom_enum",
		"pg_lsn",
		"txid_snapshot",
		"tsvector",
		"tsquery",
	}

	for _, pgType := range unsupportedTypes {
		t.Run(pgType, func(t *testing.T) {
			_, err := typeMapper.MapType(pgType, false, false)
			if err == nil {
				t.Errorf("Expected error for unsupported type %s", pgType)
			}

			// Verify error message is helpful
			if !strings.Contains(err.Error(), "unsupported") {
				t.Errorf("Error message should mention 'unsupported', got: %s", err.Error())
			}
		})
	}
}

func TestEdgeCases_CustomTypeMappings(t *testing.T) {
	// Test custom type mappings
	customMappings := map[string]string{
		"custom_type": "CustomStruct",
		"enum_type":   "EnumType",
	}

	typeMapper := NewTypeMapper(customMappings)

	testCases := []struct {
		name         string
		pgType       string
		isNullable   bool
		isArray      bool
		expectedType string
		expectError  bool
	}{
		{
			name:         "custom_type",
			pgType:       "custom_type",
			isNullable:   false,
			isArray:      false,
			expectedType: "CustomStruct",
			expectError:  false,
		},
		{
			name:         "nullable_custom_type",
			pgType:       "custom_type",
			isNullable:   true,
			isArray:      false,
			expectedType: "*CustomStruct",
			expectError:  false,
		},
		{
			name:         "custom_type_array",
			pgType:       "custom_type",
			isNullable:   false,
			isArray:      true,
			expectedType: "[]CustomStruct",
			expectError:  false,
		},
		{
			name:         "nullable_custom_type_array",
			pgType:       "custom_type",
			isNullable:   true,
			isArray:      true,
			expectedType: "[]*CustomStruct",
			expectError:  false,
		},
		{
			name:         "override_builtin_type",
			pgType:       "text",
			isNullable:   false,
			isArray:      false,
			expectedType: "string", // Should still use built-in mapping
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goType, err := typeMapper.MapType(tc.pgType, tc.isNullable, tc.isArray)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
				if goType != tc.expectedType {
					t.Errorf("MapType(%s, %v, %v) = %s, want %s",
						tc.pgType, tc.isNullable, tc.isArray, goType, tc.expectedType)
				}
			}
		})
	}
}

func TestEdgeCases_ImportGeneration(t *testing.T) {
	typeMapper := NewTypeMapper(nil)

	// Test import generation for edge cases
	testCases := []struct {
		name            string
		columns         []Column
		expectedImports []string
	}{
		{
			name:            "no_columns",
			columns:         []Column{},
			expectedImports: []string{},
		},
		{
			name: "only_basic_types",
			columns: []Column{
				{Type: "text", IsNullable: false, IsArray: false},
				{Type: "integer", IsNullable: false, IsArray: false},
				{Type: "boolean", IsNullable: false, IsArray: false},
			},
			expectedImports: []string{},
		},
		{
			name: "mixed_imports",
			columns: []Column{
				{Type: "uuid", IsNullable: false, IsArray: false},
				{Type: "text", IsNullable: true, IsArray: false},
				{Type: "timestamp", IsNullable: false, IsArray: false},
				{Type: "json", IsNullable: false, IsArray: false},
			},
			expectedImports: []string{
				"encoding/json",
				"github.com/google/uuid",
				"github.com/jackc/pgx/v5/pgtype",
				"time",
			},
		},
		{
			name: "duplicate_imports",
			columns: []Column{
				{Type: "uuid", IsNullable: false, IsArray: false},
				{Type: "uuid", IsNullable: true, IsArray: false},
				{Type: "uuid", IsNullable: false, IsArray: true},
			},
			expectedImports: []string{
				"github.com/google/uuid",
				"github.com/jackc/pgx/v5/pgtype",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imports := typeMapper.GetRequiredImports(tc.columns)

			if len(imports) != len(tc.expectedImports) {
				t.Errorf("Expected %d imports, got %d: %v",
					len(tc.expectedImports), len(imports), imports)
			}

			// Convert to maps for easier comparison
			importMap := make(map[string]bool)
			for _, imp := range imports {
				importMap[imp] = true
			}

			expectedMap := make(map[string]bool)
			for _, imp := range tc.expectedImports {
				expectedMap[imp] = true
			}

			for expectedImport := range expectedMap {
				if !importMap[expectedImport] {
					t.Errorf("Missing expected import: %s", expectedImport)
				}
			}

			for actualImport := range importMap {
				if !expectedMap[actualImport] {
					t.Errorf("Unexpected import: %s", actualImport)
				}
			}
		})
	}
}

func TestEdgeCases_TableValidation(t *testing.T) {
	// Test table validation edge cases
	testCases := []struct {
		name        string
		table       Table
		expectValid bool
	}{
		{
			name: "valid_table",
			table: Table{
				Name:   "users",
				Schema: "public",
				Columns: []Column{
					{Name: "id", Type: "uuid", IsNullable: false},
					{Name: "name", Type: "text", IsNullable: false},
				},
				PrimaryKey: []string{"id"},
			},
			expectValid: true,
		},
		{
			name: "empty_table_name",
			table: Table{
				Name:   "",
				Schema: "public",
			},
			expectValid: false,
		},
		{
			name: "empty_schema",
			table: Table{
				Name:   "users",
				Schema: "",
			},
			expectValid: false,
		},
		{
			name: "table_with_no_columns",
			table: Table{
				Name:    "empty_table",
				Schema:  "public",
				Columns: []Column{},
			},
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.table.Name != "" && tc.table.Schema != "" && len(tc.table.Columns) > 0

			if isValid != tc.expectValid {
				t.Errorf("Table validation: expected %v, got %v", tc.expectValid, isValid)
			}
		})
	}
}

func TestEdgeCases_ColumnHelperMethods(t *testing.T) {
	// Test column helper methods with edge cases
	testCases := []struct {
		name     string
		column   Column
		isUUID   bool
		isString bool
		isTime   bool
	}{
		{
			name:     "uuid_column",
			column:   Column{Type: "uuid"},
			isUUID:   true,
			isString: false,
			isTime:   false,
		},
		{
			name:     "uppercase_uuid",
			column:   Column{Type: "UUID"},
			isUUID:   true,
			isString: false,
			isTime:   false,
		},
		{
			name:     "text_column",
			column:   Column{Type: "text"},
			isUUID:   false,
			isString: true,
			isTime:   false,
		},
		{
			name:     "varchar_column",
			column:   Column{Type: "varchar"},
			isUUID:   false,
			isString: true,
			isTime:   false,
		},
		{
			name:     "character_column",
			column:   Column{Type: "character"},
			isUUID:   false,
			isString: true,
			isTime:   false,
		},
		{
			name:     "timestamp_column",
			column:   Column{Type: "timestamp"},
			isUUID:   false,
			isString: false,
			isTime:   true,
		},
		{
			name:     "timestamptz_column",
			column:   Column{Type: "timestamptz"},
			isUUID:   false,
			isString: false,
			isTime:   true,
		},
		{
			name:     "date_column",
			column:   Column{Type: "date"},
			isUUID:   false,
			isString: false,
			isTime:   true,
		},
		{
			name:     "time_column",
			column:   Column{Type: "time"},
			isUUID:   false,
			isString: false,
			isTime:   true,
		},
		{
			name:     "integer_column",
			column:   Column{Type: "integer"},
			isUUID:   false,
			isString: false,
			isTime:   false,
		},
		{
			name:     "empty_type",
			column:   Column{Type: ""},
			isUUID:   false,
			isString: false,
			isTime:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.column.IsUUID() != tc.isUUID {
				t.Errorf("IsUUID() = %v, want %v", tc.column.IsUUID(), tc.isUUID)
			}
			if tc.column.IsString() != tc.isString {
				t.Errorf("IsString() = %v, want %v", tc.column.IsString(), tc.isString)
			}
			if tc.column.IsTimestamp() != tc.isTime {
				t.Errorf("IsTimestamp() = %v, want %v", tc.column.IsTimestamp(), tc.isTime)
			}
		})
	}
}

func TestEdgeCases_IndexParsing(t *testing.T) {
	introspector := &Introspector{}

	// Test edge cases in index parsing
	testCases := []struct {
		name         string
		indexDef     string
		expectedCols []string
	}{
		{
			name:         "no_parentheses",
			indexDef:     "CREATE INDEX idx_test",
			expectedCols: []string{},
		},
		{
			name:         "empty_parentheses",
			indexDef:     "CREATE INDEX idx_test ON table ()",
			expectedCols: []string{},
		},
		{
			name:         "malformed_parentheses",
			indexDef:     "CREATE INDEX idx_test ON table (col1",
			expectedCols: []string{},
		},
		{
			name:         "nested_parentheses",
			indexDef:     "CREATE INDEX idx_test ON table (func(col1), col2)",
			expectedCols: []string{"func(col1)", "col2"},
		},
		{
			name:         "whitespace_only",
			indexDef:     "CREATE INDEX idx_test ON table (   )",
			expectedCols: []string{},
		},
		{
			name:         "trailing_comma",
			indexDef:     "CREATE INDEX idx_test ON table (col1, col2,)",
			expectedCols: []string{"col1", "col2"},
		},
		{
			name:         "leading_comma",
			indexDef:     "CREATE INDEX idx_test ON table (, col1, col2)",
			expectedCols: []string{"col1", "col2"},
		},
		{
			name:         "multiple_commas",
			indexDef:     "CREATE INDEX idx_test ON table (col1,, col2)",
			expectedCols: []string{"col1", "col2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cols := introspector.parseIndexColumns(tc.indexDef)

			if len(cols) != len(tc.expectedCols) {
				t.Errorf("Expected %d columns, got %d: %v",
					len(tc.expectedCols), len(cols), cols)
				return
			}

			for i, expectedCol := range tc.expectedCols {
				if cols[i] != expectedCol {
					t.Errorf("Column %d: expected %s, got %s", i, expectedCol, cols[i])
				}
			}
		})
	}
}

func TestEdgeCases_TypeMapperNilHandling(t *testing.T) {
	// Test type mapper with nil inputs
	typeMapper := NewTypeMapper(nil)

	t.Run("nil_column_slice", func(t *testing.T) {
		imports := typeMapper.GetRequiredImports(nil)
		if imports == nil {
			t.Error("GetRequiredImports should return empty slice, not nil")
		}
		if len(imports) != 0 {
			t.Errorf("Expected empty slice, got %v", imports)
		}
	})

	t.Run("nil_table_pointer", func(t *testing.T) {
		err := typeMapper.MapTableColumns(nil)
		if err == nil {
			t.Error("MapTableColumns with nil table should return error")
		}
	})

	t.Run("nil_column_pointer", func(t *testing.T) {
		err := typeMapper.ValidateUUIDPrimaryKey(nil)
		if err == nil {
			t.Error("ValidateUUIDPrimaryKey with nil column should return error")
		}
	})
}
