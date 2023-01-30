//go:build minimal
// +build minimal

package main

import (
	"os"

	"github.com/erkkah/margaid"
)

func main() {
	// Create a series object and add some values
	series := margaid.NewSeries()
	series.Add(margaid.MakeValue(10, 3.14), margaid.MakeValue(90, 93.8))

	// Create the diagram object:
	diagram := margaid.New(800, 600, margaid.WithBackgroundColor("white"))

	// Plot the series
	diagram.Line(series)

	// Add a frame and X axis
	diagram.Frame()
	diagram.Axis(series, margaid.XAxis, diagram.ValueTicker('f', 2, 10), false, "Values")

	// Render to stdout
	diagram.Render(os.Stdout)
}
