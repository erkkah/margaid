package brackets

import (
	"bytes"
	"container/list"
	"encoding/xml"
	"fmt"
	"strings"
)

// Brackets is a simple xml-ish structure builder
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

// Size calculates the total size in bytes of all keys + values
func (am Attributes) Size() (size int) {
	for k, v := range am {
		size += len([]byte(k)) + len([]byte(v))
	}
	return
}

type elementKind int

const (
	openingKind elementKind = iota
	closingKind
	selfClosingKind
	textKind
)

// Element represents an xml element
type Element struct {
	name        string
	attributes  Attributes
	hasChildren bool
	kind        elementKind
}

func (e *Element) String() string {
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

// Attributes returns a copy of the element attributes
func (e *Element) Attributes() Attributes {
	return e.attributes.Clone()
}

// Name returns the element name
func (e *Element) Name() string {
	return e.name
}

// SetAttribute resets an attribute to a new value
func (e *Element) SetAttribute(key, value string) {
	e.attributes[key] = value
}

func (b *Brackets) topElement() *Element {
	if b.elementStack.Len() != 0 {
		top := b.elementStack.Back()
		return top.Value.(*Element)
	}
	return nil
}

func (b *Brackets) popElement() *Element {
	top := b.elementStack.Back()
	b.elementStack.Remove(top)
	return top.Value.(*Element)
}

// Open adds a new opening element with optional attributes.
func (b *Brackets) Open(name string, attrs ...Attributes) *Brackets {
	if top := b.topElement(); top != nil {
		top.hasChildren = true
	}
	newElement := &Element{
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

// Add adds a self-closing element
func (b *Brackets) Add(name string, attrs ...Attributes) *Brackets {
	b.Open(name, attrs...)
	b.Close()
	return b
}

// Text adds text content to the current element.
// Note that the text is not automatically XML-escaped.
// Use the XMLEscape function if that is needed.
func (b *Brackets) Text(txt string) *Brackets {
	if top := b.topElement(); top != nil {
		top.hasChildren = true
	}
	newElement := &Element{
		name: txt,
		kind: textKind,
	}
	b.elements.PushBack(newElement)
	return b
}

// Close closes the current element. If there are no
// children (elements or text), the current element will
// be self-closed. Otherwise a matching close-element
// will be added.
func (b *Brackets) Close() *Brackets {
	top := b.popElement()
	if top.hasChildren {
		b.elements.PushBack(&Element{
			name: top.name,
			kind: closingKind,
		})
	} else {
		top.kind = selfClosingKind
	}
	return b
}

// CloseAll closes all elements to get a complete, matched
// structure.
func (b *Brackets) CloseAll() *Brackets {
	for b.elementStack.Len() > 0 {
		b.Close()
	}
	return b
}

// Current returns a pointer to the currently open element, or
// nil if there is none.
func (b *Brackets) Current() *Element {
	return b.topElement()
}

// Last returns a pointer to the most recently added element, or
// nil if there is none.
func (b *Brackets) Last() *Element {
	if b.elements.Len() != 0 {
		last := b.elements.Back()
		return last.Value.(*Element)
	}
	return nil
}

// Append adds all elements from another Brackets instance
func (b *Brackets) Append(other *Brackets) *Brackets {
	b.elements.PushBackList(other.elements)
	return b
}

func (b *Brackets) String() string {
	var builder strings.Builder

	for e := b.elements.Front(); e != nil; e = e.Next() {
		builder.WriteString(e.Value.(*Element).String())
	}

	return builder.String()
}

// XMLEscape returns properly escaped XML equivalent of the provided string
func XMLEscape(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}
