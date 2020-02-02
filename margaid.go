package margaid

import (
	"fmt"
	"io"
	"math"

	"github.com/erkkah/margaid/brackets"
	"github.com/erkkah/margaid/svg"
)

// Margaid == diagraM
type Margaid struct {
	g *svg.SVG

	width   float64
	height  float64
	inset   float64
	padding float64 // padding [0..1]

	projections map[Axis]Projection
	ranges      map[Axis]minmax

	plots       []string
	colorScheme int

	titleFamily string
	titleSize   int
	labelFamily string
	labelSize   int
}

const (
	defaultPadding = 0
	defaultInset   = 64
	tickDistance   = 55
	tickSize       = 6
	textSpacing    = 4
)

// minmax is the range [min, max] of a chart axis
type minmax struct{ min, max float64 }

// Option is the base type for all series options
type Option func(*Margaid)

// New - Margaid constructor
func New(width, height int, options ...Option) *Margaid {
	defaultRange := minmax{0, 100}

	self := &Margaid{
		g: svg.New(width, height),

		inset:   defaultInset,
		width:   float64(width),
		height:  float64(height),
		padding: defaultPadding,

		projections: map[Axis]Projection{
			X1Axis: Lin,
			X2Axis: Lin,
			Y1Axis: Lin,
			Y2Axis: Lin,
		},

		ranges: map[Axis]minmax{
			X1Axis: defaultRange,
			Y1Axis: defaultRange,
			X2Axis: defaultRange,
			Y2Axis: defaultRange,
		},

		colorScheme: 198,
		titleFamily: "sans-serif",
		titleSize:   18,
		labelFamily: "sans-serif",
		labelSize:   12,
	}

	for _, o := range options {
		o(self)
	}

	return self
}

/// Options

// Projection is the type for the projection constants
type Projection int

// Projection constants
const (
	Lin Projection = iota + 'p'
	Log
)

// WithProjection sets the projection for a given axis
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

// WithAutorange sets range for an axis from the values of a series
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

// WithInset sets the distance between the chart boundaries and the
// charting area.
func WithInset(inset float64) Option {
	return func(m *Margaid) {
		m.inset = inset
	}
}

// WithPadding sets the padding inside the plotting area as a factor
// [0..1] of the area width and height
func WithPadding(padding float64) Option {
	return func(m *Margaid) {
		m.padding = math.Max(0, math.Min(1, padding))
	}
}

// WithColorScheme sets the start color for selecting plot colors.
// The start color is selected as a hue value between 0 and 359.
func WithColorScheme(scheme int) Option {
	return func(m *Margaid) {
		m.colorScheme = scheme % 360
	}
}

// WithTitleFont sets title font family and size in pixels
func WithTitleFont(family string, size int) Option {
	return func(m *Margaid) {
		m.titleFamily = family
		m.titleSize = size
	}
}

// WithLabelFont sets label font family and size in pixels
func WithLabelFont(family string, size int) Option {
	return func(m *Margaid) {
		m.labelFamily = family
		m.labelSize = size
	}
}

/// Drawing

// Title draws a title top center
func (m *Margaid) Title(title string) {
	encoded := svg.EncodeText(title, svg.HAlignMiddle)
	m.g.
		Font(m.titleFamily, fmt.Sprintf("%dpx", m.titleSize)).
		FontStyle(svg.StyleNormal, svg.WeightBold).
		Alignment(svg.HAlignMiddle, svg.VAlignCentral).
		Transform().
		Fill("black").
		Text(m.width/2, m.inset/2, encoded)
}

// LegendPosition decides where to draw the legend
type LegendPosition int

// LegendPosition constants
const (
	RightTop LegendPosition = iota + 'l'
	RightBottom
	BottomLeft
)

// Legend draws a legend for named plots
func (m *Margaid) Legend(position LegendPosition) {
	type namedPlot struct {
		name  string
		color string
	}

	var plots []namedPlot

	for i, label := range m.plots {
		if label != "" {
			color := m.getPlotColor(i)
			plots = append(plots, namedPlot{
				name:  label,
				color: color,
			})
		}
	}

	boxSize := float64(m.labelSize)
	lineHeight := float64(m.labelSize) * 1.5

	listStartX := 0.0
	listStartY := 0.0

	switch position {
	case RightTop:
		listStartX = m.width - m.inset + boxSize + textSpacing
		listStartY = m.inset + 0.5*boxSize
	case RightBottom:
		listStartX = m.width - m.inset + boxSize + textSpacing
		listStartY = m.height - m.inset - lineHeight*float64(len(plots))
	case BottomLeft:
		listStartX = m.inset + 0.5*boxSize
		listStartY = m.height - m.inset + lineHeight + boxSize + tickSize
	}

	style := func(color string) {
		m.g.
			Font(m.labelFamily, fmt.Sprintf("%dpx", m.labelSize)).
			FontStyle(svg.StyleNormal, svg.WeightNormal).
			Alignment(svg.HAlignStart, svg.VAlignTop).
			Fill(color)
	}

	for i, plot := range plots {
		floatIndex := float64(i)
		yPos := listStartY + floatIndex*lineHeight
		xPos := listStartX
		style(plot.color)
		m.g.Rect(xPos, yPos, boxSize, boxSize)
		style("black")
		m.g.Text(xPos+boxSize+textSpacing, yPos, brackets.XMLEscape(plot.name))
	}
}

func (m *Margaid) error(message string) {
	m.g.
		Font(m.titleFamily, fmt.Sprintf("%dpx", m.titleSize)).
		FontStyle(svg.StyleItalic, svg.WeightBold).
		Alignment(svg.HAlignStart, svg.VAlignCentral).
		Transform().
		StrokeWidth("0").Fill("red").
		Text(5, m.inset/2, brackets.XMLEscape(message))
}

// Frame draws a frame around the chart area
func (m *Margaid) Frame() {
	m.g.Transform()
	m.g.Fill("none").Stroke("black").StrokeWidth("2px")
	m.g.Rect(m.inset, m.inset, m.width-m.inset*2, m.height-m.inset*2)
}

// Render renders the graph to the given destination.
func (m *Margaid) Render(writer io.Writer) {
	rendered := m.g.Render()
	writer.Write([]byte(rendered))
}

// Projects a value onto an axis using the current projection
// setting.
// The value returned is in user coordinates, [0..1] * width for the x-axis.
func (m *Margaid) project(value float64, axis Axis) (float64, error) {
	min := m.ranges[axis].min
	max := m.ranges[axis].max

	projected := value
	projection := m.projections[axis]

	var axisLength float64
	switch {
	case axis == X1Axis || axis == X2Axis:
		axisLength = m.width - 2*m.inset
	case axis == Y1Axis || axis == Y2Axis:
		axisLength = m.height - 2*m.inset
	}

	axisPadding := m.padding * axisLength

	if projection == Log {
		if value <= 0 {
			return 0, fmt.Errorf("Cannot draw values <= 0 on log scale")
		}

		if min <= 0 || max <= 0 {
			return 0, fmt.Errorf("Cannot have axis range <= 0 on log scale")
		}

		projected = math.Log10(value)

		min = math.Log10(min)
		max = math.Log10(max)
	}

	projected = axisPadding + (axisLength-2*axisPadding)*(projected-min)/(max-min)
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

// addPlot adds a named plot and returns its ID
func (m *Margaid) addPlot(name string) int {
	id := len(m.plots)
	m.plots = append(m.plots, name)
	return id
}

// getPlotColor picks hues and saturations around the color wheel at prime indices.
// Kind of works for a quick selection of plotting colors.
func (m *Margaid) getPlotColor(id int) string {
	color := 211*id + m.colorScheme
	hue := color % 360
	saturation := 47 + (id*41)%53
	return fmt.Sprintf("hsl(%d, %d%%, 65%%)", hue, saturation)
}
