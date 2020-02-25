// +build bars

package main

import (
	"os"
	"strconv"

	"github.com/erkkah/margaid"
)

func main() {
	// Create a series object and add some values
	seriesA := margaid.NewSeries(margaid.Titled("Team A"))
	seriesA.Add(
		margaid.MakeValue(10, 3.14),
		margaid.MakeValue(34, 12),
		margaid.MakeValue(90, 93.8),
	)

	seriesB := margaid.NewSeries(margaid.Titled("Team B"))
	seriesB.Add(
		margaid.MakeValue(10, 0.62),
		margaid.MakeValue(34, 43),
		margaid.MakeValue(90, 88.1),
	)

	labeler := func(value float64) string {
		return strconv.FormatFloat(value, 'f', 1, 64)
	}

	// Create the diagram object,
	// add some padding for the bars and extra inset for the legend
	diagram := margaid.New(800, 600, margaid.WithPadding(10), margaid.WithInset(80))

	// Plot the series
	diagram.Bar([]*margaid.Series{seriesA, seriesB})

	// Add a legend
	diagram.Legend(margaid.RightBottom)

	// Add a frame and X axis
	diagram.Frame()
	diagram.Axis(seriesA, margaid.XAxis, diagram.LabeledTicker(labeler), false, "Lemmings")
	diagram.Axis(seriesA, margaid.Y2Axis, diagram.LabeledTicker(labeler), true, "")
	diagram.Axis(seriesB, margaid.YAxis, diagram.LabeledTicker(labeler), true, "")

	// Render to stdout
	diagram.Render(os.Stdout)
}
