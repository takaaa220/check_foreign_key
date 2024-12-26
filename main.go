package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/pingcap/tidb/parser/test_driver"
)

func main() {
	sqlPathStr := flag.String("paths", "", "[required] comma separated file or directory paths")
	configPath := flag.String("config", "", "[required] config json file path")
	flag.Parse()
	if *sqlPathStr == "" || *configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	reports, err := do(strings.Split(*sqlPathStr, ","), *configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, report := range reports {
		fmt.Println(report)
	}
}

func do(filePaths []string, configFilePath string) ([]string, error) {
	tablesBySchema, err := parseConfig(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parseConfig: %v", err)
	}

	sqlFiles := []string{}
	for _, filePath := range filePaths {
		files, err := getSQLFilesRecursively(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to getSQLFilesRecursively: %v", err)
		}

		sqlFiles = append(sqlFiles, files...)
	}

	if len(sqlFiles) == 0 {
		return nil, nil
	}

	checker, err := newChecker(tablesBySchema)
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

func parseConfig(configPath string) (map[string][]string, error) {
	f, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var tablesWithSchema map[string][]string
	if err := json.Unmarshal(f, &tablesWithSchema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	return tablesWithSchema, nil
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
