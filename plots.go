package margaid

import (
	"fmt"
	"math"
	"strings"

	"github.com/erkkah/margaid/svg"
)

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
	color := m.getPlotColor(id)
	m.g.StrokeWidth("3px").Fill("none").Stroke(color)

	m.g.Transform(
		svg.Translation(m.inset, m.height-m.inset),
		svg.Scaling(1, -1),
	)
	m.g.Polyline(points...)
	m.g.Transform()
}

// Smooth draws one series as a smooth curve
func (m *Margaid) Smooth(series *Series, axes ...AxisSelection) {
	xAxis, yAxis := getSelectedAxes(axes)
	points, err := m.getProjectedValues(series, xAxis, yAxis)
	if err != nil {
		m.error(err.Error())
		return
	}

	id := m.addPlot(series.title)
	color := m.getPlotColor(id)
	m.g.StrokeWidth("3px").Fill("none").Stroke(color).
		Transform(
			svg.Translation(m.inset, m.height-m.inset),
			svg.Scaling(1, -1),
		)

	var path strings.Builder

	path.WriteString(fmt.Sprintf("M%e,%e ", points[0].X, points[0].Y))
	catmull := catmullRom2bezier(points)

	for _, p := range catmull {
		path.WriteString(fmt.Sprintf("C%e,%e %e,%e %e,%e ",
			p[0].X, p[0].Y,
			p[1].X, p[1].Y,
			p[2].X, p[2].Y,
		))
	}
	m.g.Path(path.String())
}

// Bar draws bars for the specified series
func (m *Margaid) Bar(series *Series, axes ...AxisSelection) {
	xAxis, yAxis := getSelectedAxes(axes)
	points, err := m.getProjectedValues(series, xAxis, yAxis)
	if err != nil {
		m.error(err.Error())
		return
	}

	id := m.addPlot(series.title)
	color := m.getPlotColor(id)
	m.g.StrokeWidth("1px").Fill(color).Stroke(color).
		Transform(
			svg.Translation(m.inset, m.height-m.inset),
			svg.Scaling(1, -1),
		)

	barWidth := (m.width - 2*m.inset) / float64(len(points))
	barWidth /= 2

	for _, p := range points {
		m.g.Rect(p.X-barWidth/2, 0, barWidth, p.Y)
	}
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

// BezierPoint is one Bezier curve control point
type BezierPoint [3]struct{ X, Y float64 }

// Pulled from:
// https://advancedweb.hu/plotting-charts-with-svg/

func catmullRom2bezier(points []struct{ X, Y float64 }) []BezierPoint {
	var result = []BezierPoint{}

	for i := range points[0 : len(points)-1] {
		p := []struct{ X, Y float64 }{}

		idx := int(math.Max(float64(i-1), 0))
		p = append(p, struct{ X, Y float64 }{
			X: points[idx].X,
			Y: points[idx].Y,
		})

		p = append(p, struct{ X, Y float64 }{
			X: points[i].X,
			Y: points[i].Y,
		})

		p = append(p, struct{ X, Y float64 }{
			X: points[i+1].X,
			Y: points[i+1].Y,
		})

		idx = int(math.Min(float64(i+2), float64(len(points)-1)))
		p = append(p, struct{ X, Y float64 }{
			X: points[idx].X,
			Y: points[idx].Y,
		})

		// Catmull-Rom to Cubic Bezier conversion matrix
		//    0       1       0       0
		//  -1/6      1      1/6      0
		//    0      1/6      1     -1/6
		//    0       0       1       0

		bp := BezierPoint{}

		bp[0] = struct{ X, Y float64 }{
			X: ((-p[0].X + 6*p[1].X + p[2].X) / 6),
			Y: ((-p[0].Y + 6*p[1].Y + p[2].Y) / 6),
		}

		bp[1] = struct{ X, Y float64 }{
			X: ((p[1].X + 6*p[2].X - p[3].X) / 6),
			Y: ((p[1].Y + 6*p[2].Y - p[3].Y) / 6),
		}

		bp[2] = struct{ X, Y float64 }{
			X: p[2].X,
			Y: p[2].Y,
		}

		result = append(result, bp)
	}

	return result
}
