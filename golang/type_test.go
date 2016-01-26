package golang

import (
	"testing"

	"github.com/jbert/gol/typ"
)

func TestGolangTypeString(t *testing.T) {
	testCases := []struct {
		testType typ.Type
		expected string
	}{
		{typ.Num, "int64"},
		{typ.String, "string"},
		{typ.Bool, "bool"},
		{typ.Symbol, "string"},
		{typ.NewFunc([]typ.Type{typ.String}, typ.String), "func(string) string"},
		{typ.NewFunc([]typ.Type{
			typ.String,
			typ.Num,
			typ.Bool,
		}, typ.Num), "func(string,int64,bool) int64"},
	}

	for _, tc := range testCases {
		golangStr := golangStringForType(tc.testType)
		if golangStr != tc.expected {
			t.Errorf("Failed [%s]: %s != %s", tc.testType, golangStr, tc.expected)
		} else {
			t.Logf("Worked [%s]: %s", tc.testType, golangStr)
		}
	}
}
