package margaid

import (
	"testing"
	"time"
)

func TestConstruction(t *testing.T) {
	s := NewSeries()
	s.Add(MakeValue(0, 1))
	if s.Size() != 1 {
		t.Fail()
	}
}

func TestCapBySize(t *testing.T) {
	s := NewSeries(CappedBySize(3))
	for i := 0; i < 10; i++ {
		s.Add(MakeValue(float64(i*2), float64(i)))
	}
	if s.Size() != 3 {
		t.Fail()
	}
}

func TestCapByTime(t *testing.T) {
	now := time.Now()
	reference := func() time.Time {
		return now
	}
	s := NewSeries(CappedByAge(5*time.Millisecond, reference))
	for i := 0; i < 10; i++ {
		s.Add(MakeValue(SecondsFromTime(now.Add(-2*time.Duration(10-i)*time.Millisecond)), float64(i)))
	}
	if s.Size() != 2 {
		t.Fail()
	}
}

func TestMinMax(t *testing.T) {
	s := NewSeries()
	s.Add(MakeValue(0, 1))
	s.Add(MakeValue(1000, 2))
	s.Add(MakeValue(-10, 3))
	if s.MinX() != -10 {
		t.Fail()
	}
	if s.MaxX() != 1000 {
		t.Fail()
	}
}
