package margaid

import (
	"testing"
	"time"

	"github.com/erkkah/margaid/xt"
)

func TestConstruction(t *testing.T) {
	x := xt.X(t)
	s := NewSeries()
	s.Add(MakeValue(0, 1))

	x.Equal(s.Size(), 1)
}

func TestCapBySize(t *testing.T) {
	x := xt.X(t)
	s := NewSeries(CappedBySize(3))
	for i := 0; i < 10; i++ {
		s.Add(MakeValue(float64(i*2), float64(i)))
	}
	x.Equal(s.Size(), 3)
}

func TestCapByTime(t *testing.T) {
	x := xt.X(t)
	now := time.Now()
	reference := func() time.Time {
		return now
	}
	s := NewSeries(CappedByAge(5*time.Millisecond, reference))
	for i := 0; i < 10; i++ {
		s.Add(MakeValue(SecondsFromTime(now.Add(-2*time.Duration(10-i)*time.Millisecond)), float64(i)))
	}
	x.Equal(s.Size(), 2)
	x.Assert(TimeFromSeconds(s.MinX()).After(now.Add(-5 * time.Millisecond)))
}

func TestMinMax(t *testing.T) {
	x := xt.X(t)

	s := NewSeries()
	s.Add(MakeValue(0, 1))
	s.Add(MakeValue(1000, 2))
	s.Add(MakeValue(-10, 3))

	x.Equal(s.MinX(), -10.0)
	x.Equal(s.MaxX(), 1000.0)
}

func TestAggregateAvg(t *testing.T) {
	x := xt.X(t)

	s := NewSeries(AggregatedBy(Avg, time.Second))

	now := time.Now().Truncate(time.Second).Add(time.Millisecond * 50)
	later := now.Add(time.Millisecond * 100)
	tooLate := now.Add(time.Second)

	s.Add(
		MakeValue(SecondsFromTime(now), 10),
		MakeValue(SecondsFromTime(later), 20),
		MakeValue(SecondsFromTime(tooLate), 30),
	)

	values := s.Values()
	x.True(values.Next())

	v := values.Get()
	x.Equal(v.Y, 15.0)
	x.Equal(v.X, SecondsFromTime(now.Truncate(time.Second)))

	x.False(values.Next(), "Series should be empty")
}

func TestAggregateSum(t *testing.T) {
	x := xt.X(t)

	s := NewSeries(AggregatedBy(Sum, time.Second))

	now := time.Now().Truncate(time.Second).Add(time.Millisecond * 50)
	later := now.Add(time.Millisecond * 100)
	tooLate := now.Add(time.Second)

	s.Add(
		MakeValue(SecondsFromTime(now), 10),
		MakeValue(SecondsFromTime(later), 20),
		MakeValue(SecondsFromTime(tooLate), 30),
	)

	values := s.Values()
	x.True(values.Next())

	v := values.Get()
	x.Equal(v.Y, 30.0)
	x.Equal(v.X, SecondsFromTime(now.Truncate(time.Second)))

	x.False(values.Next(), "Series should be empty")
}

func TestAggregateDelta(t *testing.T) {
	x := xt.X(t)

	s := NewSeries(AggregatedBy(Delta, time.Second))

	now := time.Now().Truncate(time.Second).Add(time.Millisecond * 50)
	later := now.Add(time.Millisecond * 100)
	tooLate := now.Add(time.Second)

	s.Add(
		MakeValue(SecondsFromTime(now), 10),
		MakeValue(SecondsFromTime(later), 20),
		MakeValue(SecondsFromTime(tooLate), 30),
	)

	values := s.Values()
	x.True(values.Next())

	v := values.Get()
	x.Equal(v.Y, 10.0)
	x.Equal(v.X, SecondsFromTime(now.Truncate(time.Second)))

	x.False(values.Next(), "Series should be empty")
}
