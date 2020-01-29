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
	start(valueRange float64, steps int) float64
	next(previous float64) float64
}

// TimeTicker returns time valued tick labels in the specified time format.
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

func (t *timeTicker) start(valueRange float64, steps int) float64 {
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
func ValueTicker(value Value, style byte, precision int) string {
	return strconv.FormatFloat(float64(value.X), style, precision, 32)
}

// Axis draws tick marks and labels using the specified ticker
func (m *Margaid) Axis(series *Series, axis Axis, ticker Ticker) {
	var min float64
	var max float64
	var xOffset float64 = m.inset
	var yOffset float64 = m.inset
	var axisLength float64
	var xMult float64 = 0
	var yMult float64 = 0
	var tickSign float64 = 1
	var vAlignment svg.VAlignment
	var hAlignment svg.HAlignment

	switch axis {
	case X1Axis:
		min = series.MinX()
		max = series.MaxX()
		yOffset = m.height - m.inset
		axisLength = m.width
		xMult = 1
		tickSign = -1
		hAlignment = svg.HAlignMiddle
		vAlignment = svg.VAlignTop
	case X2Axis:
		min = series.MinX()
		max = series.MaxX()
		axisLength = m.width
		xMult = 1
		hAlignment = svg.HAlignMiddle
		vAlignment = svg.VAlignBottom
	case Y1Axis:
		min = series.MinY()
		max = series.MaxY()
		yOffset = m.height - m.inset
		axisLength = m.height
		yMult = 1
		tickSign = -1
		hAlignment = svg.HAlignEnd
		vAlignment = svg.VAlignCentral
	case Y2Axis:
		min = series.MinY()
		max = series.MaxY()
		xOffset = m.width - m.inset
		yOffset = m.height - m.inset
		axisLength = m.height
		yMult = 1
		vAlignment = svg.VAlignCentral
		hAlignment = svg.HAlignStart
	}

	const tickDistance = 75
	steps := axisLength / tickDistance
	start := ticker.start(max-min, int(steps))

	tick := min + start
	tick -= math.Mod(tick, start)

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, -1),
	).
		StrokeWidth("2px").
		Stroke("black")

	firstTick := tick

	for tick < max {
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
		Stroke("none").
		Font("sans-serif", "10pt").
		FontStyle(svg.StyleNormal, svg.WeightLighter).
		Alignment(hAlignment, vAlignment)

	tick = firstTick

	for tick < max {
		// ??? Ignore error :(
		value, _ := m.project(tick, axis)
		m.g.Text(
			value*xMult+(tickSign)*10*(1-xMult),
			-value*yMult+(-tickSign)*10*(1-yMult),
			ticker.label(tick))

		tick = ticker.next(tick)
	}
}
