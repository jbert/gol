package golang

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

	runFrom := 0
	runFromStr := os.Getenv("GOL_TEST_RUNFROM")
	if runFromStr != "" {
		runFrom, _ = strconv.Atoi(runFromStr)
	}

	runTo := len(testCases)
	runToStr := os.Getenv("GOL_TEST_RUNTO")
	if runToStr != "" {
		runTo, _ = strconv.Atoi(runToStr)
	}
CASE:
	for i, tc := range testCases {
		if i < runFrom {
			continue CASE
		}
		if i > runTo {
			break CASE
		}

		t.Logf("%d: running: %s\n", i, tc.Code)
		fmt.Printf("%d: running: %s\n", i, tc.Code)
		output, err := runProgram(tc.Code)
		if err != nil {
			if tc.ErrOutput == "" {
				t.Errorf("%d@ unexpected error [%s] for code: %s\n", i, err.Error(), tc.Code)
				continue CASE

			}
			if !strings.HasPrefix(err.Error(), tc.ErrOutput) {
				t.Errorf("%d@ wrong error [%s] != [%s] for code: %s\n", i, err, tc.ErrOutput, tc.Code)
			}
			t.Logf("%d: AOK (correct error %s)!\n", i, err.Error())
			continue CASE
		}
		if output != tc.Result {
			t.Logf("%d@ wrong result [%s] != [%s] for code: %s\n", i, output, tc.Result, tc.Code)
			t.Errorf("%d@ wrong result [%s] != [%s] for code: %s\n", i, output, tc.Result, tc.Code)
			continue CASE
		}
		t.Logf("%d: AOK!\n", i)
	}
}

func runProgram(prog string) (string, error) {

	sourceFilename := "<internal>"
	outputFilename := tempFileName("exe")
	//	defer os.Remove(outputFilename)

	r := strings.NewReader(prog)
	err := CompileReader(sourceFilename, r, outputFilename)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(outputFilename)
	value, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	value = bytes.TrimRight(value, "\n")
	return string(value), nil
}
