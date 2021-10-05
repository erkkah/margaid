package margaid

import (
	"testing"
	"time"

	"github.com/erkkah/margaid/xt"
)

func TestTimeTickerStart(t *testing.T) {
	x := xt.X(t)

	min := SecondsFromTime(time.Date(2020, time.September, 4, 9, 10, 0, 0, time.Local))
	max := SecondsFromTime(time.Date(2020, time.September, 4, 9, 11, 0, 0, time.Local))

	m := New(100, 100, WithRange(
		XAxis,
		min,
		max,
	))
	ticker := m.TimeTicker(time.Kitchen)
	s := NewSeries()

	timestamp := time.Date(2020, time.September, 4, 9, 10, 30, 0, time.Local)
	s.Add(MakeValue(SecondsFromTime(timestamp), 123))
	start := ticker.start(XAxis, s, 10)

	// 60 secs in 10 steps leads to 6s step size
	x.Equal(start, min)
}

func TestTimeTickerStep(t *testing.T) {
	x := xt.X(t)

	min := SecondsFromTime(time.Date(2020, time.September, 4, 9, 10, 0, 0, time.Local))
	max := SecondsFromTime(time.Date(2020, time.September, 4, 9, 11, 0, 0, time.Local))

	m := New(100, 100, WithRange(
		XAxis,
		min,
		max,
	))
	ticker := m.TimeTicker(time.Kitchen)
	s := NewSeries()

	timestamp := time.Date(2020, time.September, 4, 9, 10, 30, 0, time.Local)
	s.Add(MakeValue(SecondsFromTime(timestamp), 123))
	step := ticker.start(XAxis, s, 10)

	more := true
	count := 0
	for ; step <= max && more; step, more = ticker.next(step) {
		count++
	}

	x.Equal(count, 11)
	x.Assert(more)
}

func TestValueTickerStart_Lin(t *testing.T) {
	x := xt.X(t)

	min := 12.34
	max := 56.78

	m := New(100, 100, WithRange(
		XAxis,
		min,
		max,
	))
	ticker := m.ValueTicker('f', 0, 10)
	s := NewSeries()

	s.Add(MakeValue(30, 123))
	start := ticker.start(XAxis, s, 10)

	x.Assert(start > min, "start > min")
	x.Assert(start < max, "start < max")
}

func TestValueTickerStep_Lin(t *testing.T) {
	x := xt.X(t)

	min := 12.34
	max := 56.78

	m := New(100, 100, WithRange(
		XAxis,
		min,
		max,
	))
	ticker := m.ValueTicker('f', 0, 10)
	s := NewSeries()

	s.Add(MakeValue(30, 123))
	step := ticker.start(XAxis, s, 10)

	more := true
	count := 0
	for ; step <= max && more; step, more = ticker.next(step) {
		count++
	}

	// There are four marks in the range [10.0, 50.0]
	x.Equal(count, 4)
	x.Assert(more)
}

func TestValueTickerStart_Log(t *testing.T) {
	x := xt.X(t)

	min := 12.34
	max := 56.78

	m := New(100, 100, WithRange(
		XAxis,
		min,
		max,
	), WithProjection(XAxis, Log))

	ticker := m.ValueTicker('f', 0, 10)
	s := NewSeries()

	s.Add(MakeValue(30, 123))
	start := ticker.start(XAxis, s, 10)

	x.Assert(start > min, "start > min")
	x.Assert(start < max, "start < max")
}

func TestValueTickerStep_Log(t *testing.T) {
	x := xt.X(t)

	min := 12.34
	max := 56.78

	m := New(100, 100, WithRange(
		XAxis,
		min,
		max,
	), WithProjection(XAxis, Log))
	ticker := m.ValueTicker('f', 0, 10)
	s := NewSeries()

	s.Add(MakeValue(30, 123))
	step := ticker.start(XAxis, s, 10)

	more := true
	count := 0
	for ; step <= max && more; step, more = ticker.next(step) {
		count++
	}

	// There are 4 base 10 marks in the range [20.0, 50.0]
	x.Equal(count, 4)
	x.Assert(more)
}

func TestValueTickerSimpleRange(t *testing.T) {
	x := xt.X(t)

	s := NewSeries()
	s.Add(MakeValue(1, 0))
	s.Add(MakeValue(2, 0))
	s.Add(MakeValue(3, 0))

	max := 1.0
	m := New(100, 100, WithAutorange(XAxis, s), WithAutorange(YAxis, s))
	ticker := m.ValueTicker('f', 0, 10)

	step := ticker.start(YAxis, s, 10)

	more := true
	count := 0
	for ; step <= max && more; step, more = ticker.next(step) {
		count++
	}

	// There are 5+1+5 base 10 marks in the range [-1.0, 1.0]
	x.Equal(count, 11)
	x.Assert(more)
}
