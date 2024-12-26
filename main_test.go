package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

func Test_do(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filePaths       []string
		tableWithSchema map[string][]string
		want            []string
		wantErr         bool
	}{
		"testdata1": {
			filePaths: []string{"testdata"},
			tableWithSchema: map[string][]string{
				"schema1": {"table1", "table2", "users"},
				"schema2": {"table3", "table4", "addresses"},
			},
			want: []string{
				"Foreign key constraint \"users -> addresses\" in testdata/1_.up.sql is not allowed",
				"Foreign key constraint \"addresses -> users\" in testdata/1_.up.sql is not allowed",
				"Foreign key constraint \"abs -> users\" in testdata/2_.up.sql is not allowed",
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reports, err := do(tt.filePaths, tt.tableWithSchema)
			if (err != nil) != tt.wantErr {
				t.Errorf("do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(reports, tt.want, cmpopts.SortSlices(func(i, j string) bool {
				return i < j
			})); diff != "" {
				t.Errorf("do() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
