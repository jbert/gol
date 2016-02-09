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
	"github.com/jbert/gol/typ"
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
	parseTree gol.Node
	funcDefns []string
	typeEnv   typ.Env
}

func NewGolangBackend(parseTree gol.Node) *GolangBackend {
	gb := GolangBackend{
		parseTree: parseTree,
		typeEnv:   newDefaultTypeEnv(),
	}
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
		return fmt.Errorf("Failed to create file [%s]: %s", tmpGoFilename, err)
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

	_, err = io.WriteString(f, gb.standardLib())
	if err != nil {
		return fmt.Errorf("Failed to write standard lib: %s", err)
	}

	err = gb.buildGo(tmpGoFilename, outFilename)
	if err != nil {
		return fmt.Errorf("Failed to build go file: %s", err)
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
	case gol.NodeList:
		return gb.compileList(n)
	case gol.NodeLet:
		return gb.compileLet(n)
	case gol.NodeIdentifier:
		return gb.compileIdentifier(n)

	case gol.NodeError:
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
	case gol.NodeIf:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeSet:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	case gol.NodeDefine:
		return "", gol.NodeErrorf(n, "TODO node type %T", node)
	default:
		return "", gol.NodeErrorf(n, "Unrecognised node type %T", node)

	}
}

func (gb *GolangBackend) compileIdentifier(ni gol.NodeIdentifier) (string, error) {
	return mangleIdentifier(ni.String()), nil
}

func (gb *GolangBackend) compileLet(nl gol.NodeLet) (string, error) {
	args := []string{}
	vals := []string{}

	for k, vNode := range nl.Bindings {
		golangType := golangStringForType(vNode.Type(gb.typeEnv))
		args = append(args, fmt.Sprintf("%s %s", mangleIdentifier(k), golangType))
		val, err := gb.compile(vNode)
		if err != nil {
			return "", err
		}
		vals = append(vals, val)
	}

	s := "func" + "(" + strings.Join(args, ", ") + ") int64 {"
	body, err := gb.compile(nl.Body)
	if err != nil {
		return "", err
	}
	s += "return " + body
	s += "}(" + strings.Join(vals, ", ") + ")"
	return s, nil
}

func (gb *GolangBackend) compileList(nl gol.NodeList) (string, error) {
	if nl.Len() == 0 {
		return "", gol.NodeErrorf(nl, "empty application")
	}
	switch fst := nl.First().(type) {
	case gol.NodeIdentifier:
		return gb.compileFuncCall(fst, nl.Rest())
	case gol.NodeLambda:
		return gb.compileLambda(nl)
	default:
		return "", fmt.Errorf("Non-applicable in head position: %T", fst)
	}
}

func mangleIdentifier(s string) string {
	s = strings.Replace(s, "+", "__PLUS__", -1)
	s = strings.Replace(s, "-", "__MINUS__", -1)
	return s
}

func (gb *GolangBackend) compileFuncCall(funcNameNode gol.NodeIdentifier, argNodes gol.NodeList) (string, error) {

	funcName := mangleIdentifier(funcNameNode.String())
	args := []string{}
	err := argNodes.Foreach(func(n gol.Node) error {
		nStr, err := gb.compile(n)
		if err != nil {
			return err
		}
		args = append(args, nStr)
		return nil
	})
	if err != nil {
		return "", err
	}
	s := funcName + "(" + strings.Join(args, ", ") + ")"
	return s, nil
}

func (gb *GolangBackend) compileLambda(nl gol.NodeList) (string, error) {
	panic("TODO - compileLambda")
}

func (gb *GolangBackend) compileInt(ni gol.NodeInt) (string, error) {
	return fmt.Sprintf("%d", ni.Value()), nil
}

// Emit a function call, and stack the definition for the postamble
func (gb *GolangBackend) compileProgn(progn gol.NodeProgn) (string, error) {
	if progn.Len() == 0 {
		return "", nil
	}

	first := true

	lines := []string{`func() int64 {
	var a int64
`}
	err := progn.Foreach(func(n gol.Node) error {
		if first {
			first = false
			return nil
		}
		s, err := gb.compile(n)
		if err != nil {
			return err
		}
		lines = append(lines, `a = `+s)
		return nil
	})
	if err != nil {
		return "", err
	}
	lines = append(lines, `
	return a
}()`)

	return strings.Join(lines, "\n"), nil
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

func (gb *GolangBackend) standardLib() string {
	return `
func __PLUS__(args ...int64) int64 {
	var sum int64
	for _, n := range args {
		sum += n
	}
	return sum
}

func __MINUS__(args ...int64) int64 {
	if len(args) < 2 {
		panic(fmt.Sprintf("Less than 2 args to numeric -"))
	}
	total := args[0]
	for _, n := range args[1:] {
		total -= n
	}
	return total
}

`
}

func newDefaultTypeEnv() typ.Env {
	e := typ.NewEnv()
	ints := []typ.Type{typ.Int, typ.Int}
	f := typ.Frame{
		"-": typ.NewFunc(ints, typ.Int),
		"+": typ.NewFunc(ints, typ.Int),
	}
	return e.WithFrame(f)
}
