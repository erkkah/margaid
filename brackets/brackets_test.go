package brackets

import (
	"testing"

	"github.com/erkkah/margaid/xt"
)

func TestBracketConstruction(t *testing.T) {
	x := xt.X(t)
	b := New()
	x.Equal(b.String(), "", "Should be empty string")
}

func TestEmptySelfClosingBracket(t *testing.T) {
	x := xt.X(t)
	b := New()
	b.Add("tag")
	x.Equal(b.String(), "<tag/>")
}

func TestSelfClosingBracket(t *testing.T) {
	x := xt.X(t)
	b := New()
	b.Add("tag", Attributes{
		"size": "22",
	})
	x.Equal(b.String(), `<tag size="22"/>`)
}

func TestOpenCloseBracket(t *testing.T) {
	x := xt.X(t)
	b := New()
	b.Open("tag")
	b.Close()
	x.Equal(b.String(), "<tag/>")
}

func TestNestedBrackets(t *testing.T) {
	x := xt.X(t)
	b := New()
	b.Open("head", Attributes{
		"alpha": "beta",
	})
	b.Open("title", Attributes{
		"text": "Hej",
	})
	b.CloseAll()
	x.Equal(b.String(), `<head alpha="beta"><title text="Hej"/></head>`)
}
