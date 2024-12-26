package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
)

type checker struct {
	tableWithSchemaByTable map[string]tableWithSchema // map[table]t
}

type tableWithSchema struct {
	table  string
	schema string
}

func newChecker(tablesBySchema map[string][]string) (checker, error) {
	tableWithSchemaByTable := make(map[string]tableWithSchema)
	for schema, tables := range tablesBySchema {
		for _, table := range tables {
			if _, exists := tableWithSchemaByTable[table]; exists {
				return checker{}, fmt.Errorf("table %s already exists", table)
			}

			tableWithSchemaByTable[table] = tableWithSchema{
				table:  table,
				schema: schema,
			}
		}
	}

	return checker{tableWithSchemaByTable: tableWithSchemaByTable}, nil
}

func (c checker) check(_ context.Context, sql string) ([]violation, error) {
	parser := parser.New()
	stmtNodes, _, err := parser.ParseSQL(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	visitor := newForeignKeyVisitor(sql, c.isAllowedReference)
	for _, stmtNode := range stmtNodes {
		stmtNode.Accept(visitor)
	}

	violations := []violation{}
	for _, result := range visitor.Results {
		violations = append(violations, result)
	}

	return violations, nil
}

func (c checker) isAllowedReference(source, target string) bool {
	sourceT, sourceTExists := c.tableWithSchemaByTable[source]
	targetT, targetTExists := c.tableWithSchemaByTable[target]

	if !sourceTExists && !targetTExists {
		return true
	}
	if sourceTExists && !targetTExists {
		return true
	}
	if !sourceTExists && targetTExists {
		return false
	}
	return sourceT.schema == targetT.schema
}

type violation struct {
	Stmt        string
	SourceTable string
	TargetTable string
	Row         int
	Column      int
}

func (v violation) String() string {
	return fmt.Sprintf("%s: %s -> %s", v.Stmt, v.SourceTable, v.TargetTable)
}

func newForeignKeyVisitor(sql string, isAllowedReference func(source, target string) bool) *foreignKeyVisitor {
	return &foreignKeyVisitor{
		IsAllowedReference: isAllowedReference,
		Lines:              strings.Split(sql, "\n"),
		Results:            []violation{},
	}
}

type foreignKeyVisitor struct {
	IsAllowedReference func(source, target string) bool
	Lines              []string
	Results            []violation
}

func (v *foreignKeyVisitor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	switch node := in.(type) {
	case *ast.CreateTableStmt:
		v.processCreateTable(node)
	case *ast.AlterTableStmt:
		v.processAlterTable(node)
	}
	return in, false
}

func (v *foreignKeyVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (v *foreignKeyVisitor) processCreateTable(node *ast.CreateTableStmt) {
	sourceTable := node.Table.Name.O
	for _, constraint := range node.Constraints {
		if constraint.Tp == ast.ConstraintForeignKey {
			targetTable := constraint.Refer.Table.Name.O
			if !v.IsAllowedReference(sourceTable, targetTable) {
				v.Results = append(v.Results, violation{
					Stmt:        "Create Table",
					SourceTable: sourceTable,
					TargetTable: targetTable,
					Row:         0,
					Column:      0,
				})
			}
		}
	}
}

func (v *foreignKeyVisitor) processAlterTable(node *ast.AlterTableStmt) {
	sourceTable := node.Table.Name.O
	for _, spec := range node.Specs {
		if spec.Tp == ast.AlterTableAddConstraint && spec.Constraint.Tp == ast.ConstraintForeignKey {
			targetTable := spec.Constraint.Refer.Table.Name.O
			if !v.IsAllowedReference(sourceTable, targetTable) {
				v.Results = append(v.Results, violation{
					Stmt:        "Alter Table",
					SourceTable: sourceTable,
					TargetTable: targetTable,
					Row:         0,
					Column:      0,
				})
			}
		}
	}
}
