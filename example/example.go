//go:build example
// +build example

package main

import (
	"math/rand"
	"os"
	"time"

	m "github.com/erkkah/margaid"
)

func main() {

	randomSeries := m.NewSeries()
	rand.Seed(time.Now().Unix())
	for i := float64(0); i < 10; i++ {
		randomSeries.Add(m.MakeValue(i+1, 200*rand.Float64()))
	}

	testSeries := m.NewSeries()
	multiplier := 2.1
	v := 0.33
	for i := float64(0); i < 10; i++ {
		v *= multiplier
		testSeries.Add(m.MakeValue(i+1, v))
	}

	diagram := m.New(800, 600,
		m.WithAutorange(m.XAxis, testSeries),
		m.WithAutorange(m.YAxis, testSeries),
		m.WithAutorange(m.Y2Axis, testSeries),
		m.WithProjection(m.YAxis, m.Log),
		m.WithInset(70),
		m.WithPadding(2),
		m.WithColorScheme(90),
		m.WithBackgroundColor("white"),
	)

	diagram.Line(testSeries, m.UsingAxes(m.XAxis, m.YAxis), m.UsingMarker("square"), m.UsingStrokeWidth(1))
	diagram.Smooth(testSeries, m.UsingAxes(m.XAxis, m.Y2Axis), m.UsingStrokeWidth(3.14))
	diagram.Smooth(randomSeries, m.UsingAxes(m.XAxis, m.YAxis), m.UsingMarker("filled-circle"))
	diagram.Axis(testSeries, m.XAxis, diagram.ValueTicker('f', 0, 10), false, "X")
	diagram.Axis(testSeries, m.YAxis, diagram.ValueTicker('f', 1, 2), true, "Y")

	diagram.Frame()
	diagram.Title("A diagram of sorts 📊 📈")

	diagram.Render(os.Stdout)
}
