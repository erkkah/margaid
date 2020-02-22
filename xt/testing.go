package xt

import "testing"

// XT is a testing.T - extension, adding a tiny bit
// of convenience, making tests more fun to write.
type XT struct {
	*testing.T
}

// X wraps a *testing.T and extends its functionality
func X(t *testing.T) XT {
	return XT{t}
}

// Assert verifies that a condition is true
func (x XT) Assert(cond bool, msg ...interface{}) {
	if !cond {
		x.Log(msg...)
		x.Fail()
	}
}

// True verifies that a condition is true
func (x XT) True(cond bool, msg ...interface{}) {
	if !cond {
		x.Log(msg...)
		x.Fail()
	}
}

// False verifies that a condition is false
func (x XT) False(cond bool, msg ...interface{}) {
	if cond {
		x.Log(msg...)
		x.Fail()
	}
}

// Equal verifies that the two arguments are equal
func (x XT) Equal(a interface{}, b interface{}, msg ...interface{}) {
	if a != b {
		x.Logf("%v should equal %v", a, b)
		x.Log(msg...)
		x.Fail()
	}
}

// NotEqual verifies that the two arguments are not equal
func (x XT) NotEqual(a interface{}, b interface{}, msg ...interface{}) {
	if a == b {
		x.Logf("%v should not equal %v", a, b)
		x.Log(msg...)
		x.Fail()
	}
}

// Nil verifies that the argument is nil
func (x XT) Nil(a interface{}, msg ...interface{}) {
	if a != nil {
		x.Log(msg...)
		x.Fail()
	}
}

// NotNil verifies that the argument is not nil
func (x XT) NotNil(a interface{}, msg ...interface{}) {
	if a == nil {
		x.Log(msg...)
		x.Fail()
	}
}
