package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/pingcap/tidb/parser/test_driver"
)

var (
	tablesWithSchema = map[string][]string{
		"schema1": {"table1", "table2", "users"},
		"schema2": {"table3", "table4", "addresses"},
	}
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [file_or_dir_path1] [file_or_dir_path2]...", os.Args[0])
	}

	reports, err := do(os.Args[1:], tablesWithSchema)
	if err != nil {
		log.Fatal(err)
	}

	for _, report := range reports {
		fmt.Print(report)
	}
}

func do(filePaths []string, tablesBySchema map[string][]string) ([]string, error) {
	sqlFiles := []string{}
	for _, filePath := range filePaths {
		files, err := getSQLFilesRecursively(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to getSQLFilesRecursively: %v", err)
		}

		sqlFiles = append(sqlFiles, files...)
	}

	if len(sqlFiles) == 0 {
		log.Print("No SQL files found\n")
		return nil, nil
	}

	checker, err := newChecker(tablesWithSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to newChecker: %v", err)
	}

	var reports []string
	for _, sqlFile := range sqlFiles {
		sql, err := os.ReadFile(sqlFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %v", sqlFile, err)
		}

		violations, err := checker.check(context.TODO(), string(sql))
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %v", sqlFile, err)
		}

		if len(violations) > 0 {
			for _, violation := range violations {
				reports = append(reports, fmt.Sprintf("Foreign key constraint \"%s -> %s\" in %s is not allowed", violation.SourceTable, violation.TargetTable, sqlFile))
			}
		}
	}

	return reports, nil
}

func getSQLFilesRecursively(filePath string) ([]string, error) {
	var sqlFiles []string
	if err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			sqlFiles = append(sqlFiles, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return sqlFiles, nil
}
