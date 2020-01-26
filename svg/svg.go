package svg

import (
	"fmt"
	"strconv"
	"strings"

	br "github.com/erkkah/margaid/brackets"
)

// SVG builds SVG format images using a small subset of the standard
type SVG struct {
	brackets br.Brackets

	transforms  []Transform
	attributes  br.Attributes
	styleInSync bool
}

// Transform represents a transform function
type Transform struct {
	function  string
	arguments []float64
}

// New - SVG constructor
func New(width int, height int) *SVG {
	self := &SVG{
		brackets: br.New(),
		attributes: br.Attributes{
			"fill":            "green",
			"stroke":          "black",
			"stroke-width":    "1px",
			"stroke-linecap":  "round",
			"stroke-linejoin": "round",
		},
	}
	self.brackets.Open("svg", br.Attributes{
		"width":               strconv.Itoa(width),
		"height":              strconv.Itoa(height),
		"viewbox":             fmt.Sprintf("0 0 %d %d", width, height),
		"preserveAspectRatio": "xMidYMid meet",
		"xmlns":               "http://www.w3.org/2000/svg",
	})
	return self
}

// Render generates SVG code for the current image, and clears the canvas
func (svg *SVG) Render() string {
	svg.brackets.CloseAll()
	result := svg.brackets.String()
	svg.brackets = br.New()
	return result
}

func (svg *SVG) updateStyle() {
	if !svg.styleInSync {
		if svg.brackets.Current() == "g" {
			svg.brackets.Close()
		}
		svg.brackets.Open("g", svg.attributes)
		svg.styleInSync = true
	}
}

func (svg *SVG) setAttribute(attr string, value string) {
	if svg.attributes[attr] != value {
		svg.attributes[attr] = value
		svg.styleInSync = false
	}
}

/// Drawing

// Path adds a SVG style path
func (svg *SVG) Path(path string) *SVG {
	svg.updateStyle()
	svg.brackets.Add("path", br.Attributes{
		"d":             path,
		"vector-effect": "non-scaling-stroke",
	})
	return svg
}

// Polyline adds a polyline from a list of points
func (svg *SVG) Polyline(points ...struct{ X, Y float64 }) *SVG {
	var builder strings.Builder

	for _, point := range points {
		builder.WriteString(fmt.Sprintf("%.3f,%.3f ", point.X, point.Y))
	}

	svg.updateStyle()
	svg.brackets.Add("polyline", br.Attributes{
		"points":        builder.String(),
		"vector-effect": "non-scaling-stroke",
	})
	return svg
}

// Rect adds a rect defined by x, y, width and height
func (svg *SVG) Rect(x, y, width, height float64) *SVG {
	svg.updateStyle()
	svg.brackets.Add("rect", br.Attributes{
		"x":             ftos(x),
		"y":             ftos(y),
		"width":         ftos(width),
		"height":        ftos(height),
		"vector-effect": "non-scaling-stroke",
	})
	return svg
}

// Text draws text at x, y
func (svg *SVG) Text(x, y float64, txt string) *SVG {
	svg.updateStyle()
	svg.brackets.Open("text", br.Attributes{
		"x":             ftos(x),
		"y":             ftos(y),
		"vector-effect": "non-scaling-stroke",
	})
	svg.brackets.Text(txt)
	svg.brackets.Close()
	return svg
}

/// Transformations

// Rotation rotates by angle degrees counter-clockwise
// around (x, y)
func Rotation(angle, x, y float64) Transform {
	return Transform{
		"rotate",
		[]float64{angle, x, y},
	}
}

// Translation moves by (x, y)
func Translation(x, y float64) Transform {
	return Transform{
		"translate",
		[]float64{x, y},
	}
}

// Scaling scales by (xScale, yScale)
func Scaling(xScale, yScale float64) Transform {
	return Transform{
		"scale",
		[]float64{xScale, yScale},
	}
}

// Transform sets the current list of transforms
// that will be used by the next set of drawing operations.
// Specifying no transforms resets the transformation matrix
// to identity.
func (svg *SVG) Transform(transforms ...Transform) *SVG {
	var builder strings.Builder

	for _, t := range transforms {
		builder.WriteString(t.function)
		builder.WriteRune('(')
		for _, a := range t.arguments {
			builder.WriteString(fmt.Sprintf("%.3f ", a))
		}
		builder.WriteRune(')')
	}

	svg.setAttribute("transform", builder.String())
	return svg
}

/// Style

// Fill sets current fill style
func (svg *SVG) Fill(fill string) *SVG {
	svg.setAttribute("fill", fill)
	return svg
}

// Stroke sets current stroke
func (svg *SVG) Stroke(stroke string) *SVG {
	svg.setAttribute("stroke", stroke)
	return svg
}

// Color sets current stroke and fill
func (svg *SVG) Color(color string) *SVG {
	svg.Stroke(color)
	svg.Fill(color)
	return svg
}

// StrokeWidth sets current stroke width
func (svg *SVG) StrokeWidth(width string) *SVG {
	svg.setAttribute("stroke-width", width)
	return svg
}

// Font sets current font family and size
func (svg *SVG) Font(font string, size string) *SVG {
	svg.setAttribute("font-family", font)
	svg.setAttribute("font-size", size)
	return svg
}

// Style is the type for the text style constants
type Style string

// Weight is the type for the text weight constants
type Weight string

// Text style constants
const (
	StyleNormal Style = "normal"
	StyleItalic       = "italic"
)

// Text weight constants
const (
	WeightNormal  Weight = "normal"
	WeightBold           = "bold"
	WeightLighter        = "lighter"
)

// FontStyle sets the current font style and weight
func (svg *SVG) FontStyle(style Style, weight Weight) *SVG {
	svg.setAttribute("font-style", string(style))
	svg.setAttribute("font-weight", string(weight))
	return svg
}

// VAlignment is the type for the vertical alignment constants
type VAlignment string

// HAlignment is the type for the horizontal alignment constants
type HAlignment string

// Horizontal text alignment constants
const (
	HAlignStart  HAlignment = "start"
	HAlignMiddle            = "middle"
	HAlignEnd               = "end"
)

// Vertical text alignment constants
const (
	VAlignTop     VAlignment = "top"
	VAlignCentral            = "central"
	VAlignBottom             = "bottom"
)

// Alignment sets current text alignment
func (svg *SVG) Alignment(horizontal HAlignment, vertical VAlignment) *SVG {
	svg.setAttribute("text-anchor", string(horizontal))
	svg.setAttribute("alignment-baseline", string(vertical))
	return svg
}

/// Utilities

func ftos(value float64) string {
	return fmt.Sprintf("%.3f", value)
}
