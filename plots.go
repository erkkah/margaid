package margaid

import "github.com/erkkah/margaid/svg"

// AxisSelection specified which axes to use
type AxisSelection struct {
	x Axis
	y Axis
}

// UsingAxes selects the x and y axis for plotting
func UsingAxes(x, y Axis) AxisSelection {
	return AxisSelection{
		x: x,
		y: y,
	}
}

// Line draws a series using straight lines
func (m *Margaid) Line(series *Series, axes ...AxisSelection) {
	xAxis, yAxis := getSelectedAxes(axes)
	points, err := m.getProjectedValues(series, xAxis, yAxis)
	if err != nil {
		m.error(err.Error())
		return
	}

	id := m.addPlot(series.title)
	color := getPlotColor(id)
	m.g.StrokeWidth("3px").Fill("none").Stroke(color)

	m.g.Transform(
		svg.Translation(m.inset, m.height-m.inset),
		svg.Scaling(1, -1),
	)
	m.g.Polyline(points...)
	m.g.Transform()
}

// Smooth draws one series as a smooth curve
func (m *Margaid) Smooth(series string) {
}

// Bar draws bars for the specified series
func (m *Margaid) Bar(series string) {
}

func getSelectedAxes(selection []AxisSelection) (xAxis, yAxis Axis) {
	xAxis = XAxis
	yAxis = YAxis
	for _, s := range selection {
		xAxis = s.x
		yAxis = s.y
	}
	return
}
