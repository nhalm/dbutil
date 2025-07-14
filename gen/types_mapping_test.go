package gen

import (
	"reflect"
	"sort"
	"testing"
)

func TestTypeMapper_MapType(t *testing.T) {
	tm := NewTypeMapper(nil)

	tests := []struct {
		name       string
		pgType     string
		isNullable bool
		isArray    bool
		want       string
		wantErr    bool
	}{
		// UUID types
		{
			name:       "uuid type",
			pgType:     "uuid",
			isNullable: false,
			isArray:    false,
			want:       "uuid.UUID",
			wantErr:    false,
		},
		{
			name:       "nullable uuid type",
			pgType:     "uuid",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.UUID",
			wantErr:    false,
		},
		{
			name:       "uuid array type",
			pgType:     "uuid",
			isNullable: false,
			isArray:    true,
			want:       "[]uuid.UUID",
			wantErr:    false,
		},
		{
			name:       "nullable uuid array type",
			pgType:     "uuid",
			isNullable: true,
			isArray:    true,
			want:       "[]pgtype.UUID",
			wantErr:    false,
		},

		// String types
		{
			name:       "text type",
			pgType:     "text",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "nullable text type",
			pgType:     "text",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Text",
			wantErr:    false,
		},
		{
			name:       "varchar type",
			pgType:     "varchar",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "character varying type",
			pgType:     "character varying",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "char type",
			pgType:     "char",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "character type",
			pgType:     "character",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},

		// Integer types
		{
			name:       "smallint type",
			pgType:     "smallint",
			isNullable: false,
			isArray:    false,
			want:       "int16",
			wantErr:    false,
		},
		{
			name:       "int2 type",
			pgType:     "int2",
			isNullable: false,
			isArray:    false,
			want:       "int16",
			wantErr:    false,
		},
		{
			name:       "nullable smallint type",
			pgType:     "smallint",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Int2",
			wantErr:    false,
		},
		{
			name:       "integer type",
			pgType:     "integer",
			isNullable: false,
			isArray:    false,
			want:       "int32",
			wantErr:    false,
		},
		{
			name:       "int type",
			pgType:     "int",
			isNullable: false,
			isArray:    false,
			want:       "int32",
			wantErr:    false,
		},
		{
			name:       "int4 type",
			pgType:     "int4",
			isNullable: false,
			isArray:    false,
			want:       "int32",
			wantErr:    false,
		},
		{
			name:       "nullable integer type",
			pgType:     "integer",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Int4",
			wantErr:    false,
		},
		{
			name:       "bigint type",
			pgType:     "bigint",
			isNullable: false,
			isArray:    false,
			want:       "int64",
			wantErr:    false,
		},
		{
			name:       "int8 type",
			pgType:     "int8",
			isNullable: false,
			isArray:    false,
			want:       "int64",
			wantErr:    false,
		},
		{
			name:       "nullable bigint type",
			pgType:     "bigint",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Int8",
			wantErr:    false,
		},

		// Floating point types
		{
			name:       "real type",
			pgType:     "real",
			isNullable: false,
			isArray:    false,
			want:       "float32",
			wantErr:    false,
		},
		{
			name:       "float4 type",
			pgType:     "float4",
			isNullable: false,
			isArray:    false,
			want:       "float32",
			wantErr:    false,
		},
		{
			name:       "nullable real type",
			pgType:     "real",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Float4",
			wantErr:    false,
		},
		{
			name:       "double precision type",
			pgType:     "double precision",
			isNullable: false,
			isArray:    false,
			want:       "float64",
			wantErr:    false,
		},
		{
			name:       "float8 type",
			pgType:     "float8",
			isNullable: false,
			isArray:    false,
			want:       "float64",
			wantErr:    false,
		},
		{
			name:       "nullable double precision type",
			pgType:     "double precision",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Float8",
			wantErr:    false,
		},
		{
			name:       "numeric type",
			pgType:     "numeric",
			isNullable: false,
			isArray:    false,
			want:       "float64",
			wantErr:    false,
		},
		{
			name:       "decimal type",
			pgType:     "decimal",
			isNullable: false,
			isArray:    false,
			want:       "float64",
			wantErr:    false,
		},

		// Boolean type
		{
			name:       "boolean type",
			pgType:     "boolean",
			isNullable: false,
			isArray:    false,
			want:       "bool",
			wantErr:    false,
		},
		{
			name:       "bool type",
			pgType:     "bool",
			isNullable: false,
			isArray:    false,
			want:       "bool",
			wantErr:    false,
		},
		{
			name:       "nullable boolean type",
			pgType:     "boolean",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Bool",
			wantErr:    false,
		},

		// Date/time types
		{
			name:       "date type",
			pgType:     "date",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "time type",
			pgType:     "time",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "time without time zone type",
			pgType:     "time without time zone",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "timetz type",
			pgType:     "timetz",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "time with time zone type",
			pgType:     "time with time zone",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "timestamp type",
			pgType:     "timestamp",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "timestamp without time zone type",
			pgType:     "timestamp without time zone",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "timestamptz type",
			pgType:     "timestamptz",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "timestamp with time zone type",
			pgType:     "timestamp with time zone",
			isNullable: false,
			isArray:    false,
			want:       "time.Time",
			wantErr:    false,
		},
		{
			name:       "nullable timestamp type",
			pgType:     "timestamp",
			isNullable: true,
			isArray:    false,
			want:       "pgtype.Timestamptz",
			wantErr:    false,
		},

		// Binary types
		{
			name:       "bytea type",
			pgType:     "bytea",
			isNullable: false,
			isArray:    false,
			want:       "[]byte",
			wantErr:    false,
		},
		{
			name:       "nullable bytea type",
			pgType:     "bytea",
			isNullable: true,
			isArray:    false,
			want:       "*[]byte",
			wantErr:    false,
		},

		// JSON types
		{
			name:       "json type",
			pgType:     "json",
			isNullable: false,
			isArray:    false,
			want:       "json.RawMessage",
			wantErr:    false,
		},
		{
			name:       "jsonb type",
			pgType:     "jsonb",
			isNullable: false,
			isArray:    false,
			want:       "json.RawMessage",
			wantErr:    false,
		},
		{
			name:       "nullable json type",
			pgType:     "json",
			isNullable: true,
			isArray:    false,
			want:       "*json.RawMessage",
			wantErr:    false,
		},

		// Network types
		{
			name:       "inet type",
			pgType:     "inet",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "cidr type",
			pgType:     "cidr",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "macaddr type",
			pgType:     "macaddr",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},

		// Geometric types
		{
			name:       "point type",
			pgType:     "point",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "line type",
			pgType:     "line",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "polygon type",
			pgType:     "polygon",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},

		// Range types
		{
			name:       "int4range type",
			pgType:     "int4range",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},
		{
			name:       "tsrange type",
			pgType:     "tsrange",
			isNullable: false,
			isArray:    false,
			want:       "string",
			wantErr:    false,
		},

		// Array types
		{
			name:       "text array type",
			pgType:     "text",
			isNullable: false,
			isArray:    true,
			want:       "[]string",
			wantErr:    false,
		},
		{
			name:       "integer array type",
			pgType:     "integer",
			isNullable: false,
			isArray:    true,
			want:       "[]int32",
			wantErr:    false,
		},
		{
			name:       "nullable text array type",
			pgType:     "text",
			isNullable: true,
			isArray:    true,
			want:       "[]pgtype.Text",
			wantErr:    false,
		},

		// Unsupported types
		{
			name:       "unsupported type",
			pgType:     "unsupported_type",
			isNullable: false,
			isArray:    false,
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.MapType(tt.pgType, tt.isNullable, tt.isArray)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MapType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeMapper_MapType_WithCustomMappings(t *testing.T) {
	customMappings := map[string]string{
		"custom_type": "MyCustomType",
		"uuid":        "MyUUID", // Override built-in mapping
	}
	tm := NewTypeMapper(customMappings)

	tests := []struct {
		name       string
		pgType     string
		isNullable bool
		isArray    bool
		want       string
		wantErr    bool
	}{
		{
			name:       "custom type mapping",
			pgType:     "custom_type",
			isNullable: false,
			isArray:    false,
			want:       "MyCustomType",
			wantErr:    false,
		},
		{
			name:       "override built-in mapping",
			pgType:     "uuid",
			isNullable: false,
			isArray:    false,
			want:       "MyUUID",
			wantErr:    false,
		},
		{
			name:       "nullable custom type",
			pgType:     "custom_type",
			isNullable: true,
			isArray:    false,
			want:       "*MyCustomType",
			wantErr:    false,
		},
		{
			name:       "array custom type",
			pgType:     "custom_type",
			isNullable: false,
			isArray:    true,
			want:       "[]MyCustomType",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.MapType(tt.pgType, tt.isNullable, tt.isArray)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MapType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeMapper_GetRequiredImports(t *testing.T) {
	tm := NewTypeMapper(nil)

	tests := []struct {
		name    string
		columns []Column
		want    []string
	}{
		{
			name: "basic types with imports",
			columns: []Column{
				{Type: "uuid", IsNullable: false, IsArray: false},
				{Type: "timestamp", IsNullable: false, IsArray: false},
				{Type: "json", IsNullable: false, IsArray: false},
				{Type: "text", IsNullable: true, IsArray: false},
			},
			want: []string{
				"github.com/google/uuid",
				"time",
				"encoding/json",
				"github.com/jackc/pgx/v5/pgtype",
			},
		},
		{
			name: "only basic types",
			columns: []Column{
				{Type: "text", IsNullable: false, IsArray: false},
				{Type: "integer", IsNullable: false, IsArray: false},
				{Type: "boolean", IsNullable: false, IsArray: false},
			},
			want: []string{},
		},
		{
			name: "array types",
			columns: []Column{
				{Type: "uuid", IsNullable: false, IsArray: true},
				{Type: "text", IsNullable: false, IsArray: true},
			},
			want: []string{
				"github.com/google/uuid",
			},
		},
		{
			name: "nullable types",
			columns: []Column{
				{Type: "integer", IsNullable: true, IsArray: false},
				{Type: "boolean", IsNullable: true, IsArray: false},
				{Type: "uuid", IsNullable: true, IsArray: false},
			},
			want: []string{
				"github.com/jackc/pgx/v5/pgtype",
			},
		},
		{
			name: "mixed types",
			columns: []Column{
				{Type: "uuid", IsNullable: false, IsArray: false},
				{Type: "uuid", IsNullable: true, IsArray: false},
				{Type: "text", IsNullable: false, IsArray: true},
				{Type: "timestamp", IsNullable: false, IsArray: false}, // Non-nullable timestamp for time.Time
				{Type: "json", IsNullable: false, IsArray: false},
			},
			want: []string{
				"github.com/google/uuid",
				"github.com/jackc/pgx/v5/pgtype",
				"time",
				"encoding/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Map the column types first
			for i := range tt.columns {
				goType, err := tm.MapType(tt.columns[i].Type, tt.columns[i].IsNullable, tt.columns[i].IsArray)
				if err != nil {
					t.Fatalf("Failed to map type for column %d: %v", i, err)
				}
				tt.columns[i].GoType = goType
			}

			got := tm.GetRequiredImports(tt.columns)

			// Sort both slices for comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRequiredImports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeMapper_MapTableColumns(t *testing.T) {
	tm := NewTypeMapper(nil)

	table := &Table{
		Name:   "test_table",
		Schema: "public",
		Columns: []Column{
			{Name: "id", Type: "uuid", IsNullable: false, IsArray: false},
			{Name: "name", Type: "text", IsNullable: true, IsArray: false},
			{Name: "age", Type: "integer", IsNullable: false, IsArray: false},
			{Name: "tags", Type: "text", IsNullable: false, IsArray: true},
			{Name: "created_at", Type: "timestamp", IsNullable: false, IsArray: false},
		},
	}

	err := tm.MapTableColumns(table)
	if err != nil {
		t.Fatalf("MapTableColumns() error = %v", err)
	}

	expectedGoTypes := []string{
		"uuid.UUID",
		"pgtype.Text",
		"int32",
		"[]string",
		"time.Time",
	}

	for i, col := range table.Columns {
		if col.GoType != expectedGoTypes[i] {
			t.Errorf("Column %s: GoType = %v, want %v", col.Name, col.GoType, expectedGoTypes[i])
		}
	}
}

func TestTypeMapper_MapTableColumns_WithError(t *testing.T) {
	tm := NewTypeMapper(nil)

	table := &Table{
		Name:   "test_table",
		Schema: "public",
		Columns: []Column{
			{Name: "id", Type: "uuid", IsNullable: false, IsArray: false},
			{Name: "invalid", Type: "unsupported_type", IsNullable: false, IsArray: false},
		},
	}

	err := tm.MapTableColumns(table)
	if err == nil {
		t.Error("MapTableColumns() expected error for unsupported type, got nil")
	}
}

func TestTypeMapper_ValidateUUIDPrimaryKey(t *testing.T) {
	tm := NewTypeMapper(nil)

	tests := []struct {
		name    string
		column  Column
		wantErr bool
	}{
		{
			name: "valid UUID primary key",
			column: Column{
				Name:       "id",
				Type:       "uuid",
				IsNullable: false,
				IsArray:    false,
			},
			wantErr: false,
		},
		{
			name: "UUID uppercase",
			column: Column{
				Name:       "id",
				Type:       "UUID",
				IsNullable: false,
				IsArray:    false,
			},
			wantErr: false,
		},
		{
			name: "non-UUID type",
			column: Column{
				Name:       "id",
				Type:       "integer",
				IsNullable: false,
				IsArray:    false,
			},
			wantErr: true,
		},
		{
			name: "nullable UUID",
			column: Column{
				Name:       "id",
				Type:       "uuid",
				IsNullable: true,
				IsArray:    false,
			},
			wantErr: true,
		},
		{
			name: "UUID array",
			column: Column{
				Name:       "id",
				Type:       "uuid",
				IsNullable: false,
				IsArray:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.ValidateUUIDPrimaryKey(&tt.column)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUIDPrimaryKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeMapper_makeNullable(t *testing.T) {
	tm := NewTypeMapper(nil)

	tests := []struct {
		name   string
		goType string
		want   string
	}{
		{
			name:   "string type",
			goType: "string",
			want:   "pgtype.Text",
		},
		{
			name:   "int16 type",
			goType: "int16",
			want:   "pgtype.Int2",
		},
		{
			name:   "int32 type",
			goType: "int32",
			want:   "pgtype.Int4",
		},
		{
			name:   "int64 type",
			goType: "int64",
			want:   "pgtype.Int8",
		},
		{
			name:   "float32 type",
			goType: "float32",
			want:   "pgtype.Float4",
		},
		{
			name:   "float64 type",
			goType: "float64",
			want:   "pgtype.Float8",
		},
		{
			name:   "bool type",
			goType: "bool",
			want:   "pgtype.Bool",
		},
		{
			name:   "time.Time type",
			goType: "time.Time",
			want:   "pgtype.Timestamptz",
		},
		{
			name:   "uuid.UUID type",
			goType: "uuid.UUID",
			want:   "pgtype.UUID",
		},
		{
			name:   "json.RawMessage type",
			goType: "json.RawMessage",
			want:   "*json.RawMessage",
		},
		{
			name:   "[]byte type",
			goType: "[]byte",
			want:   "*[]byte",
		},
		{
			name:   "array of strings",
			goType: "[]string",
			want:   "[]pgtype.Text",
		},
		{
			name:   "array of int32",
			goType: "[]int32",
			want:   "[]pgtype.Int4",
		},
		{
			name:   "custom type",
			goType: "MyCustomType",
			want:   "*MyCustomType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.makeNullable(tt.goType)
			if got != tt.want {
				t.Errorf("makeNullable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTypeMapper(t *testing.T) {
	tests := []struct {
		name           string
		customMappings map[string]string
		want           *TypeMapper
	}{
		{
			name:           "nil custom mappings",
			customMappings: nil,
			want:           &TypeMapper{customMappings: nil},
		},
		{
			name:           "empty custom mappings",
			customMappings: map[string]string{},
			want:           &TypeMapper{customMappings: map[string]string{}},
		},
		{
			name: "with custom mappings",
			customMappings: map[string]string{
				"custom_type": "MyType",
			},
			want: &TypeMapper{
				customMappings: map[string]string{
					"custom_type": "MyType",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTypeMapper(tt.customMappings)
			if !reflect.DeepEqual(got.customMappings, tt.want.customMappings) {
				t.Errorf("NewTypeMapper() = %v, want %v", got.customMappings, tt.want.customMappings)
			}
		})
	}
}
