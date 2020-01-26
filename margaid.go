package margaid

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/erkkah/margaid/svg"
)

//https://developer.mozilla.org/en-US/docs/Web/SVG/Element/marker

/*
   <marker id="arrow" viewBox="0 0 10 10" refX="5" refY="5"
        markerWidth="6" markerHeight="6"
        orient="auto-start-reverse">
      <path d="M 0 0 L 10 5 L 0 10 z" />
	</marker>
*/

// Margaid == diagraM
type Margaid struct {
	g *svg.SVG

	width  float64
	height float64
	inset  float64

	projections map[Axis]Projection
	ranges      map[Axis]minmax

	plots []string
}

type minmax struct{ min, max float64 }

// Option is the base type for all series options
type Option func(*Margaid)

// New - Margaid constructor
func New(width, height int, options ...Option) *Margaid {
	self := &Margaid{
		g: svg.New(width, height),

		inset:  32,
		width:  float64(width),
		height: float64(height),

		projections: map[Axis]Projection{
			X1Axis: Lin,
			X2Axis: Lin,
			Y1Axis: Lin,
			Y2Axis: Lin,
		},

		ranges: map[Axis]minmax{
			X1Axis: {0, 100},
			Y1Axis: {0, 100},
			X2Axis: {0, 100},
			Y2Axis: {0, 100},
		},
	}

	for _, o := range options {
		o(self)
	}

	return self
}

// Projection is the type for the projection constants
type Projection int

// Projection constants
const (
	Lin Projection = iota + 'p'
	Log
)

// WithProjection sets the projection
func WithProjection(axis Axis, proj Projection) Option {
	return func(m *Margaid) {
		m.projections[axis] = proj
	}
}

// WithRange sets a fixed plotting range for a given axis
func WithRange(axis Axis, min, max float64) Option {
	return func(m *Margaid) {
		m.ranges[axis] = minmax{min, max}
	}
}

// WithAutorange sets range for an axis from the values in a series
func WithAutorange(axis Axis, series *Series) Option {
	return func(m *Margaid) {
		if axis == X1Axis || axis == X2Axis {
			m.ranges[axis] = minmax{
				series.MinX(),
				series.MaxX(),
			}
		}
		if axis == Y1Axis || axis == Y2Axis {
			m.ranges[axis] = minmax{
				series.MinY(),
				series.MaxY(),
			}
		}
	}
}

// Grid draws a grid for an axis based on the current range
func (m *Margaid) Grid(axis ...Axis) {

}

// Title draws a title top center
func (m *Margaid) Title(title string) {
	m.g.
		Font("sans-serif", "12pt").
		FontStyle(svg.StyleNormal, svg.WeightLighter).
		Alignment(svg.HAlignMiddle, svg.VAlignCentral).
		Transform().
		StrokeWidth("0").Fill("black").
		Text(m.width/2, m.inset/2, title)
}

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

// Ticker converts values to strings for drawing tick marks
type Ticker func(Value) string

// TimeTicker returns time valued tick labels in the specified time format.
func TimeTicker(value Value, format string) string {
	return value.GetXAsTime().Format(format)
}

// ValueTicker returns tick labels by converting floats using strconv.FormatFloat
func ValueTicker(value Value, style byte, precision int) string {
	return strconv.FormatFloat(float64(value.X), style, precision, 32)
}

// Axis draws tick marks and labels using the specified ticker
func (m *Margaid) Axis(series string, axis Axis, ticker Ticker) {
}

// Legend draws a legend box for the specified set of series
func (m *Margaid) Legend(series []string) {
}

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

	plotWidth := m.width - 2*m.inset
	plotHeight := m.height - 2*m.inset

	m.g.Transform(
		svg.Translation(m.inset, m.height-m.inset),
		svg.Scaling(plotWidth, -plotHeight),
	)
	m.g.Polyline(points...)
	m.g.Transform()
}

func xmlEscape(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func (m *Margaid) error(message string) {
	m.g.
		Font("sans-serif", "10pt").
		FontStyle(svg.StyleItalic, svg.WeightBold).
		Alignment(svg.HAlignStart, svg.VAlignCentral).
		Transform().
		StrokeWidth("0").Fill("red").
		Text(5, m.inset/2, xmlEscape(message))
}

// Smooth draws one series as a smooth curve
func (m *Margaid) Smooth(series string) {
}

// Bar draws bars for the specified series
func (m *Margaid) Bar(series string) {
}

// Frame draws a frame around the chart area
func (m *Margaid) Frame() {
	m.g.Fill("none").Stroke("black").StrokeWidth("2px")
	m.g.Rect(m.inset, m.inset, m.width-m.inset*2, m.height-m.inset*2)
}

// Render renders the graph to the given destination.
func (m *Margaid) Render(writer io.Writer) {
	rendered := m.g.Render()
	writer.Write([]byte(rendered))
}

func (m *Margaid) project(value float64, axis Axis) (float64, error) {
	min := m.ranges[axis].min
	max := m.ranges[axis].max

	projected := value
	projection := m.projections[axis]

	if projection == Log {
		if min < 0 || value < 0 {
			return 0, fmt.Errorf("Cannot draw values <= 0 on log scale")
		}

		projected = math.Log10(projected)
		min = math.Log10(min)
		max = math.Log10(max)
	}

	projected = (projected - min) / (max - min)
	return projected, nil
}

func (m *Margaid) getProjectedValues(series *Series, xAxis, yAxis Axis) (points []struct{ X, Y float64 }, err error) {
	values := series.Values()
	for values.Next() {
		v := values.Get()
		v.X, err = m.project(v.X, xAxis)
		v.Y, err = m.project(v.Y, yAxis)
		points = append(points, v)
	}
	return
}

func getSelectedAxes(axes []AxisSelection) (xAxis, yAxis Axis) {
	xAxis = XAxis
	yAxis = YAxis
	for _, s := range axes {
		xAxis = s.x
		yAxis = s.y
	}
	return
}

func (m *Margaid) addPlot(name string) int {
	id := len(m.plots)
	m.plots = append(m.plots, name)
	return id
}

func getPlotColor(id int) string {
	color := 127*id + 270
	hue := color % 360
	saturation := 70 + ((color/360)*13)%30
	return fmt.Sprintf("hsl(%d, %d%%, 50%%)", hue, saturation)
}
