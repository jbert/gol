package eval

import (
	"strings"
	"testing"

	"github.com/jbert/gol/test"
)

func TestGolFunc(t *testing.T) {
	runCases(t, test.FuncTestCases())
}

func TestGolError(t *testing.T) {
	runCases(t, test.ErrorTestCases())
}
func TestGolQuote(t *testing.T) {
	runCases(t, test.QuoteTestCases())
}

func TestGolBasicTestCases(t *testing.T) {
	runCases(t, test.BasicTestCases())
}

func runCases(t *testing.T, testCases []test.TestCase) {

CASE:
	for i, tc := range testCases {
		//		fmt.Printf("%d: running: %s\n", i, tc.code)
		evalStr, errStr := evaluateProgram(tc.Code)
		if !strings.HasPrefix(errStr, tc.ErrOutput) {
			t.Errorf("%d@ wrong error [%s] != [%s] for code: %s\n", i, errStr, tc.ErrOutput, tc.Code)
			continue CASE
		}
		if evalStr != tc.Result {
			t.Errorf("%d@ wrong result [%s] != [%s] for code: %s\n", i, evalStr, tc.Result, tc.Code)
			continue CASE
		}
		//		t.Logf("%d: AOK!\n", i)
	}
}

func evaluateProgram(prog string) (string, string) {

	fname := "<internal>"
	g := New()
	value, err := g.EvalProgram(fname, prog)
	if err != nil {
		return "", err.Error()
	}
	return value.String(), ""
}
