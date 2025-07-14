package main

import (
	"fmt"

	"github.com/nhalm/dbutil/gen"
)

func main() {
	parser := gen.NewQueryParser("../test-queries")
	queries, err := parser.ParseQueries()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d queries:\n", len(queries))
	for _, query := range queries {
		fmt.Printf("- %s (%s) from %s\n", query.Name, query.Type, query.SourceFile)

		// Validate each query
		if err := parser.ValidateQuery(query); err != nil {
			fmt.Printf("  ❌ Validation failed: %v\n", err)
		} else {
			fmt.Printf("  ✅ Valid\n")
		}
	}
}
