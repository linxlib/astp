// maybe some package description
package main

// this is a test golang file
// and this is second line comment
// @copyright linx

import (
	"github.com/linxlib/astp/example/testpackage"
	"net/http"
	http3 "net/http"
)

var (
	// test1
	test1, test2 int    = 1, 2 //hhhhh
	name         string = "1111"
	name1               = ""
	b                   = 11
	ccc          string
	f            testpackage.Fuck
	ff           *testpackage.Fuck
	fff          *string
	fffs         []string
	// fffa hhhh
	fffa *[]string  //fffa
	fffb *[]*string //fffb
	ma   map[string]*int
	mb   map[string]string
)

// const uhi = ""
// const ijgh string = ""
// const (
//
//	ijghfg     = "okkk"
//	uhdfdf     = 90
//	uhfdf  int = 8
//
// )
const (
	User = iota + 1 //iota
	Pass            //password
	Fake            //fake
)

var Cc = 2

type A struct {
	testpackage.Model
	Hu      string `json:"hu"`
	Kapi    string `k:"kkk"`
	Header  http.Header
	Header2 http3.Header
}

type B struct {
	A
}
type C struct {
	*A
}

func TestMethod(a *A) (i *A, err error) {
	return nil, nil
}
func TestMethod1(a ...string) (i *A, err error) {
	return nil, nil
}

// some comment in middle of file

func (receiver A) name(a string, b int) string {
	testpackage.TestMethod2()
	return ""
}

// NameA hhhh
func (receiver *A) NameA(a, b string) (rtn *string) {
	dd := ""
	return &dd
}
func (receiver *A) NameB(a, b string) (rtn, rtn2 *string, rtn3 int) {
	dd := ""
	return &dd, &dd, -1
}

// IInterface sss
type IInterface interface {
	//Get Name
	GetName() string //pu
	//Get by id
	GetBy(id int) bool // get
}

func TestMethod2(somemap map[string]A) (i *A, err error) {
	return nil, nil
}
