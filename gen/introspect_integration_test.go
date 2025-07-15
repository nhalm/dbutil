package gen

import (
	"context"
	"testing"
)

// getTestDB is now in test_helpers.go

func TestIntrospector_GetTables_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// We should have at least the tables from our test schema
	expectedTables := []string{
		"users",
		"profiles",
		"posts",
		"comments",
		"categories",
		"post_categories",
		"files",
		"data_types_test",
		"invalid_pk_table",
		"composite_pk_table",
	}

	if len(tables) < len(expectedTables) {
		t.Errorf("Expected at least %d tables, got %d", len(expectedTables), len(tables))
	}

	// Create a map for easy lookup
	tableMap := make(map[string]Table)
	for _, table := range tables {
		tableMap[table.Name] = table
	}

	// Verify specific tables exist
	for _, expectedTable := range expectedTables {
		if _, exists := tableMap[expectedTable]; !exists {
			t.Errorf("Expected table %s not found", expectedTable)
		}
	}
}

func TestIntrospector_GetTables_UsersTable_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Find the users table
	var usersTable *Table
	for _, table := range tables {
		if table.Name == "users" {
			usersTable = &table
			break
		}
	}

	if usersTable == nil {
		t.Fatal("Users table not found")
	}

	// Verify table structure
	if usersTable.Schema != "public" {
		t.Errorf("Users table schema = %v, want public", usersTable.Schema)
	}

	// Verify primary key
	if len(usersTable.PrimaryKey) != 1 {
		t.Errorf("Users table primary key length = %v, want 1", len(usersTable.PrimaryKey))
	} else if usersTable.PrimaryKey[0] != "id" {
		t.Errorf("Users table primary key = %v, want [id]", usersTable.PrimaryKey)
	}

	// Verify expected columns exist
	expectedColumns := map[string]struct{}{
		"id":                  {},
		"name":                {},
		"email":               {},
		"password_hash":       {},
		"is_active":           {},
		"created_at":          {},
		"updated_at":          {},
		"last_login":          {},
		"metadata":            {},
		"age":                 {},
		"balance":             {},
		"profile_picture_url": {},
	}

	if len(usersTable.Columns) < len(expectedColumns) {
		t.Errorf("Users table has %d columns, expected at least %d", len(usersTable.Columns), len(expectedColumns))
	}

	// Verify specific columns
	columnMap := make(map[string]Column)
	for _, col := range usersTable.Columns {
		columnMap[col.Name] = col
	}

	// Check ID column
	if idCol, exists := columnMap["id"]; exists {
		if idCol.Type != "uuid" {
			t.Errorf("ID column type = %v, want uuid", idCol.Type)
		}
		if idCol.IsNullable {
			t.Errorf("ID column should not be nullable")
		}
	} else {
		t.Error("ID column not found")
	}

	// Check email column
	if emailCol, exists := columnMap["email"]; exists {
		if !emailCol.IsString() {
			t.Errorf("Email column should be string type, got %v", emailCol.Type)
		}
		if emailCol.IsNullable {
			t.Errorf("Email column should not be nullable")
		}
	} else {
		t.Error("Email column not found")
	}

	// Check nullable column
	if lastLoginCol, exists := columnMap["last_login"]; exists {
		if !lastLoginCol.IsNullable {
			t.Errorf("Last login column should be nullable")
		}
		if !lastLoginCol.IsTimestamp() {
			t.Errorf("Last login column should be timestamp type, got %v", lastLoginCol.Type)
		}
	} else {
		t.Error("Last login column not found")
	}

	// Check JSONB column
	if metadataCol, exists := columnMap["metadata"]; exists {
		if metadataCol.Type != "jsonb" {
			t.Errorf("Metadata column type = %v, want jsonb", metadataCol.Type)
		}
		if !metadataCol.IsNullable {
			t.Errorf("Metadata column should be nullable")
		}
	} else {
		t.Error("Metadata column not found")
	}
}

func TestIntrospector_GetTables_DataTypesTest_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Find the data_types_test table
	var dataTypesTable *Table
	for _, table := range tables {
		if table.Name == "data_types_test" {
			dataTypesTable = &table
			break
		}
	}

	if dataTypesTable == nil {
		t.Fatal("data_types_test table not found")
	}

	// Create column map for easy lookup
	columnMap := make(map[string]Column)
	for _, col := range dataTypesTable.Columns {
		columnMap[col.Name] = col
	}

	// Test various PostgreSQL data types
	testCases := []struct {
		columnName   string
		expectedType string
		shouldExist  bool
	}{
		{"id", "uuid", true},
		{"text_field", "text", true},
		{"varchar_field", "varchar", true},
		{"char_field", "character", true},
		{"smallint_field", "smallint", true},
		{"integer_field", "integer", true},
		{"bigint_field", "bigint", true},
		{"decimal_field", "numeric", true},
		{"numeric_field", "numeric", true},
		{"real_field", "real", true},
		{"double_field", "double precision", true},
		{"boolean_field", "boolean", true},
		{"date_field", "date", true},
		{"time_field", "time without time zone", true},
		{"timestamp_field", "timestamp", true},
		{"timestamptz_field", "timestamptz", true},
		{"uuid_field", "uuid", true},
		{"json_field", "json", true},
		{"jsonb_field", "jsonb", true},
		{"bytea_field", "bytea", true},
		{"inet_field", "inet", true},
		{"cidr_field", "cidr", true},
		{"macaddr_field", "macaddr", true},
	}

	for _, tc := range testCases {
		t.Run(tc.columnName, func(t *testing.T) {
			col, exists := columnMap[tc.columnName]
			if !exists && tc.shouldExist {
				t.Errorf("Column %s not found", tc.columnName)
				return
			}
			if exists && col.Type != tc.expectedType {
				t.Errorf("Column %s type = %v, want %v", tc.columnName, col.Type, tc.expectedType)
			}
		})
	}

	// Test array columns
	arrayTestCases := []struct {
		columnName string
		baseType   string
	}{
		{"text_array_field", "text"},
		{"integer_array_field", "int4"},
		{"uuid_array_field", "uuid"},
	}

	for _, tc := range arrayTestCases {
		t.Run(tc.columnName+"_array", func(t *testing.T) {
			col, exists := columnMap[tc.columnName]
			if !exists {
				t.Errorf("Array column %s not found", tc.columnName)
				return
			}
			if !col.IsArray {
				t.Errorf("Column %s should be array type", tc.columnName)
			}
			if col.Type != tc.baseType {
				t.Errorf("Array column %s base type = %v, want %v", tc.columnName, col.Type, tc.baseType)
			}
		})
	}
}

func TestIntrospector_GetTables_PrimaryKeys_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Test primary key detection
	testCases := []struct {
		tableName          string
		expectedPrimaryKey []string
	}{
		{"users", []string{"id"}},
		{"profiles", []string{"id"}},
		{"posts", []string{"id"}},
		{"comments", []string{"id"}},
		{"categories", []string{"id"}},
		{"post_categories", []string{"id"}},
		{"files", []string{"id"}},
		{"data_types_test", []string{"id"}},
		{"invalid_pk_table", []string{"id"}},                     // This has a serial PK, not UUID
		{"composite_pk_table", []string{"tenant_id", "user_id"}}, // Composite PK
	}

	// Create table map
	tableMap := make(map[string]Table)
	for _, table := range tables {
		tableMap[table.Name] = table
	}

	for _, tc := range testCases {
		t.Run(tc.tableName, func(t *testing.T) {
			table, exists := tableMap[tc.tableName]
			if !exists {
				t.Errorf("Table %s not found", tc.tableName)
				return
			}

			if len(table.PrimaryKey) != len(tc.expectedPrimaryKey) {
				t.Errorf("Table %s primary key length = %v, want %v", tc.tableName, len(table.PrimaryKey), len(tc.expectedPrimaryKey))
				return
			}

			for i, expectedCol := range tc.expectedPrimaryKey {
				if table.PrimaryKey[i] != expectedCol {
					t.Errorf("Table %s primary key[%d] = %v, want %v", tc.tableName, i, table.PrimaryKey[i], expectedCol)
				}
			}
		})
	}
}

func TestIntrospector_GetTables_Indexes_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Find the users table (which should have indexes)
	var usersTable *Table
	for _, table := range tables {
		if table.Name == "users" {
			usersTable = &table
			break
		}
	}

	if usersTable == nil {
		t.Fatal("Users table not found")
	}

	// The users table should have at least some indexes
	if len(usersTable.Indexes) == 0 {
		t.Error("Users table should have indexes")
	}

	// Look for the email index
	var emailIndex *Index
	for _, index := range usersTable.Indexes {
		if index.Name == "idx_users_email" {
			emailIndex = &index
			break
		}
	}

	if emailIndex != nil {
		if len(emailIndex.Columns) != 1 || emailIndex.Columns[0] != "email" {
			t.Errorf("Email index columns = %v, want [email]", emailIndex.Columns)
		}
	}
}

func TestIntrospector_GetTables_Relationships_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Create table map
	tableMap := make(map[string]Table)
	for _, table := range tables {
		tableMap[table.Name] = table
	}

	// Test foreign key relationships by checking for foreign key columns
	testCases := []struct {
		tableName    string
		fkColumnName string
		expectedType string
		shouldBeUUID bool
	}{
		{"profiles", "user_id", "uuid", true},
		{"posts", "user_id", "uuid", true},
		{"comments", "post_id", "uuid", true},
		{"comments", "user_id", "uuid", true},
		{"comments", "parent_id", "uuid", true}, // Self-referencing FK
		{"post_categories", "post_id", "uuid", true},
		{"post_categories", "category_id", "uuid", true},
		{"files", "user_id", "uuid", true},
	}

	for _, tc := range testCases {
		t.Run(tc.tableName+"_"+tc.fkColumnName, func(t *testing.T) {
			table, exists := tableMap[tc.tableName]
			if !exists {
				t.Errorf("Table %s not found", tc.tableName)
				return
			}

			// Find the foreign key column
			var fkColumn *Column
			for _, col := range table.Columns {
				if col.Name == tc.fkColumnName {
					fkColumn = &col
					break
				}
			}

			if fkColumn == nil {
				t.Errorf("Foreign key column %s not found in table %s", tc.fkColumnName, tc.tableName)
				return
			}

			if fkColumn.Type != tc.expectedType {
				t.Errorf("FK column %s.%s type = %v, want %v", tc.tableName, tc.fkColumnName, fkColumn.Type, tc.expectedType)
			}

			if tc.shouldBeUUID && !fkColumn.IsUUID() {
				t.Errorf("FK column %s.%s should be UUID type", tc.tableName, tc.fkColumnName)
			}
		})
	}
}

func TestIntrospector_GetTables_InvalidPrimaryKeys_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Find tables with invalid primary keys (for testing UUID validation)
	var invalidPKTable *Table
	for _, table := range tables {
		if table.Name == "invalid_pk_table" {
			invalidPKTable = &table
			break
		}
	}

	if invalidPKTable != nil {
		// This table should have a serial primary key, not UUID
		if len(invalidPKTable.PrimaryKey) == 1 {
			pkCol := invalidPKTable.GetPrimaryKeyColumn()
			if pkCol != nil && pkCol.IsUUID() {
				t.Error("invalid_pk_table should not have UUID primary key")
			}
		}
	}
}

func TestIntrospector_TypeMapping_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	// Test that type mapping works correctly with real database types
	typeMapper := NewTypeMapper(nil)

	for _, table := range tables {
		err := typeMapper.MapTableColumns(&table)
		if err != nil {
			t.Errorf("Failed to map types for table %s: %v", table.Name, err)
			continue
		}

		// Verify all columns have Go types assigned
		for _, col := range table.Columns {
			if col.GoType == "" {
				t.Errorf("Column %s.%s has no Go type assigned", table.Name, col.Name)
			}
		}
	}
}

func TestIntrospector_UUIDValidation_Integration(t *testing.T) {
	pool := getTestDB(t)
	defer pool.Close()

	introspector := NewIntrospector(pool, "public")
	ctx := context.Background()

	tables, err := introspector.GetTables(ctx)
	if err != nil {
		t.Fatalf("GetTables() error = %v", err)
	}

	typeMapper := NewTypeMapper(nil)

	// Test UUID validation on all tables
	for _, table := range tables {
		pkCol := table.GetPrimaryKeyColumn()
		if pkCol == nil {
			continue // Skip tables without single-column primary keys
		}

		err := typeMapper.ValidateUUIDPrimaryKey(pkCol)

		// Tables with UUID primary keys should pass validation
		if pkCol.IsUUID() && err != nil {
			t.Errorf("Table %s UUID primary key validation failed: %v", table.Name, err)
		}

		// Tables with non-UUID primary keys should fail validation
		if !pkCol.IsUUID() && err == nil {
			t.Errorf("Table %s non-UUID primary key should fail validation", table.Name)
		}
	}
}
