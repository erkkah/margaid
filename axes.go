package margaid

import (
	"math"
	"strconv"
	"time"

	"github.com/erkkah/margaid/svg"
)

// Axis is the type for all axis constants
type Axis int

// Axis constants
const (
	XAxis Axis = iota + 'x'
	YAxis
	X2Axis
	Y2Axis

	X1Axis = XAxis
	Y1Axis = YAxis
)

// Ticker provides tick marks and labels for axes
type Ticker interface {
	label(value float64) string
	start(axis Axis, valueRange float64, steps int) float64
	next(previous float64) float64
}

// TimeTicker returns time valued tick labels in the specified time format.
// TimeTicker assumes that time is linear.
func TimeTicker(format string) Ticker {
	return &timeTicker{format, 1}
}

type timeTicker struct {
	format string
	step   float64
}

func (t *timeTicker) label(value float64) string {
	return TimeFromSeconds(value).Format(t.format)
}

func (t *timeTicker) start(axis Axis, valueRange float64, steps int) float64 {
	scaleDuration := TimeFromSeconds(valueRange).Sub(time.Unix(0, 0))

	t.step = math.Pow(10.0, math.Trunc(math.Log10(scaleDuration.Seconds()/float64(steps))))
	for int(valueRange/t.step) > steps {
		t.step *= 2
	}
	return t.step
}

func (t *timeTicker) next(previous float64) float64 {
	return previous + t.step
}

// ValueTicker returns tick labels by converting floats using strconv.FormatFloat
func (m *Margaid) ValueTicker(style byte, precision int, base int) Ticker {
	return &valueTicker{
		m:         m,
		step:      1,
		scale:     1,
		style:     style,
		precision: precision,
		base:      base,
	}
}

type valueTicker struct {
	m         *Margaid
	step      float64
	scale     float64
	style     byte
	precision int
	base      int
}

func (t *valueTicker) label(value float64) string {
	return strconv.FormatFloat(value, t.style, t.precision, 64)
}

func (t *valueTicker) start(axis Axis, valueRange float64, steps int) float64 {
	projection := t.m.projections[axis]

	startValue := 0.0
	floatBase := float64(t.base)

	if projection == Lin {
		truncatedLog := math.Trunc(math.Log(valueRange/float64(steps)) / math.Log(floatBase))
		t.step = math.Pow(floatBase, truncatedLog)

		for int(valueRange/t.step) > steps {
			t.step *= 2
		}
		startValue = t.step
		return startValue
	}

	roundedLog := math.Ceil((math.Log(valueRange) / math.Log(floatBase)) / float64(steps))
	t.scale = math.Pow(floatBase, math.Max(1, roundedLog))
	t.step = 0
	startValue = math.Pow(floatBase, roundedLog-1)
	return startValue
}

func (t *valueTicker) next(previous float64) float64 {
	if t.step != 0 {
		return previous + t.step
	}

	floatBase := float64(t.base)
	truncatedLog := math.Trunc(math.Log(previous) / math.Log(floatBase))
	next := previous + math.Pow(floatBase, truncatedLog)
	return next
}

// Axis draws tick marks and labels using the specified ticker
func (m *Margaid) Axis(series *Series, axis Axis, ticker Ticker, grid bool) {
	var min float64
	var max float64
	var xOffset float64 = m.inset
	var yOffset float64 = m.inset
	var axisLength float64
	var crossLength float64
	var xMult float64 = 0
	var yMult float64 = 0
	var tickSign float64 = 1
	var vAlignment svg.VAlignment
	var hAlignment svg.HAlignment

	xAttributes := func() {
		min = series.MinX()
		max = series.MaxX()
		axisLength = m.width - 2*m.inset
		crossLength = m.height - 2*m.inset
		xMult = 1
		hAlignment = svg.HAlignMiddle
	}

	yAttributes := func() {
		min = series.MinY()
		max = series.MaxY()
		yOffset = m.height - m.inset
		axisLength = m.height - 2*m.inset
		crossLength = m.width - 2*m.inset
		yMult = 1
		vAlignment = svg.VAlignCentral
	}

	switch axis {
	case X1Axis:
		xAttributes()
		yOffset = m.height - m.inset
		tickSign = -1
		vAlignment = svg.VAlignTop
	case X2Axis:
		xAttributes()
		vAlignment = svg.VAlignBottom
	case Y1Axis:
		yAttributes()
		tickSign = -1
		hAlignment = svg.HAlignEnd
	case Y2Axis:
		yAttributes()
		xOffset = m.width - m.inset
		hAlignment = svg.HAlignStart
	}

	const tickDistance = 75
	steps := axisLength / tickDistance
	step := ticker.start(axis, max-min, int(steps))

	tick := min
	if math.Mod(tick, step) != 0 {
		tick += step
		tick -= math.Mod(tick, step)
	}

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, -1),
	).
		StrokeWidth("2px").
		Stroke("black")

	firstTick := tick

	for tick <= max {
		// ??? Ignore error :(
		value, _ := m.project(tick, axis)
		m.g.Polyline([]struct{ X, Y float64 }{
			{value * xMult, value * yMult},
			{value*xMult + tickSign*6*(1-xMult), value*yMult + tickSign*(1-yMult)*6},
		}...)

		tick = ticker.next(tick)
	}

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, 1),
	).
		Font("sans-serif", "10pt").
		FontStyle(svg.StyleNormal, svg.WeightLighter).
		Alignment(hAlignment, vAlignment)

	tick = firstTick
	lastLabel := -m.inset
	const labelThreshold = 10.0

	for tick <= max {
		// ??? Ignore error :(
		value, _ := m.project(tick, axis)

		if value-lastLabel > labelThreshold {
			m.g.Text(
				value*xMult+(tickSign)*10*(1-xMult),
				-value*yMult+(-tickSign)*10*(1-yMult),
				ticker.label(tick))
			lastLabel = value
		}

		tick = ticker.next(tick)
	}

	if grid {
		m.g.Transform(
			svg.Translation(xOffset, yOffset),
			svg.Scaling(1, -1),
		).
			StrokeWidth("0.5px").Stroke("gray")
		tick = firstTick

		for tick < max {
			// ??? Ignore error :(
			value, _ := m.project(tick, axis)
			m.g.Polyline([]struct{ X, Y float64 }{
				{value * xMult, value * yMult},
				{value*xMult - tickSign*crossLength*(1-xMult), value*yMult - tickSign*(1-yMult)*crossLength},
			}...)

			tick = ticker.next(tick)
		}
	}
}
