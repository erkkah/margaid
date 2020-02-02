package main

import (
	"math/rand"
	"os"
	"time"

	m "github.com/erkkah/margaid"
)

func main() {

	randomSeries := m.NewSeries(m.Titled("Random"))
	rand.Seed(time.Now().Unix())
	for i := float64(0); i < 10; i++ {
		randomSeries.Add(m.MakeValue(i+1, 200*rand.Float64()))
	}

	testSeries := m.NewSeries(m.Titled("Exponential"))
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
		m.WithProjection(m.XAxis, m.Lin),
		m.WithProjection(m.YAxis, m.Log),
		m.WithProjection(m.Y2Axis, m.Lin),
		m.WithInset(70),
		m.WithColorScheme(90),
	)

	diagram.Line(testSeries, m.UsingAxes(m.XAxis, m.YAxis), m.UsingMarker("square"))
	diagram.Smooth(testSeries, m.UsingAxes(m.XAxis, m.Y2Axis))
	diagram.Smooth(randomSeries, m.UsingAxes(m.XAxis, m.YAxis), m.UsingMarker("filled-circle"))
	diagram.Axis(testSeries, m.XAxis, diagram.ValueTicker('f', 0, 10), false)
	diagram.Axis(testSeries, m.YAxis, diagram.ValueTicker('f', 1, 2), true)

	diagram.Frame()
	diagram.Title("A diagram of sorts 📊 📈")

	diagram.Render(os.Stdout)
}
