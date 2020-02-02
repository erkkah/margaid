package margaid

import (
	"fmt"

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

// Axis draws tick marks and labels using the specified ticker
func (m *Margaid) Axis(series *Series, axis Axis, ticker Ticker, grid bool, title string) {
	var xOffset float64 = m.inset
	var yOffset float64 = m.inset
	var axisLength float64
	var crossLength float64
	var xMult float64 = 0
	var yMult float64 = 0
	var tickSign float64 = 1
	var vAlignment svg.VAlignment
	var hAlignment svg.HAlignment
	var axisLabelRotation = 0.0
	var axisLabelAlignment svg.VAlignment
	var axisLabelSign float64 = 1

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
		axisLabelRotation = -90.0
	}

	switch axis {
	case X1Axis:
		xAttributes()
		yOffset = m.height - m.inset
		tickSign = -1
		vAlignment = svg.VAlignTop
		axisLabelAlignment = svg.VAlignBottom
		axisLabelSign = -1
	case X2Axis:
		xAttributes()
		vAlignment = svg.VAlignBottom
		axisLabelAlignment = svg.VAlignTop
	case Y1Axis:
		yAttributes()
		tickSign = -1
		hAlignment = svg.HAlignEnd
		axisLabelAlignment = svg.VAlignTop
	case Y2Axis:
		yAttributes()
		xOffset = m.width - m.inset
		hAlignment = svg.HAlignStart
		axisLabelAlignment = svg.VAlignBottom
		axisLabelSign = -1
	}

	steps := axisLength / tickDistance
	start := ticker.start(axis, series, int(steps))

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, -1),
	).
		StrokeWidth("2px").
		Stroke("black")

	var tick float64
	var hasMore = true

	for tick = start; tick <= max && hasMore; tick, hasMore = ticker.next(tick) {
		value, err := m.project(tick, axis)

		if err == nil {
			m.g.Polyline([]struct{ X, Y float64 }{
				{value * xMult, value * yMult},
				{value*xMult + tickSign*tickSize*(yMult), value*yMult + tickSign*(xMult)*tickSize},
			}...)
		}
	}

	m.g.Transform(
		svg.Translation(xOffset, yOffset),
		svg.Scaling(1, 1),
	).
		Font(m.labelFamily, fmt.Sprintf("%dpx", m.labelSize)).
		FontStyle(svg.StyleNormal, svg.WeightLighter).
		Alignment(hAlignment, vAlignment).
		Fill("black")

	lastLabel := -m.inset
	textOffset := float64(tickSize + textSpacing)
	hasMore = true

	for tick = start; tick <= max && hasMore; tick, hasMore = ticker.next(tick) {
		value, err := m.project(tick, axis)

		if err == nil {
			if value-lastLabel > float64(m.labelSize) {
				m.g.Text(
					value*xMult+(tickSign)*textOffset*(yMult),
					-value*yMult+(-tickSign)*textOffset*(xMult),
					ticker.label(tick))
				lastLabel = value
			}
		}
	}

	if title != "" {
		m.g.Transform(
			svg.Translation(xOffset, yOffset),
			svg.Scaling(1, 1),
			svg.Rotation(axisLabelRotation, 0, 0),
		).
			Font(m.labelFamily, fmt.Sprintf("%dpx", m.labelSize)).
			FontStyle(svg.StyleNormal, svg.WeightBold).
			Alignment(svg.HAlignMiddle, axisLabelAlignment).
			Fill("black")

		x := axisLength / 2
		y := float64(tickSize * axisLabelSign)
		m.g.Text(
			x, y, title,
		)
	}

	if grid {
		m.g.Transform(
			svg.Translation(xOffset, yOffset),
			svg.Scaling(1, -1),
		).
			StrokeWidth("0.5px").Stroke("gray")

		hasMore = true

		for tick = start; tick <= max && hasMore; tick, hasMore = ticker.next(tick) {
			value, err := m.project(tick, axis)

			if err == nil {
				m.g.Polyline([]struct{ X, Y float64 }{
					{value * xMult, value * yMult},
					{value*xMult - tickSign*crossLength*(yMult), value*yMult - tickSign*(xMult)*crossLength},
				}...)
			}
		}
	}
}
