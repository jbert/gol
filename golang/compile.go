package golang

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
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
	//	fmt.Printf("TRANSFORM: %s\n", nodeTree)
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
	//defer os.Remove(tmpGoFilename)

	preamble, err := gb.compilePreamble()
	if err != nil {
		return fmt.Errorf("Failed to make preamble: %s", err)
	}
	_, err = io.WriteString(f, preamble)
	if err != nil {
		return fmt.Errorf("Failed to write preamble: %s", err)
	}

	code, err := gb.compileBody()
	if err != nil {
		return fmt.Errorf("Failed to compile to go code : %s", err)
	}
	_, err = io.WriteString(f, code)
	if err != nil {
		return fmt.Errorf("Failed to write go code: %s", err)
	}

	postamble, err := gb.compilePostamble()
	if err != nil {
		return fmt.Errorf("Failed to write postamble: %s", err)
	}
	_, err = io.WriteString(f, postamble)
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

func (gb *GolangBackend) compilePreamble() (string, error) {
	info := struct {
		Packages []string
	}{
		Packages: gb.neededPackages(),
	}
	tmpl := template.Must(template.New("preamble").Parse(templatePreamble))

	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, info)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

var templatePreamble = `package main

import (
	{{range .Packages}} "{{.}}" {{end}}
) 

func main() {
`

func (gb *GolangBackend) compileBody() (string, error) {
	node, ok := gb.parseTree.(gol.NodeProgn)
	if !ok {
		return "", fmt.Errorf("Tree isn't a progn")
	}
	s, err := gb.compile(node)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`
	fmt.Printf("%%d\n", %s)
`, s), nil
}

func (gb *GolangBackend) compile(node gol.Node) (string, error) {
	switch n := node.(type) {
	case gol.NodeProgn:
		return gb.compileProgn(n)
	case gol.NodeInt:
		return gb.compileInt(n)

	case gol.NodeError:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeIdentifier:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeSymbol:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeString:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeBool:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeQuote:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeUnQuote:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeLambda:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeList:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeIf:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeSet:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeLet:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeDefine:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	default:
		return "", gol.NodeErrorf(n, "Unrecognised node type %T", node)

	}
}

func (gb *GolangBackend) compileInt(node gol.NodeInt) (string, error) {
	return fmt.Sprintf("%d", node.Value()), nil
}

// Emit a function call, and stack the definition for the postamble
func (gb *GolangBackend) compileProgn(node gol.NodeProgn) (string, error) {
	funcName := gb.makeFunctionName()
	if node.Len() == 0 {
		return "", nil
	}

	first := true

	lines := []string{fmt.Sprintf(`
func %s() int64 {
	var a int64
`, funcName)}
	_, err := node.Map(func(n gol.Node) (gol.Node, error) {
		if first {
			first = false
			return nil, nil
		}
		s, err := gb.compile(n)
		if err != nil {
			return nil, err
		}
		lines = append(lines, `a = `+s)
		return nil, nil
	})
	if err != nil {
		return "", err
	}
	lines = append(lines, `
	return a
}
`)

	gb.saveFunc(strings.Join(lines, "\n"))
	return fmt.Sprintf("%s()", funcName), nil
}

func (gb *GolangBackend) makeFunctionName() string {
	gb.symbolIndex++
	return fmt.Sprintf("func_%d", gb.symbolIndex)
}

func (gb *GolangBackend) saveFunc(s string) {
	gb.funcDefns = append(gb.funcDefns, s)
}

func (gb *GolangBackend) compilePostamble() (string, error) {
	buf := &bytes.Buffer{}

	buf.WriteString("}\n")
	for _, s := range gb.funcDefns {
		buf.WriteString("\n")
		buf.WriteString(s)

		buf.WriteString("\n")
	}
	return buf.String(), nil
}

func (gb *GolangBackend) buildGo(goFilename string, outFilename string) error {
	cmd := exec.Command("go", "build", "-o", outFilename, goFilename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to compile [%s]\n%s\n\n", err, output)
	}
	return nil
}
