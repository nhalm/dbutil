package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nhalm/dbutil/gen"
)

func main() {
	var (
		dsn         = flag.String("dsn", "", "PostgreSQL connection string (or use DATABASE_URL env var)")
		output      = flag.String("output", "./repositories", "Output directory for generated files")
		schema      = flag.String("schema", "public", "Database schema to introspect")
		queries     = flag.String("queries", "", "Directory containing SQL query files")
		tables      = flag.Bool("tables", false, "Generate table-based repositories")
		include     = flag.String("include", "", "Comma-separated list of tables to include")
		exclude     = flag.String("exclude", "", "Comma-separated list of tables to exclude")
		config      = flag.String("config", "", "Path to configuration file")
		packageName = flag.String("package", "repositories", "Package name for generated code")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Get DSN from environment if not provided
	if *dsn == "" {
		*dsn = os.Getenv("DATABASE_URL")
	}

	// TODO: try to pull env vars using the connection package.
	if *dsn == "" {
		fmt.Fprintf(os.Stderr, "Error: Database connection string required (use --dsn or DATABASE_URL)\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate that at least one generation mode is enabled
	if !*tables && *queries == "" {
		fmt.Fprintf(os.Stderr, "Error: Must specify --tables and/or --queries\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse include/exclude lists
	var includeList, excludeList []string
	if *include != "" {
		includeList = strings.Split(*include, ",")
		for i := range includeList {
			includeList[i] = strings.TrimSpace(includeList[i])
		}
	}
	if *exclude != "" {
		excludeList = strings.Split(*exclude, ",")
		for i := range excludeList {
			excludeList[i] = strings.TrimSpace(excludeList[i])
		}
	}

	// Create generator configuration
	cfg := &gen.Config{
		DSN:         *dsn,
		Schema:      *schema,
		OutputDir:   *output,
		PackageName: *packageName,
		QueriesDir:  *queries,
		Tables:      *tables,
		Include:     includeList,
		Exclude:     excludeList,
		Verbose:     *verbose,
	}

	// Load config file if specified
	if *config != "" {
		fileConfig, err := gen.LoadConfig(*config)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
		cfg = cfg.Merge(fileConfig)
	}

	// Create and run generator
	generator := gen.New(cfg)
	ctx := context.Background()

	if err := generator.Generate(ctx); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	fmt.Printf("Successfully generated code in %s\n", *output)
}
