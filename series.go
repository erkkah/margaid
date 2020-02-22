package margaid

import (
	"container/list"
	"math"
	"time"
)

// Series is the plottable type in Margaid
type Series struct {
	values *list.List
	minX   float64
	maxX   float64
	minY   float64
	maxY   float64

	title string

	capper     Capper
	aggregator Aggregator
	interval   time.Duration
	buffer     []Value
	at         time.Time
}

// SeriesOption is the base type for all series options
type SeriesOption func(s *Series)

// NewSeries - series constructor
func NewSeries(options ...SeriesOption) *Series {
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

// MinX returns the series smallest x value, or 0.0 if
// the series is empty
func (s *Series) MinX() float64 {
	return s.minX
}

// MaxX returns the series largest x value, or 0.0 if
// the series is empty
func (s *Series) MaxX() float64 {
	return s.maxX
}

// MinY returns the series smallest y value, or 0.0 if
// the series is empty
func (s *Series) MinY() float64 {
	return s.minY
}

// MaxY returns the series largest y value, or 0.0 if
// the series is empty
func (s *Series) MaxY() float64 {
	return s.maxY
}

// SeriesIterator helps iterating series values
type SeriesIterator struct {
	list    *list.List
	element *list.Element
}

// Get returns the iterator current value.
// A newly created iterator has no current value.
func (si *SeriesIterator) Get() Value {
	return si.element.Value.(Value)
}

// Next steps to the next value.
// A newly created iterator has no current value.
func (si *SeriesIterator) Next() bool {
	if si.element == nil {
		si.element = si.list.Front()
	} else {
		si.element = si.element.Next()
	}
	return si.element != nil
}

// Values returns an iterator to the series values
func (s *Series) Values() SeriesIterator {
	return SeriesIterator{
		list: s.values,
	}
}

// Value is the type of each series element.
// The X part represents a position on the X axis, which could
// be time or a regular value.
type Value struct {
	X float64
	Y float64
}

// MakeValue creates a Value from x and y values.
func MakeValue(x float64, y float64) Value {
	return Value{X: x, Y: y}
}

// Add appends one or more values, optionally
// peforming aggregation.
// If the series is capped, capping will be applied
// after aggregation.
func (s *Series) Add(values ...Value) {
	if len(values) == 0 {
		return
	}

	if s.aggregator != nil {

		if s.values.Len() == 0 && len(s.buffer) == 0 {
			s.at = TimeFromSeconds(values[0].X).Truncate(s.interval)
		}

		var aggregated []Value

		for _, v := range values {
			at := TimeFromSeconds(v.X)
			if s.at.Add(s.interval).Before(at) {
				agg := s.aggregator(s.buffer, s.at)
				aggregated = append(aggregated, agg)
				s.at = at.Truncate(s.interval)
				s.buffer = append(s.buffer[0:0], v)
			} else {
				s.buffer = append(s.buffer, v)
			}
		}

		values = aggregated
	}

	for _, v := range values {
		if s.values.Len() == 0 {
			s.minX = v.X
			s.maxX = v.X
			s.minY = v.Y
			s.maxY = v.Y
		} else {
			s.minX = math.Min(s.minX, v.X)
			s.maxX = math.Max(s.maxX, v.X)
			s.minY = math.Min(s.minY, v.Y)
			s.maxY = math.Max(s.maxY, v.Y)
		}
		s.values.PushBack(v)
		if s.capper != nil {
			s.capper(s.values)
		}
	}
}

// Zip merges two slices of floats into pairs and adds
// them to the series. It is assumed that the two slices
// have the same length.
func (s *Series) Zip(xValues, yValues []float64) {
	valueCount := len(xValues)
	if len(yValues) < valueCount {
		valueCount = len(yValues)
	}

	zipped := make([]Value, valueCount)
	for i := range zipped {
		zipped[i] = MakeValue(xValues[i], yValues[i])
	}
	s.Add(zipped...)
}

func (s *Series) updateMinMax() {
	if s.values.Len() == 0 {
		return
	}

	values := s.Values()
	values.Next()
	current := values.Get()
	s.minX = current.X
	s.maxX = current.X
	s.minY = current.Y
	s.maxY = current.Y

	for values.Next() {
		current = values.Get()
		s.minX = math.Min(s.minX, current.X)
		s.maxX = math.Max(s.maxX, current.X)
		s.minY = math.Min(s.minY, current.Y)
		s.maxY = math.Max(s.maxY, current.Y)
	}
}

// Capper is the capping function type
type Capper func(values *list.List)

// Aggregator is the aggregating function type
type Aggregator func(values []Value, at time.Time) Value

// CappedBySize caps a series to at most cap values.
func CappedBySize(cap int) SeriesOption {
	return func(s *Series) {
		s.capper = func(values *list.List) {
			removed := false
			for values.Len() > cap {
				values.Remove(values.Front())
				removed = true
			}
			if removed {
				s.updateMinMax()
			}
		}
	}
}

// CappedByAge caps a series by removing values older than cap
// in relation to the current value of the reference funcion.
func CappedByAge(cap time.Duration, reference func() time.Time) SeriesOption {
	return func(s *Series) {
		s.capper = func(values *list.List) {
			removed := false
			for values.Len() > 0 {
				first := values.Front()
				val := first.Value.(Value)
				xTime := TimeFromSeconds(val.X)
				if !xTime.Before(reference().Add(-cap)) {
					break
				}
				values.Remove(first)
				removed = true
			}
			if removed {
				s.updateMinMax()
			}
		}
	}
}

// Avg calculates the Y average of a list of values,
// reported as observed at the given time.
func Avg(values []Value, at time.Time) Value {
	if len(values) == 0 {
		return Value{}
	}

	var sum float64
	for _, v := range values {
		sum += v.Y
	}
	avg := sum / float64(len(values))
	return Value{SecondsFromTime(at), avg}
}

// Sum calculates the Y sum of a list of values,
// reported as observed at the given time.
func Sum(values []Value, at time.Time) Value {
	var sum float64
	for _, v := range values {
		sum += v.Y
	}
	return Value{SecondsFromTime(at), sum}
}

// Delta calculates the Y difference between the first and
// last value in the list, reported as observed at the given time.
func Delta(values []Value, at time.Time) Value {
	if len(values) < 2 {
		return Value{}
	}

	first := values[0].Y
	last := values[len(values)-1].Y
	return Value{SecondsFromTime(at), last - first}
}

// AggregatedBy sets the series aggregator
func AggregatedBy(f Aggregator, interval time.Duration) SeriesOption {
	return func(s *Series) {
		s.aggregator = f
		s.interval = interval
	}
}

// Titled sets the series title
func Titled(title string) SeriesOption {
	return func(s *Series) {
		s.title = title
	}
}

// TimeFromSeconds converts from seconds since the epoch to time.Time
func TimeFromSeconds(seconds float64) time.Time {
	wholeSecs := int64(seconds)
	nanos := int64((seconds - float64(wholeSecs)) * 1E9)
	return time.Unix(wholeSecs, nanos)
}

// SecondsFromTime converts from time.Time to seconds since the epoch
func SecondsFromTime(time time.Time) float64 {
	return float64(time.UnixNano()) / 1E9
}
