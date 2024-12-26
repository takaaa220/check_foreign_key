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
		filePaths      []string
		configFilePath string
		want           []string
		wantErr        bool
	}{
		"testdata1": {
			filePaths:      []string{"testdata/sql"},
			configFilePath: "testdata/config.json",
			want: []string{
				"Foreign key constraint \"users -> addresses\" in testdata/sql/1_.up.sql is not allowed",
				"Foreign key constraint \"addresses -> users\" in testdata/sql/1_.up.sql is not allowed",
				"Foreign key constraint \"abs -> users\" in testdata/sql/2_.up.sql is not allowed",
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reports, err := do(tt.filePaths, tt.configFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, reports, cmpopts.SortSlices(func(i, j string) bool {
				return i < j
			})); diff != "" {
				t.Errorf("do() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
