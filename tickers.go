package margaid

import (
	"math"
	"strconv"
	"time"

	"github.com/erkkah/margaid/svg"
)

// Ticker provides tick marks and labels for axes
type Ticker interface {
	label(value float64) string
	start(axis Axis, series *Series, steps int) float64
	next(previous float64) (next float64, hasMore bool)
}

// TimeTicker returns time valued tick labels in the specified time format.
// TimeTicker assumes that time is linear.
func (m *Margaid) TimeTicker(format string) Ticker {
	return &timeTicker{m, format, 1}
}

type timeTicker struct {
	m      *Margaid
	format string
	step   float64
}

func (t *timeTicker) label(value float64) string {
	formatted := TimeFromSeconds(value).Format(t.format)
	return svg.EncodeText(formatted, svg.HAlignMiddle)
}

func (t *timeTicker) start(axis Axis, series *Series, steps int) float64 {
	minmax := t.m.ranges[axis]
	scaleRange := minmax.max - minmax.min
	scaleDuration := TimeFromSeconds(scaleRange).Sub(time.Unix(0, 0))

	t.step = math.Pow(10.0, math.Trunc(math.Log10(scaleDuration.Seconds()/float64(steps))))
	for int(scaleRange/t.step) > steps {
		t.step += 1.0
	}
	start := minmax.min - math.Mod(minmax.min, t.step) + t.step
	return start
}

func (t *timeTicker) next(previous float64) (float64, bool) {
	return previous + t.step, true
}

// ValueTicker returns tick labels by converting floats using strconv.FormatFloat
func (m *Margaid) ValueTicker(style byte, precision int, base int) Ticker {
	return &valueTicker{
		m:         m,
		step:      1,
		style:     style,
		precision: precision,
		base:      base,
	}
}

type valueTicker struct {
	m          *Margaid
	projection Projection
	step       float64
	style      byte
	precision  int
	base       int
}

func (t *valueTicker) label(value float64) string {
	return strconv.FormatFloat(value, t.style, t.precision, 64)
}

func (t *valueTicker) start(axis Axis, series *Series, steps int) float64 {
	t.projection = t.m.projections[axis]
	minmax := t.m.ranges[axis]
	scaleRange := minmax.max - minmax.min

	startValue := 0.0
	floatBase := float64(t.base)

	if t.projection == Lin {
		roundedLog := math.Floor(math.Log(scaleRange/float64(steps)) / math.Log(floatBase))
		t.step = math.Pow(floatBase, roundedLog)

		for int(scaleRange/t.step) > steps {
			t.step += 1.0
		}
		startValue := minmax.min - math.Mod(minmax.min, t.step) + t.step
		return startValue
	}

	t.step = 0
	startValue = math.Pow(floatBase, math.Round(math.Log(minmax.min)/math.Log(floatBase)))
	for startValue < minmax.min {
		startValue, _ = t.next(startValue)
	}
	return startValue
}

func (t *valueTicker) next(previous float64) (float64, bool) {
	if t.projection == Lin {
		return previous + t.step, true
	}

	floatBase := float64(t.base)
	log := math.Log(previous) / math.Log(floatBase)
	if log < 0 {
		log = -math.Ceil(-log)
	} else {
		log = math.Floor(log)
	}
	increment := math.Pow(floatBase, log)
	next := previous + increment
	next /= increment
	next = math.Round(next) * increment
	return next, true
}

// LabeledTicker places tick marks and labels for all values
// of a series. The labels are provided by the labeler function.
func (m *Margaid) LabeledTicker(labeler func(float64) string) Ticker {
	return &labeledTicker{
		labeler: labeler,
	}
}

type labeledTicker struct {
	labeler func(float64) string
	values  []float64
	index   int
}

func (t *labeledTicker) label(value float64) string {
	return svg.EncodeText(t.labeler(value), svg.HAlignMiddle)
}

func (t *labeledTicker) start(axis Axis, series *Series, steps int) float64 {
	var values []float64

	var get func(v Value) float64

	if axis == X1Axis || axis == X2Axis {
		get = func(v Value) float64 {
			return v.X
		}
	}
	if axis == Y1Axis || axis == Y2Axis {
		get = func(v Value) float64 {
			return v.Y
		}
	}

	v := series.Values()
	for v.Next() {
		val := v.Get()
		values = append(values, get(val))
	}

	t.values = values
	return values[0]
}

func (t *labeledTicker) next(previous float64) (float64, bool) {
	// Tickers are not supposed to have state.
	// LabeledTicker breaks this, assuming a strict linear calling order.

	if previous != t.values[t.index] {
		return previous, false
	}

	t.index++
	if t.index < len(t.values) {
		return t.values[t.index], true
	}
	t.index = 0
	return 0, false
}
