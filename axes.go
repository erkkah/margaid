package margaid

import (
	"fmt"
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
	start(axis Axis, steps int) float64
	next(previous float64) float64
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
	return TimeFromSeconds(value).Format(t.format)
}

func (t *timeTicker) start(axis Axis, steps int) float64 {
	minmax := t.m.ranges[axis]
	scaleRange := minmax.max - minmax.min
	scaleDuration := TimeFromSeconds(scaleRange).Sub(time.Unix(0, 0))

	t.step = math.Pow(10.0, math.Trunc(math.Log10(scaleDuration.Seconds()/float64(steps))))
	for int(scaleRange/t.step) > steps {
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

func (t *valueTicker) start(axis Axis, steps int) float64 {
	t.projection = t.m.projections[axis]
	minmax := t.m.ranges[axis]
	scaleRange := minmax.max - minmax.min

	startValue := 0.0
	floatBase := float64(t.base)

	if t.projection == Lin {
		roundedLog := math.Round(math.Log(scaleRange/float64(steps)) / math.Log(floatBase))
		t.step = math.Pow(floatBase, roundedLog)

		for int(scaleRange/t.step) > steps {
			t.step *= 2
		}
		startValue = t.step
		return startValue
	}

	t.step = 0
	startValue = math.Pow(floatBase, math.Round(math.Log(minmax.min)/math.Log(floatBase)))
	for startValue < minmax.min {
		startValue = t.next(startValue)
	}
	return startValue
}

func (t *valueTicker) next(previous float64) float64 {
	if t.projection == Lin {
		return previous + t.step
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
	return next
}

// Axis draws tick marks and labels using the specified ticker
func (m *Margaid) Axis(series *Series, axis Axis, ticker Ticker, grid bool) {
	var xOffset float64 = m.inset
	var yOffset float64 = m.inset
	var axisLength float64
	var crossLength float64
	var xMult float64 = 0
	var yMult float64 = 0
	var tickSign float64 = 1
	var vAlignment svg.VAlignment
	var hAlignment svg.HAlignment

	max := m.ranges[axis].max

	xAttributes := func() {
		axisLength = m.width - 2*m.inset
		crossLength = m.height - 2*m.inset
		xMult = 1
		hAlignment = svg.HAlignMiddle
	}

	yAttributes := func() {
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

	const tickDistance = 55
	steps := axisLength / tickDistance
	start := ticker.start(axis, int(steps))

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, -1),
	).
		StrokeWidth("2px").
		Stroke("black")

	tick := start

	for tick <= max {
		value, err := m.project(tick, axis)

		if err == nil {
			m.g.Polyline([]struct{ X, Y float64 }{
				{value * xMult, value * yMult},
				{value*xMult + tickSign*6*(1-xMult), value*yMult + tickSign*(1-yMult)*6},
			}...)
		}

		tick = ticker.next(tick)
	}

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, 1),
	).
		Font(m.labelFamily, fmt.Sprintf("%dpx", m.labelSize)).
		FontStyle(svg.StyleNormal, svg.WeightLighter).
		Alignment(hAlignment, vAlignment).
		Fill("black")

	tick = start
	lastLabel := -m.inset

	for tick <= max {
		value, err := m.project(tick, axis)

		if err == nil {
			if value-lastLabel > float64(m.labelSize) {
				m.g.Text(
					value*xMult+(tickSign)*10*(1-xMult),
					-value*yMult+(-tickSign)*10*(1-yMult),
					ticker.label(tick))
				lastLabel = value
			}
		}

		tick = ticker.next(tick)
	}

	if grid {
		m.g.Transform(
			svg.Translation(xOffset, yOffset),
			svg.Scaling(1, -1),
		).
			StrokeWidth("0.5px").Stroke("gray")

		tick = start

		for tick <= max {
			value, err := m.project(tick, axis)

			if err == nil {
				m.g.Polyline([]struct{ X, Y float64 }{
					{value * xMult, value * yMult},
					{value*xMult - tickSign*crossLength*(1-xMult), value*yMult - tickSign*(1-yMult)*crossLength},
				}...)
			}

			tick = ticker.next(tick)
		}
	}
}
