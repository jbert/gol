package golang

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"text/template"

	"github.com/jbert/gol"
)

// TODO: pull this out as ParseFile and call from Evaluatator too
func CompileReader(filename string, r io.Reader, outFilename string) error {
	l := gol.NewLexer(filename, r)

	// Run the lexer until EOF or error
	var lexErr error
	lexDone := make(chan struct{})
	go func() {
		lexErr = l.Run()
		close(lexDone)
	}()

	// Run the parser until the lexer finishes
	p := gol.NewParser(l.Tokens)
	nodeTree, parseErr := p.Parse()
	if parseErr != nil {
		return parseErr
	}

	// Hoover up any lexing errors
	<-lexDone
	if lexErr != nil {
		return lexErr
	}

	// We have a basic parse tree, decorate it with additional
	// node information
	nodeTree, parseErr = gol.Transform(nodeTree)
	if parseErr != nil {
		return parseErr
	}

	// TODO -
	gb := NewGolangBackend(nodeTree)
	err := gb.CompileTo(outFilename)
	if err != nil {
		return err
	}

	return nil
}

func CompileFile(filename string, outFilename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	return CompileReader(filename, f, outFilename)
}

type GolangBackend struct {
	parseTree   gol.Node
	tmpDir      string
	symbolIndex int
	w           io.Writer
	funcDefns   []string
}

func NewGolangBackend(parseTree gol.Node) *GolangBackend {
	gb := GolangBackend{parseTree: parseTree}
	return &gb
}

func tempFileName(extension string) string {
	randomNumber := rand.Int63()
	return fmt.Sprintf("%s/gol-%x.%s", os.TempDir(), randomNumber, extension)
}

func (gb *GolangBackend) CompileTo(outFilename string) error {

	tmpGoFilename := tempFileName("go")
	f, err := os.Create(tmpGoFilename)
	if err != nil {
		return fmt.Errorf("Failded to create file [%s]: %s", tmpGoFilename, err)
	}
	defer f.Close()
	gb.w = f
	defer os.Remove(tmpGoFilename)

	err = gb.writePreamble()
	if err != nil {
		return fmt.Errorf("Failed to write preamble: %s", err)
	}
	err = gb.write()
	if err != nil {
		return fmt.Errorf("Failed to write go code : %s", err)
	}
	err = gb.writePostamble()
	if err != nil {
		return fmt.Errorf("Failed to write postamble: %s", err)
	}

	err = gb.buildGo(tmpGoFilename, outFilename)
	if err != nil {
		return fmt.Errorf("Failded to build go file: %s", err)
	}

	return nil
}

func (gb *GolangBackend) neededPackages() []string {
	return []string{"fmt"}
}

func (gb *GolangBackend) writePreamble() error {
	info := struct {
		Packages []string
	}{
		Packages: gb.neededPackages(),
	}
	tmpl := template.Must(template.New("preamble").Parse(templatePreamble))
	err := tmpl.Execute(gb.w, info)
	return err
}

var templatePreamble = `package main

import (
	{{range .Packages}} "{{.}}" {{end}}
) 

func main() {
`

func (gb *GolangBackend) write() error {
	node, ok := gb.parseTree.(gol.NodeProgn)
	if !ok {
		return fmt.Errorf("Tree isn't a progn")
	}
	gb.writeProgn(node)
	return nil
}

func (gb *GolangBackend) emit(s string, rest ...interface{}) {
	_, err := fmt.Fprintf(gb.w, s, rest...)
	if err != nil {
		panic(fmt.Sprintf("Failed to emit compiled output: %s\n", err))
	}
}

func (gb *GolangBackend) makeFunctionName() string {
	gb.symbolIndex++
	return fmt.Sprintf("func%d", gb.symbolIndex)
}

func (gb *GolangBackend) saveFunc(s string) {
	gb.funcDefns = append(gb.funcDefns, s)
}

// Emit a function call, and stack the definition for the postamble
func (gb *GolangBackend) writeProgn(node gol.NodeProgn) error {
	funcName := gb.makeFunctionName()
	gb.emit("%s()\n", funcName)
	gb.saveFunc(fmt.Sprintf(`func %s() { fmt.Printf("hi\n") }
`, funcName))
	return nil
}

func (gb *GolangBackend) writePostamble() error {
	gb.emit("}\n")
	for _, s := range gb.funcDefns {
		gb.emit("\n")
		gb.emit(s)
		gb.emit("\n")
	}
	return nil
}

func (gb *GolangBackend) buildGo(goFilename string, outFilename string) error {
	cmd := exec.Command("go", "build", "-o", outFilename, goFilename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to compile [%s]\n%s\n\n", err, output)
	}
	return nil
}
