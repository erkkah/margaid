package svg

import (
	"fmt"
	"strconv"
	"strings"

	br "github.com/erkkah/margaid/brackets"
)

// SVG builds SVG format images using a small subset of the standard
type SVG struct {
	brackets *br.Brackets

	transforms  []Transform
	attributes  br.Attributes
	styleInSync bool

	parent *SVG
}

// Transform represents a transform function
type Transform struct {
	function  string
	arguments []float64
}

// New - SVG constructor
func New(width int, height int) *SVG {
	self := makeSVG()
	self.brackets.Open("svg", br.Attributes{
		"width":               strconv.Itoa(width),
		"height":              strconv.Itoa(height),
		"viewbox":             fmt.Sprintf("0 0 %d %d", width, height),
		"preserveAspectRatio": "xMidYMid meet",
		"xmlns":               "http://www.w3.org/2000/svg",
	})
	self.addMarkers()
	return &self
}

func (svg *SVG) addMarkers() {
	svg.brackets.Open("defs").
		Open("marker", br.Attributes{
			"id":           "circle",
			"viewBox":      "0 0 10 10 ",
			"refX":         "5",
			"refY":         "5",
			"markerUnits":  "userSpaceOnUse",
			"markerWidth":  "2%",
			"markerHeight": "2%",
		}).
		Add("circle", br.Attributes{
			"cx":     "5",
			"cy":     "5",
			"r":      "3",
			"fill":   "none",
			"stroke": "black",
		}).
		Close().
		Open("marker", br.Attributes{
			"id":           "filled-circle",
			"viewBox":      "0 0 10 10 ",
			"refX":         "5",
			"refY":         "5",
			"markerUnits":  "userSpaceOnUse",
			"markerWidth":  "2%",
			"markerHeight": "2%",
		}).
		Add("circle", br.Attributes{
			"cx":     "5",
			"cy":     "5",
			"r":      "3",
			"fill":   "black",
			"stroke": "none",
		}).
		Close().
		Open("marker", br.Attributes{
			"id":           "square",
			"viewBox":      "0 0 10 10 ",
			"refX":         "5",
			"refY":         "5",
			"markerUnits":  "userSpaceOnUse",
			"markerWidth":  "2%",
			"markerHeight": "2%",
		}).
		Add("rect", br.Attributes{
			"x":      "2",
			"y":      "2",
			"width":  "6",
			"height": "6",
			"fill":   "none",
			"stroke": "black",
		}).
		Close().
		Open("marker", br.Attributes{
			"id":           "filled-square",
			"viewBox":      "0 0 10 10 ",
			"refX":         "5",
			"refY":         "5",
			"markerUnits":  "userSpaceOnUse",
			"markerWidth":  "2%",
			"markerHeight": "2%",
		}).
		Add("rect", br.Attributes{
			"x":      "2",
			"y":      "2",
			"width":  "6",
			"height": "6",
			"fill":   "black",
			"stroke": "none",
		}).
		Close().
		Close()
}

func makeSVG() SVG {
	return SVG{
		brackets: br.New(),
		attributes: br.Attributes{
			"fill":            "green",
			"stroke":          "black",
			"stroke-width":    "1px",
			"stroke-linecap":  "round",
			"stroke-linejoin": "round",
		},
	}
}

// Child adds a sub-SVG at x, y
func (svg *SVG) Child(x, y float64) *SVG {
	self := makeSVG()
	self.parent = svg
	self.brackets.Open("svg", br.Attributes{
		"x": ftos(x),
		"y": ftos(y),
	})
	return &self
}

// Close closes a child SVG and returns the parent.
// If the current SVG is not a child, this is a noop.
func (svg *SVG) Close() *SVG {
	if svg.parent != nil {
		svg.brackets.CloseAll()
		svg.parent.brackets.Append(svg.brackets)
		return svg.parent
	}
	return svg
}

// Render generates SVG code for the current image, and clears the canvas
func (svg *SVG) Render() string {
	svg.brackets.CloseAll()
	result := svg.brackets.String()
	svg.brackets = br.New()
	return result
}

func attributeDiff(old, new br.Attributes) (diff br.Attributes, extendable bool) {
	diff = br.Attributes{}
	extendable = true

	for k, newValue := range new {
		if k == "transform" {
			extendable = false
			return
		}
		if oldValue, found := old[k]; found {
			if oldValue != newValue {
				diff[k] = newValue
			}
		} else {
			diff[k] = newValue
		}
	}

	for k := range old {
		if _, found := new[k]; !found {
			extendable = false
			return
		}
	}
	return
}

func shouldExtendParentStyle(old, new, diff br.Attributes) bool {
	return diff.Size() < new.Size()
}

func (svg *SVG) updateStyle() {
	if !svg.styleInSync {
		current := svg.brackets.Current()
		nextAttributes := svg.attributes
		if current != nil && current.Name() == "g" {
			diff, extendable := attributeDiff(current.Attributes(), svg.attributes)
			if extendable && shouldExtendParentStyle(current.Attributes(), svg.attributes, diff) {
				nextAttributes = diff
			} else {
				svg.brackets.Close()
			}
		}
		svg.brackets.Open("g", nextAttributes)
		svg.styleInSync = true
	}
}

func (svg *SVG) setAttribute(attr string, value string) {
	if svg.attributes[attr] != value {
		if value == "" {
			delete(svg.attributes, attr)
		} else {
			svg.attributes[attr] = value
		}
		svg.styleInSync = false
	}
}

/// Drawing

// Path adds a SVG style path
func (svg *SVG) Path(path string) *SVG {
	svg.updateStyle()
	top := svg.brackets.Last()
	if top.Name() == "path" && strings.TrimSpace(path)[0] == 'M' {
		commands := top.Attributes()["d"]
		commands += path
		top.SetAttribute("d", commands)
	} else {
		svg.brackets.Add("path", br.Attributes{
			"d":             path,
			"vector-effect": "non-scaling-stroke",
		})
	}
	return svg
}

// Polyline adds a polyline from a list of points
func (svg *SVG) Polyline(points ...struct{ X, Y float64 }) *SVG {
	if len(points) < 2 {
		return svg
	}

	var path strings.Builder
	first := points[0]
	path.WriteString(fmt.Sprintf("M%s,%s ", ftos(first.X), ftos(first.Y)))

	for _, p := range points[1:] {
		path.WriteString(fmt.Sprintf("L%s,%s ", ftos(p.X), ftos(p.Y)))
	}
	return svg.Path(path.String())
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
	attributes := br.Attributes{
		"x":             ftos(x),
		"y":             ftos(y),
		"stroke":        "none",
		"vector-effect": "non-scaling-stroke",
	}
	if alignment, ok := svg.attributes["dominant-baseline"]; ok {
		attributes["dominant-baseline"] = alignment
	}
	svg.brackets.Open("text", attributes)
	svg.brackets.Text(txt)
	svg.brackets.Close()
	return svg
}

/// Transformations

// Rotation rotates by angle degrees clockwise around (x, y)
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
			builder.WriteString(ftos(a))
			builder.WriteRune(' ')
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

// Marker adds start, mid and end markers to all following strokes.
// The specified marker has to be one of "circle" and "square".
// Setting the marker to the empty string clears the marker.
func (svg *SVG) Marker(marker string) *SVG {
	reference := ""
	if marker != "" {
		reference = fmt.Sprintf("url(#%s)", marker)
	}
	svg.setAttribute("marker-start", reference)
	svg.setAttribute("marker-mid", reference)
	svg.setAttribute("marker-end", reference)
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
	VAlignTop     VAlignment = "hanging"
	VAlignCentral            = "middle"
	VAlignBottom             = "baseline"
)

// Alignment sets current text alignment
func (svg *SVG) Alignment(horizontal HAlignment, vertical VAlignment) *SVG {
	svg.setAttribute("text-anchor", string(horizontal))
	svg.setAttribute("dominant-baseline", string(vertical))
	return svg
}

/// Utilities

func ftos(value float64) string {
	if float64(int(value)) == value {
		return strconv.Itoa(int(value))
	}
	return fmt.Sprintf("%e", value)
}

// EncodeText applies proper xml escaping and svg line breaking
// at each newline in the raw text and returns a section ready
// for inclusion in a <text> element.
// NOTE: Line breaking is kind of hacky, since we cannot actually
// measure text in SVG, and assume that all characters are 1ex wide.
func EncodeText(raw string, alignment HAlignment) string {
	chunks := strings.Split(raw, "\n")
	if len(chunks) > 1 {
		var lines strings.Builder

		lines.WriteString(br.XMLEscape(chunks[0]))
		previousLength := float64(len(chunks[0]))

		for _, chunk := range chunks[1:] {
			chunk = br.XMLEscape(chunk)
			currentLength := float64(len(chunk))

			carriageReturn := 0.0
			switch alignment {
			case HAlignStart:
				carriageReturn = previousLength
			case HAlignMiddle:
				carriageReturn = (previousLength + currentLength) / 2
			case HAlignEnd:
				carriageReturn = currentLength
			}

			lines.WriteString(fmt.Sprintf(`<tspan dx="-%fex" dy="1em">%s</tspan>`, carriageReturn, chunk))
			previousLength = float64(len(chunk))
		}
		return lines.String()
	}
	return br.XMLEscape(raw)
}
