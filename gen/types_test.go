package gen

import (
	"testing"
)

func TestTable_GetColumn(t *testing.T) {
	table := Table{
		Name:   "users",
		Schema: "public",
		Columns: []Column{
			{Name: "id", Type: "uuid", GoType: "uuid.UUID"},
			{Name: "name", Type: "text", GoType: "string"},
			{Name: "email", Type: "varchar", GoType: "string"},
		},
	}

	tests := []struct {
		name       string
		columnName string
		want       *Column
	}{
		{
			name:       "existing column",
			columnName: "name",
			want:       &Column{Name: "name", Type: "text", GoType: "string"},
		},
		{
			name:       "non-existing column",
			columnName: "nonexistent",
			want:       nil,
		},
		{
			name:       "first column",
			columnName: "id",
			want:       &Column{Name: "id", Type: "uuid", GoType: "uuid.UUID"},
		},
		{
			name:       "last column",
			columnName: "email",
			want:       &Column{Name: "email", Type: "varchar", GoType: "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := table.GetColumn(tt.columnName)
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("GetColumn() = %v, want %v", got, tt.want)
				return
			}
			if got.Name != tt.want.Name || got.Type != tt.want.Type || got.GoType != tt.want.GoType {
				t.Errorf("GetColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GetPrimaryKeyColumn(t *testing.T) {
	tests := []struct {
		name  string
		table Table
		want  *Column
	}{
		{
			name: "single primary key",
			table: Table{
				Name:       "users",
				PrimaryKey: []string{"id"},
				Columns: []Column{
					{Name: "id", Type: "uuid", GoType: "uuid.UUID"},
					{Name: "name", Type: "text", GoType: "string"},
				},
			},
			want: &Column{Name: "id", Type: "uuid", GoType: "uuid.UUID"},
		},
		{
			name: "composite primary key",
			table: Table{
				Name:       "user_roles",
				PrimaryKey: []string{"user_id", "role_id"},
				Columns: []Column{
					{Name: "user_id", Type: "uuid", GoType: "uuid.UUID"},
					{Name: "role_id", Type: "uuid", GoType: "uuid.UUID"},
				},
			},
			want: nil,
		},
		{
			name: "no primary key",
			table: Table{
				Name:       "logs",
				PrimaryKey: []string{},
				Columns: []Column{
					{Name: "message", Type: "text", GoType: "string"},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.table.GetPrimaryKeyColumn()
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("GetPrimaryKeyColumn() = %v, want %v", got, tt.want)
				return
			}
			if got.Name != tt.want.Name || got.Type != tt.want.Type {
				t.Errorf("GetPrimaryKeyColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GoStructName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      string
	}{
		{
			name:      "simple table name",
			tableName: "users",
			want:      "Users",
		},
		{
			name:      "snake_case table name",
			tableName: "user_profiles",
			want:      "UserProfiles",
		},
		{
			name:      "multiple underscores",
			tableName: "user_profile_settings",
			want:      "UserProfileSettings",
		},
		{
			name:      "single character",
			tableName: "a",
			want:      "A",
		},
		{
			name:      "empty string",
			tableName: "",
			want:      "",
		},
		{
			name:      "already camelCase",
			tableName: "userProfiles",
			want:      "UserProfiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := Table{Name: tt.tableName}
			got := table.GoStructName()
			if got != tt.want {
				t.Errorf("GoStructName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GoFileName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      string
	}{
		{
			name:      "simple table name",
			tableName: "users",
			want:      "users_generated.go",
		},
		{
			name:      "snake_case table name",
			tableName: "user_profiles",
			want:      "user_profiles_generated.go",
		},
		{
			name:      "PascalCase table name",
			tableName: "UserProfiles",
			want:      "user_profiles_generated.go",
		},
		{
			name:      "camelCase table name",
			tableName: "userProfiles",
			want:      "user_profiles_generated.go",
		},
		{
			name:      "mixed case",
			tableName: "UserProfileSettings",
			want:      "user_profile_settings_generated.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := Table{Name: tt.tableName}
			got := table.GoFileName()
			if got != tt.want {
				t.Errorf("GoFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_IsUUID(t *testing.T) {
	tests := []struct {
		name   string
		column Column
		want   bool
	}{
		{
			name:   "uuid type",
			column: Column{Type: "uuid"},
			want:   true,
		},
		{
			name:   "UUID uppercase",
			column: Column{Type: "UUID"},
			want:   true,
		},
		{
			name:   "text type",
			column: Column{Type: "text"},
			want:   false,
		},
		{
			name:   "integer type",
			column: Column{Type: "integer"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.IsUUID()
			if got != tt.want {
				t.Errorf("IsUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_IsString(t *testing.T) {
	tests := []struct {
		name   string
		column Column
		want   bool
	}{
		{
			name:   "text type",
			column: Column{Type: "text"},
			want:   true,
		},
		{
			name:   "varchar type",
			column: Column{Type: "varchar"},
			want:   true,
		},
		{
			name:   "character varying type",
			column: Column{Type: "character varying"},
			want:   true,
		},
		{
			name:   "char type",
			column: Column{Type: "char"},
			want:   true,
		},
		{
			name:   "character type",
			column: Column{Type: "character"},
			want:   true,
		},
		{
			name:   "TEXT uppercase",
			column: Column{Type: "TEXT"},
			want:   true,
		},
		{
			name:   "integer type",
			column: Column{Type: "integer"},
			want:   false,
		},
		{
			name:   "uuid type",
			column: Column{Type: "uuid"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.IsString()
			if got != tt.want {
				t.Errorf("IsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_IsInteger(t *testing.T) {
	tests := []struct {
		name   string
		column Column
		want   bool
	}{
		{
			name:   "integer type",
			column: Column{Type: "integer"},
			want:   true,
		},
		{
			name:   "int type",
			column: Column{Type: "int"},
			want:   true,
		},
		{
			name:   "int4 type",
			column: Column{Type: "int4"},
			want:   true,
		},
		{
			name:   "bigint type",
			column: Column{Type: "bigint"},
			want:   true,
		},
		{
			name:   "int8 type",
			column: Column{Type: "int8"},
			want:   true,
		},
		{
			name:   "smallint type",
			column: Column{Type: "smallint"},
			want:   true,
		},
		{
			name:   "int2 type",
			column: Column{Type: "int2"},
			want:   true,
		},
		{
			name:   "INTEGER uppercase",
			column: Column{Type: "INTEGER"},
			want:   true,
		},
		{
			name:   "text type",
			column: Column{Type: "text"},
			want:   false,
		},
		{
			name:   "decimal type",
			column: Column{Type: "decimal"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.IsInteger()
			if got != tt.want {
				t.Errorf("IsInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_IsBoolean(t *testing.T) {
	tests := []struct {
		name   string
		column Column
		want   bool
	}{
		{
			name:   "boolean type",
			column: Column{Type: "boolean"},
			want:   true,
		},
		{
			name:   "bool type",
			column: Column{Type: "bool"},
			want:   true,
		},
		{
			name:   "BOOLEAN uppercase",
			column: Column{Type: "BOOLEAN"},
			want:   true,
		},
		{
			name:   "text type",
			column: Column{Type: "text"},
			want:   false,
		},
		{
			name:   "integer type",
			column: Column{Type: "integer"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.IsBoolean()
			if got != tt.want {
				t.Errorf("IsBoolean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_IsTimestamp(t *testing.T) {
	tests := []struct {
		name   string
		column Column
		want   bool
	}{
		{
			name:   "timestamp type",
			column: Column{Type: "timestamp"},
			want:   true,
		},
		{
			name:   "timestamptz type",
			column: Column{Type: "timestamptz"},
			want:   true,
		},
		{
			name:   "timestamp with time zone",
			column: Column{Type: "timestamp with time zone"},
			want:   true,
		},
		{
			name:   "timestamp without time zone",
			column: Column{Type: "timestamp without time zone"},
			want:   true,
		},
		{
			name:   "TIMESTAMP uppercase",
			column: Column{Type: "TIMESTAMP"},
			want:   true,
		},
		{
			name:   "date type",
			column: Column{Type: "date"},
			want:   true,
		},
		{
			name:   "time type",
			column: Column{Type: "time"},
			want:   true,
		},
		{
			name:   "text type",
			column: Column{Type: "text"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.IsTimestamp()
			if got != tt.want {
				t.Errorf("IsTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_GoFieldName(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
		want       string
	}{
		{
			name:       "simple column name",
			columnName: "name",
			want:       "Name",
		},
		{
			name:       "snake_case column name",
			columnName: "first_name",
			want:       "FirstName",
		},
		{
			name:       "multiple underscores",
			columnName: "user_profile_id",
			want:       "UserProfileId",
		},
		{
			name:       "single character",
			columnName: "a",
			want:       "A",
		},
		{
			name:       "empty string",
			columnName: "",
			want:       "",
		},
		{
			name:       "id field",
			columnName: "id",
			want:       "Id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := Column{Name: tt.columnName}
			got := column.GoFieldName()
			if got != tt.want {
				t.Errorf("GoFieldName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_GoStructTag(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
		want       string
	}{
		{
			name:       "simple column name",
			columnName: "name",
			want:       `json:"name" db:"name"`,
		},
		{
			name:       "snake_case column name",
			columnName: "first_name",
			want:       `json:"first_name" db:"first_name"`,
		},
		{
			name:       "id field",
			columnName: "id",
			want:       `json:"id" db:"id"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := Column{Name: tt.columnName}
			got := column.GoStructTag()
			if got != tt.want {
				t.Errorf("GoStructTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_GoFunctionName(t *testing.T) {
	tests := []struct {
		name      string
		queryName string
		want      string
	}{
		{
			name:      "simple query name",
			queryName: "get_user",
			want:      "GetUser",
		},
		{
			name:      "snake_case query name",
			queryName: "get_user_by_email",
			want:      "GetUserByEmail",
		},
		{
			name:      "multiple underscores",
			queryName: "get_user_profile_settings",
			want:      "GetUserProfileSettings",
		},
		{
			name:      "single word",
			queryName: "users",
			want:      "Users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := Query{Name: tt.queryName}
			got := query.GoFunctionName()
			if got != tt.want {
				t.Errorf("GoFunctionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_GoFileName(t *testing.T) {
	tests := []struct {
		name       string
		sourceFile string
		want       string
	}{
		{
			name:       "simple file name",
			sourceFile: "users.sql",
			want:       "users_queries_generated.go",
		},
		{
			name:       "snake_case file name",
			sourceFile: "user_queries.sql",
			want:       "user_queries_queries_generated.go",
		},
		{
			name:       "file with path",
			sourceFile: "sql/queries/users.sql",
			want:       "users_queries_generated.go",
		},
		{
			name:       "nested path",
			sourceFile: "internal/sql/user_management.sql",
			want:       "user_management_queries_generated.go",
		},
		{
			name:       "PascalCase file name",
			sourceFile: "UserQueries.sql",
			want:       "user_queries_queries_generated.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := Query{SourceFile: tt.sourceFile}
			got := query.GoFileName()
			if got != tt.want {
				t.Errorf("GoFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple word",
			input: "user",
			want:  "User",
		},
		{
			name:  "snake_case",
			input: "user_profile",
			want:  "UserProfile",
		},
		{
			name:  "multiple underscores",
			input: "user_profile_settings",
			want:  "UserProfileSettings",
		},
		{
			name:  "single character",
			input: "a",
			want:  "A",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "already PascalCase",
			input: "UserProfile",
			want:  "UserProfile",
		},
		{
			name:  "mixed case with underscores",
			input: "user_Profile_Settings",
			want:  "UserProfileSettings",
		},
		{
			name:  "trailing underscore",
			input: "user_profile_",
			want:  "UserProfile",
		},
		{
			name:  "leading underscore",
			input: "_user_profile",
			want:  "UserProfile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("toPascalCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple word",
			input: "user",
			want:  "user",
		},
		{
			name:  "PascalCase",
			input: "UserProfile",
			want:  "user_profile",
		},
		{
			name:  "camelCase",
			input: "userProfile",
			want:  "user_profile",
		},
		{
			name:  "multiple words",
			input: "UserProfileSettings",
			want:  "user_profile_settings",
		},
		{
			name:  "single character",
			input: "A",
			want:  "a",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "already snake_case",
			input: "user_profile",
			want:  "user_profile",
		},
		{
			name:  "all uppercase",
			input: "USER",
			want:  "u_s_e_r",
		},
		{
			name:  "mixed with numbers",
			input: "User123Profile",
			want:  "user123_profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryType_Constants(t *testing.T) {
	// Test that query type constants are defined correctly
	tests := []struct {
		name  string
		value QueryType
		want  string
	}{
		{
			name:  "QueryTypeOne",
			value: QueryTypeOne,
			want:  "one",
		},
		{
			name:  "QueryTypeMany",
			value: QueryTypeMany,
			want:  "many",
		},
		{
			name:  "QueryTypeExec",
			value: QueryTypeExec,
			want:  "exec",
		},
		{
			name:  "QueryTypePaginated",
			value: QueryTypePaginated,
			want:  "paginated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.value)
			if got != tt.want {
				t.Errorf("QueryType constant = %v, want %v", got, tt.want)
			}
		})
	}
}
