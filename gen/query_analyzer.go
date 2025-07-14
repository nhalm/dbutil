package gen

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryAnalyzer analyzes SQL queries using PostgreSQL EXPLAIN to determine column types and validate queries
type QueryAnalyzer struct {
	db         *pgxpool.Pool
	typeMapper *TypeMapper
}

// NewQueryAnalyzer creates a new query analyzer
func NewQueryAnalyzer(db *pgxpool.Pool) *QueryAnalyzer {
	return &QueryAnalyzer{
		db:         db,
		typeMapper: NewTypeMapper(nil),
	}
}

// AnalyzeQuery analyzes a query using PostgreSQL EXPLAIN to determine column types and parameters
func (qa *QueryAnalyzer) AnalyzeQuery(ctx context.Context, query *Query) error {
	if query == nil {
		return fmt.Errorf("query cannot be nil")
	}

	// Extract parameters from the query
	if err := qa.extractParameters(query); err != nil {
		return fmt.Errorf("failed to extract parameters: %w", err)
	}

	// For SELECT queries, analyze columns using EXPLAIN
	if qa.isSelectQuery(query.Type) {
		if err := qa.analyzeSelectQuery(ctx, query); err != nil {
			return fmt.Errorf("failed to analyze SELECT query: %w", err)
		}
	}

	// Validate query syntax by attempting to prepare it
	if err := qa.validateQuerySyntax(ctx, query); err != nil {
		return fmt.Errorf("query syntax validation failed: %w", err)
	}

	return nil
}

// extractParameters extracts parameter placeholders from the SQL query
func (qa *QueryAnalyzer) extractParameters(query *Query) error {
	// Find all parameter placeholders ($1, $2, etc.)
	// Match $digits followed by non-digit or end of string
	paramRegex := regexp.MustCompile(`\$(\d+)(?:\D|$)`)
	matches := paramRegex.FindAllStringSubmatch(query.SQL, -1)

	if len(matches) == 0 {
		query.Parameters = []Parameter{}
		return nil
	}

	// Create a map to track unique parameter indices
	paramMap := make(map[int]bool)
	for _, match := range matches {
		if len(match) >= 2 {
			paramNum, err := strconv.Atoi(match[1])
			if err != nil {
				return fmt.Errorf("invalid parameter number: %s", match[1])
			}
			paramMap[paramNum] = true
		}
	}

	// Create parameter list
	var parameters []Parameter
	for i := 1; i <= len(paramMap); i++ {
		if !paramMap[i] {
			return fmt.Errorf("parameter $%d is missing (parameters must be sequential starting from $1)", i)
		}

		// For now, we'll use a generic parameter type
		// In a more advanced implementation, we could try to infer types from context
		param := Parameter{
			Name:   fmt.Sprintf("param%d", i),
			Type:   "text", // Default to text, can be overridden by type inference
			GoType: "string",
			Index:  i,
		}
		parameters = append(parameters, param)
	}

	query.Parameters = parameters
	return nil
}

// isSelectQuery checks if the query type requires column analysis
func (qa *QueryAnalyzer) isSelectQuery(queryType QueryType) bool {
	return queryType == QueryTypeOne || queryType == QueryTypeMany || queryType == QueryTypePaginated
}

// analyzeSelectQuery uses EXPLAIN to analyze a SELECT query and determine column types
func (qa *QueryAnalyzer) analyzeSelectQuery(ctx context.Context, query *Query) error {
	// Create a prepared statement to analyze the query structure
	// We'll use EXPLAIN with a dummy parameter set to analyze the query
	explainSQL := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", query.SQL)

	// Replace parameters with dummy values for EXPLAIN
	analyzableSQL := qa.replaceParametersForExplain(query.SQL, query.Parameters)
	explainSQL = fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", analyzableSQL)

	// Execute EXPLAIN query
	rows, err := qa.db.Query(ctx, explainSQL)
	if err != nil {
		return fmt.Errorf("failed to execute EXPLAIN query: %w", err)
	}
	defer rows.Close()

	// For now, we'll use a simpler approach: try to execute the query with dummy parameters
	// to get the column information from the result set
	return qa.analyzeQueryColumns(ctx, query)
}

// replaceParametersForExplain replaces parameter placeholders with dummy values for EXPLAIN
func (qa *QueryAnalyzer) replaceParametersForExplain(sql string, parameters []Parameter) string {
	result := sql
	for i := len(parameters); i >= 1; i-- {
		placeholder := fmt.Sprintf("$%d", i)
		// Use appropriate dummy values based on common types
		dummyValue := qa.getDummyValueForParameter(i)
		result = strings.ReplaceAll(result, placeholder, dummyValue)
	}
	return result
}

// getDummyValueForParameter returns a dummy value for a parameter based on its index
func (qa *QueryAnalyzer) getDummyValueForParameter(paramIndex int) string {
	// Use NULL which works with all types and avoids type conversion issues
	return "NULL"
}

// analyzeQueryColumns analyzes the columns returned by a SELECT query
func (qa *QueryAnalyzer) analyzeQueryColumns(ctx context.Context, query *Query) error {
	// Create a modified query that returns column information
	// We'll use a LIMIT 0 query to get column metadata without executing the full query
	limitedSQL := fmt.Sprintf("SELECT * FROM (%s) AS subquery LIMIT 0", query.SQL)

	// Replace parameters with dummy values
	analyzableSQL := qa.replaceParametersForExplain(limitedSQL, query.Parameters)

	// Execute the query to get column information
	rows, err := qa.db.Query(ctx, analyzableSQL)
	if err != nil {
		return fmt.Errorf("failed to analyze query columns: %w", err)
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	var columns []Column

	for _, field := range fieldDescriptions {
		// Map PostgreSQL OID to type name
		pgType := qa.mapOIDToTypeName(field.DataTypeOID)

		// Determine if the column is nullable (this is a simplified approach)
		isNullable := true // Default to nullable for query results

		// Map to Go type
		goType, err := qa.typeMapper.MapType(pgType, isNullable, false)
		if err != nil {
			return fmt.Errorf("failed to map column type for %s: %w", field.Name, err)
		}

		column := Column{
			Name:       field.Name,
			Type:       pgType,
			GoType:     goType,
			IsNullable: isNullable,
			IsArray:    false, // TODO: Detect array types from OID
		}
		columns = append(columns, column)
	}

	query.Columns = columns
	return nil
}

// mapOIDToTypeName maps PostgreSQL OID to type name
func (qa *QueryAnalyzer) mapOIDToTypeName(oid uint32) string {
	// Common PostgreSQL type OIDs
	// This is a simplified mapping - in a production system, you'd want a more comprehensive mapping
	switch oid {
	case 16:
		return "boolean"
	case 20:
		return "bigint"
	case 21:
		return "smallint"
	case 23:
		return "integer"
	case 25:
		return "text"
	case 700:
		return "real"
	case 701:
		return "double precision"
	case 1043:
		return "varchar"
	case 1082:
		return "date"
	case 1114:
		return "timestamp"
	case 1184:
		return "timestamptz"
	case 1700:
		return "numeric"
	case 2950:
		return "uuid"
	case 114:
		return "json"
	case 3802:
		return "jsonb"
	case 17:
		return "bytea"
	default:
		return "text" // Default to text for unknown types
	}
}

// validateQuerySyntax validates that the query is syntactically correct
func (qa *QueryAnalyzer) validateQuerySyntax(ctx context.Context, query *Query) error {
	// For exec queries, we can't use LIMIT 0, so we'll use a different approach
	if query.Type == QueryTypeExec {
		return qa.validateExecQuery(ctx, query)
	}

	// For SELECT queries, we already validated them in analyzeQueryColumns
	return nil
}

// validateExecQuery validates an EXEC query by preparing it
func (qa *QueryAnalyzer) validateExecQuery(ctx context.Context, query *Query) error {
	// Try to prepare the statement to validate syntax
	// We'll use a transaction that we roll back to avoid side effects
	tx, err := qa.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for validation: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare the statement with a unique name
	stmtName := fmt.Sprintf("validate_query_%s", query.Name)
	stmt, err := tx.Prepare(ctx, stmtName, query.SQL)
	if err != nil {
		return fmt.Errorf("query preparation failed: %w", err)
	}

	// Check that the parameter count matches
	if len(stmt.ParamOIDs) != len(query.Parameters) {
		return fmt.Errorf("parameter count mismatch: query expects %d parameters, found %d", len(stmt.ParamOIDs), len(query.Parameters))
	}

	// Update parameter types based on the prepared statement
	for i, paramOID := range stmt.ParamOIDs {
		if i < len(query.Parameters) {
			pgType := qa.mapOIDToTypeName(paramOID)
			goType, err := qa.typeMapper.MapType(pgType, false, false)
			if err != nil {
				return fmt.Errorf("failed to map parameter type: %w", err)
			}
			query.Parameters[i].Type = pgType
			query.Parameters[i].GoType = goType
		}
	}

	return nil
}

// InferParameterTypes attempts to infer parameter types from query context
func (qa *QueryAnalyzer) InferParameterTypes(ctx context.Context, query *Query) error {
	// This is a more advanced feature that could analyze the query context
	// to infer parameter types based on how they're used
	// For now, we'll keep the basic implementation from extractParameters
	return nil
}

// ValidateQueryExecution validates that a query can be executed successfully
func (qa *QueryAnalyzer) ValidateQueryExecution(ctx context.Context, query *Query) error {
	// This could be used to validate that the query executes without errors
	// using test data or in a test transaction
	return nil
}
