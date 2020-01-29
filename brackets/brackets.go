package brackets

import (
	"container/list"
	"fmt"
	"strings"
)

// Brackets is a simple DOM-ish structure builder
type Brackets struct {
	elementStack *list.List
	elements     *list.List
}

// New - Brackets constructor
func New() *Brackets {
	return &Brackets{
		elementStack: list.New(),
		elements:     list.New(),
	}
}

// Attributes represents bracket attributes
type Attributes map[string]string

// Clone creates a deep copy of a map
func (am Attributes) Clone() Attributes {
	clone := Attributes{}
	for k, v := range am {
		clone[k] = v
	}
	return clone
}

func (am Attributes) String() string {
	attributes := []string{}

	for k, v := range am {
		attributes = append(attributes, fmt.Sprintf("%s=%q", k, v))
	}

	return strings.Join(attributes, " ")
}

type elementKind int

const (
	openingKind elementKind = iota
	closingKind
	selfClosingKind
	textKind
)

type element struct {
	name        string
	attributes  Attributes
	hasChildren bool
	kind        elementKind
}

func (e *element) String() string {
	if e.kind == textKind {
		return e.name
	}

	var builder strings.Builder
	builder.WriteRune('<')
	if e.kind == closingKind {
		builder.WriteRune('/')
	}
	builder.WriteString(e.name)

	if e.kind == selfClosingKind || e.kind == openingKind {
		if len(e.attributes) != 0 {
			builder.WriteRune(' ')
			builder.WriteString(e.attributes.String())
		}
		if e.kind == selfClosingKind {
			builder.WriteRune('/')
		}
	}
	builder.WriteRune('>')

	return builder.String()
}

func (b *Brackets) Add(name string, attrs ...Attributes) *Brackets {
	b.Open(name, attrs...)
	b.Close()
	return b
}

func (b *Brackets) topElement() *element {
	if b.elementStack.Len() != 0 {
		top := b.elementStack.Back()
		return top.Value.(*element)
	}
	return nil
}

func (b *Brackets) popElement() *element {
	top := b.elementStack.Back()
	b.elementStack.Remove(top)
	return top.Value.(*element)
}

func (b *Brackets) Open(name string, attrs ...Attributes) *Brackets {
	if top := b.topElement(); top != nil {
		top.hasChildren = true
	}
	newElement := &element{
		name: name,
		kind: openingKind,
	}
	if len(attrs) > 0 {
		newElement.attributes = attrs[0].Clone()
	}
	b.elementStack.PushBack(newElement)
	b.elements.PushBack(newElement)
	return b
}

func (b *Brackets) Text(txt string) *Brackets {
	if top := b.topElement(); top != nil {
		top.hasChildren = true
	}
	newElement := &element{
		name: txt,
		kind: textKind,
	}
	b.elements.PushBack(newElement)
	return b
}

func (b *Brackets) Close() *Brackets {
	top := b.popElement()
	if top.hasChildren {
		b.elements.PushBack(&element{
			name: top.name,
			kind: closingKind,
		})
	} else {
		top.kind = selfClosingKind
	}
	return b
}

func (b *Brackets) CloseAll() *Brackets {
	for b.elementStack.Len() > 0 {
		b.Close()
	}
	return b
}

func (b *Brackets) Current() string {
	top := b.topElement()
	if top != nil {
		return top.name
	}
	return ""
}

func (b *Brackets) Append(other *Brackets) *Brackets {
	b.elements.PushBackList(other.elements)
	return b
}

func (b *Brackets) String() string {
	var builder strings.Builder

	for e := b.elements.Front(); e != nil; e = e.Next() {
		builder.WriteString(e.Value.(*element).String())
	}

	return builder.String()
}
