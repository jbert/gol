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
		{typ.Int, "int64"},
		{typ.String, "string"},
		{typ.Bool, "bool"},
		{typ.Symbol, "string"},
		{typ.NewFunc([]typ.Type{typ.String}, typ.String), "func(string) string"},
		{typ.NewFunc([]typ.Type{
			typ.String,
			typ.Int,
			typ.Bool,
		}, typ.Int), "func(string,int64,bool) int64"},
	}

	for _, tc := range testCases {
		golangStr, err := golangStringForType(tc.testType)
		if err != nil {
			t.Errorf("Error return [%s]", err)
		}
		if golangStr != tc.expected {
			t.Errorf("Failed [%s]: %s != %s", tc.testType, golangStr, tc.expected)
		} else {
			t.Logf("Worked [%s]: %s", tc.testType, golangStr)
		}
	}
}
