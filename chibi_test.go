package gol

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func getScmFiles(baseDir string) []string {
	fnames, err := filepath.Glob(baseDir + "/*.scm")
	if err != nil {
		panic("Can't glob: " + baseDir)
	}
	return fnames
}

func TestChibi(t *testing.T) {

	fnames := getScmFiles("chibi-tests/basic")
	// Fail on first, so we see error details
	//	fnames := []string{"chibi-tests/basic/test00-fact-3.scm"}
FNAME:
	for _, fname := range fnames {
		ok, output := runFile(t, fname)
		if !ok {
			f, err := os.Open(fname)
			if err != nil {
				t.Logf("Failed to open scm file [%s]!: %s", fname, err)
				continue
			}
			defer f.Close()

			contents, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("Failed to read scheme src: %s", err)
				continue
			}

			t.Logf("NOTOK %s failed: [%s]\n->%s\n", fname, contents, output)
			break FNAME
		} else {
			t.Logf("OK!!! %s succeeded\n", fname)
		}
	}
}

func baseName(fname string) string {
	lastDot := strings.LastIndex(fname, ".")
	if lastDot < 0 {
		return fname
	}
	return fname[:lastDot]
}

func runFile(t *testing.T, fname string) (bool, string) {
	base := baseName(fname)
	t.Logf("Running %s", base)

	scm := base + ".scm"
	res := base + ".res"

	cmd := exec.Command("go", "run", "cmd/gol/gol.go", "-f", scm)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Failed to run gol: %s", err)
		return false, string(output)
	}

	f, err := os.Open(res)
	if err != nil {
		t.Errorf("Failed to open res file %s: %s", res, err)
		return false, string(output)
	}
	defer f.Close()

	expected, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("Failed to read expected result: %s", err)
		return false, string(output)
	}

	if string(expected) == string(output) {
		t.Logf("Got expected result: %s", output)
		return true, string(output)
	} else {
		t.Errorf("Got wrong result: [%v] != [%v]", output, expected)
		return false, string(output)
	}
}
