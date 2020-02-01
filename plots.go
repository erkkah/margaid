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

type plotOptions struct {
	xAxis  Axis
	yAxis  Axis
	marker string
}

// Using is the base type for plotting options
type Using func(*plotOptions)

func getPlotOptions(using []Using) plotOptions {
	options := plotOptions{
		xAxis: XAxis,
		yAxis: YAxis,
	}

	for _, u := range using {
		u(&options)
	}

	return options
}

// UsingAxes selects the x and y axis for plotting
func UsingAxes(x, y Axis) Using {
	return func(o *plotOptions) {
		o.xAxis = x
		o.yAxis = y
	}
}

// UsingMarker selects a marker for highlighting plotted values.
// See svg.Marker for valid marker types.
func UsingMarker(marker string) Using {
	return func(o *plotOptions) {
		o.marker = marker
	}
}

// Line draws a series using straight lines
func (m *Margaid) Line(series *Series, using ...Using) {
	options := getPlotOptions(using)

	points, err := m.getProjectedValues(series, options.xAxis, options.yAxis)
	if err != nil {
		m.error(err.Error())
		return
	}

	id := m.addPlot(series.title)
	color := m.getPlotColor(id)
	m.g.
		StrokeWidth("3px").
		Fill("none").
		Stroke(color).
		Marker(options.marker).
		Transform(
			svg.Translation(m.inset, m.height-m.inset),
			svg.Scaling(1, -1),
		).
		Polyline(points...).
		Marker("").
		Transform()
}

// Smooth draws one series as a smooth curve
func (m *Margaid) Smooth(series *Series, using ...Using) {
	options := getPlotOptions(using)

	points, err := m.getProjectedValues(series, options.xAxis, options.yAxis)
	if err != nil {
		m.error(err.Error())
		return
	}

	id := m.addPlot(series.title)
	color := m.getPlotColor(id)
	m.g.
		StrokeWidth("4").
		Fill("none").
		Stroke(color).
		Marker(options.marker).
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
	m.g.Path(path.String()).Marker("").Transform()
}

// Bar draws bars for the specified group of series.
func (m *Margaid) Bar(series []*Series, using ...Using) {
	if len(series) == 0 {
		return
	}

	options := getPlotOptions(using)

	maxSize := 0
	for _, s := range series {
		if s.Size() > maxSize {
			maxSize = s.Size()
		}
	}

	barWidth := (m.width - 2*m.inset) / float64(maxSize)
	barWidth /= 1.5
	barWidth /= float64(len(series))
	barOffset := -(barWidth / 2) * float64(len(series)-1)

	for i, s := range series {
		points, err := m.getProjectedValues(s, options.xAxis, options.yAxis)

		if err != nil {
			m.error(err.Error())
			return
		}
		id := m.addPlot(s.title)
		color := m.getPlotColor(id)
		m.g.
			StrokeWidth("1px").
			Fill(color).
			Stroke(color).
			Transform(
				svg.Translation(m.inset, m.height-m.inset),
				svg.Scaling(1, -1),
			)

		for _, p := range points {
			m.g.Rect(barOffset+float64(i)*barWidth+p.X-barWidth/2, 0, barWidth, p.Y)
		}
	}

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
