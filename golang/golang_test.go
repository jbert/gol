package golang

import (
	"fmt"
	"os/exec"
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
		fmt.Printf("%d: running: %s\n", i, tc.Code)
		output, errStr := runProgram(tc.Code)
		if !strings.HasPrefix(errStr, tc.ErrOutput) {
			t.Errorf("%d@ wrong error [%s] != [%s] for code: %s\n", i, errStr, tc.ErrOutput, tc.Code)
			continue CASE
		}
		if output != tc.Result {
			fmt.Printf("%d@ wrong result [%s] != [%s] for code: %s\n", i, output, tc.Result, tc.Code)
			t.Errorf("%d@ wrong result [%s] != [%s] for code: %s\n", i, output, tc.Result, tc.Code)
			continue CASE
		}
		fmt.Printf("%d: AOK!\n", i)
	}
}

func runProgram(prog string) (string, string) {

	sourceFilename := "<internal>"
	outputFilename := tempFileName("exe")
	//	defer os.Remove(outputFilename)

	r := strings.NewReader(prog)
	err := CompileReader(sourceFilename, r, outputFilename)
	if err != nil {
		return "", err.Error()
	}

	cmd := exec.Command(outputFilename)
	value, err := cmd.CombinedOutput()
	if err != nil {
		return "", err.Error()
	}
	return string(value), ""
}
