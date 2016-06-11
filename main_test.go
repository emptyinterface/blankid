package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"
)

func TestBlankId(t *testing.T) {

	const (
		input = `
package main

type Struct struct {
	A int
}

func Add(a,b,c int) int {
	return a+b
}

type Example int

func (e Example) Add(a,b int) int {
	return int(e)+a
}

func (e Example) Sub(a,b int) int {
	return a-b
}

func (e Example) Empty() {}

func (*Example) NoReceiver() {}

func missingRet() (err error, ok bool) {
	ok = true
	return
}

func main() {
	func(a,b string) string {
		return a
	}("dog","cat")

	type recurse func(f recurse, n int)
	var r recurse
	r(func(f recurse, n int) {
		f(r)
	})

	func(s *Struct) {
		fmt.Println(s.A)
	} (&Struct{})
}

`
		output = `package main

type Struct struct {
	A int
}

func Add(a, b, _ int) int {
	return a + b
}

type Example int

func (e Example) Add(a, _ int) int {
	return int(e) + a
}

func (_ Example) Sub(a, b int) int {
	return a - b
}

func (_ Example) Empty()	{}

func (*Example) NoReceiver()	{}

func missingRet() (_ error, ok bool) {
	ok = true
	return
}

func main() {
	func(a, _ string) string {
		return a
	}("dog", "cat")

	type recurse func(f recurse, n int)
	var r recurse
	r(func(f recurse, _ int) {
		f(r)
	})

	func(s *Struct) {
		fmt.Println(s.A)
	}(&Struct{})
}
`
	)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", input, parser.ParseComments)
	if err != nil {
		t.Error(err)
	}

	// set all flags
	*receivers = true
	*returns = true

	if altered := blankId(f); !altered {
		t.Errorf("Expected altered to be true")
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, f); err != nil {
		t.Error(err)
	}

	if buf.String() != output {
		fmt.Println(buf.String())
	}

}
