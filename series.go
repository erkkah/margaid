package margaid

import (
	"container/list"
	"math"
	"time"
)

// Series is the plottable type in Margaid
type Series struct {
	values *list.List
	min    float64
	max    float64

	capper Capper
}

// Option is the base type for all series options
type Option func(s *Series)

// NewSeries - series constructor
func NewSeries(options ...Option) *Series {
	self := &Series{
		values: list.New(),
	}

	for _, option := range options {
		option(self)
	}

	return self
}

// Size returns the current series value count
func (s *Series) Size() int {
	return s.values.Len()
}

// Min returns the series smallest y value, or 0.0 if
// the series is empty
func (s *Series) Min() float64 {
	return s.min
}

// Max returns the series largest y value, or 0.0 if
// the series is empty
func (s *Series) Max() float64 {
	return s.max
}

// Value is the type of each series element.
// The X part represents a position on the X axis, which could
// be time or a regular value.
type Value struct {
	X float64
	Y float64
}

// ValueAtTime creates a Value from a Y value and timestamp.
func ValueAtTime(y float64, timestamp time.Time) Value {
	x := float64(timestamp.UnixNano()) / 1E9
	return Value{
		X: x,
		Y: y,
	}
}

// MakeValue creates a Value from y and x values.
func MakeValue(y float64, x float64) Value {
	return Value{X: x, Y: y}
}

// GetXAsTime returns a Value X part as time.
func (v Value) GetXAsTime() time.Time {
	secs := int64(v.X)
	nanos := int64((v.X - float64(secs)) * 1E9)
	return time.Unix(secs, nanos)
}

// Add appends a value. If the series is capped, capping
// will be applied.
func (s *Series) Add(v Value) {
	if s.values.Len() == 0 {
		s.min = v.Y
		s.max = v.Y
	} else {
		s.min = math.Min(s.min, v.Y)
		s.max = math.Max(s.max, v.Y)
	}
	s.values.PushBack(v)
	if s.capper != nil {
		s.capper(s.values)
	}
}

// Capper is the capping function type
type Capper func(values *list.List)

// CappedBySize caps a series to at most cap values.
func CappedBySize(cap int) Option {
	return func(s *Series) {
		s.capper = func(values *list.List) {
			for values.Len() > cap {
				values.Remove(values.Front())
			}
		}
	}
}

// CappedByAge caps a series by removing values older than cap
// in relation to the current value of the reference funcion.
func CappedByAge(cap time.Duration, reference func() time.Time) Option {
	return func(s *Series) {
		s.capper = func(values *list.List) {
			for values.Len() > 0 {
				first := values.Front()
				val := first.Value.(Value)
				xTime := val.GetXAsTime()
				if !xTime.Before(reference().Add(-cap)) {
					break
				}
				values.Remove(first)
			}
		}
	}
}
