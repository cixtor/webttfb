package main

import (
	"fmt"
)

// Colorizer returns the value for a specific test (connection time, first byte
// time and total time) with a colored background either green, yellow or red
// for good, strange and bad results. The comparison values are extracted from
// the JavaScript library that powers the functionality of the external API
// service.
//
// {{API}}/assets/loadtime-parser.js::colorize_data_attr()
type Colorizer interface {
	success() float64
	warning() float64
	danger() float64
}

// ColorConn defines the limit for the danger, warning and successful tests for
// the connection time and subsequently defines the color that will be applied
// to the values during the rendering of the table (if applicable). Failed tests
// will not be colorized even if they fall under one of these limits, usually
// the successful one.
type ColorConn struct{}

func (c ColorConn) success() float64 { return 0.18 }
func (c ColorConn) warning() float64 { return 0.55 }
func (c ColorConn) danger() float64  { return 0.70 }

// ColorTTFB defines the limit for the danger, warning and successful tests for
// the time to first byte and subsequently defines the color that will be
// applied to the values during the rendering of the table (if applicable).
// Failed tests will not be colorized even if they fall under one of these
// limits, usually the successful one.
type ColorTTFB struct{}

func (c ColorTTFB) success() float64 { return 0.40 }
func (c ColorTTFB) warning() float64 { return 0.99 }
func (c ColorTTFB) danger() float64  { return 1.28 }

// ColorTTL defines the limit for the danger, warning and successful tests for
// the total time and subsequently defines the color that will be applied to the
// values during the rendering of the table (if applicable). Failed tests will
// not be colorized even if they fall under one of these limits, usually the
// successful one.
type ColorTTL struct{}

func (c ColorTTL) success() float64 { return 0.55 }
func (c ColorTTL) warning() float64 { return 1.15 }
func (c ColorTTL) danger() float64  { return 1.45 }

// Grade evaluates the average HTTP request total time through all the testing
// servers and assigns a grade to the website's responsiveness. If there were
// too many failures during the testing process the program defaults to the
// worst grade.
type Grade struct{}

func (g Grade) perfect() float64   { return 0.510 }
func (g Grade) excellent() float64 { return 0.850 }
func (g Grade) good() float64      { return 1.150 }
func (g Grade) bad() float64       { return 1.550 }
func (g Grade) awful() float64     { return 1.950 }
func (g Grade) worst() float64     { return 2.500 }

// Paint builds the escape sequence to render the colors.
func Paint(c Colorizer, value float64) string {
	if value > c.danger() {
		return fmt.Sprintf("\033[38;5;255;48;5;009m%.3f\033[0m", value)
	}

	if value > c.warning() {
		return fmt.Sprintf("\033[38;5;008;48;5;226m%.3f\033[0m", value)
	}

	if value < c.success() {
		return fmt.Sprintf("\033[38;5;255;48;5;034m%.3f\033[0m", value)
	}

	return fmt.Sprintf("%.3f", value)
}

// Colorize returns the floating point with a background color.
func Colorize(group string, value float64) string {
	if value == 0.0 {
		// Do not colorize failed tests.
		return fmt.Sprintf("%.3f", value)
	}

	if group == connectionTime {
		return Paint(ColorConn{}, value)
	}

	if group == timeToFirstByte {
		return Paint(ColorTTFB{}, value)
	}

	if group == totalTime {
		return Paint(ColorTTL{}, value)
	}

	return fmt.Sprintf("%.3f", value)
}

// PerformanceGrade evaluates the average HTTP request total time through all
// the testing servers and assigns a grade to the website's responsiveness. If
// there were too many failures during the testing process the program defaults
// to the worst grade.
func PerformanceGrade(t *TTFB) string {
	type score struct {
		Grade string
		Color string
		Cond  bool
	}

	var g Grade
	var level score
	avg := t.Average(totalTime)
	scores := []score{
		{Grade: "F", Color: "38;5;000;48;5;007m", Cond: len(t.Messages) > 4 || avg <= 0},
		{Grade: "A+", Color: "38;5;255;48;5;038m", Cond: avg <= g.perfect()},
		{Grade: "A", Color: "38;5;255;48;5;034m", Cond: avg <= g.excellent()},
		{Grade: "B", Color: "38;5;008;48;5;226m", Cond: avg <= g.good()},
		{Grade: "C", Color: "38;5;255;48;5;009m", Cond: avg <= g.bad()},
		{Grade: "D", Color: "38;5;255;48;5;196m", Cond: avg <= g.awful()},
		{Grade: "E", Color: "38;5;255;48;5;124m", Cond: avg <= g.worst()},
		{Grade: "~", Color: "38;5;008;48;5;007m", Cond: true},
	}

	for _, item := range scores {
		if item.Cond {
			level = item
			break
		}
	}

	return fmt.Sprintf(
		"\033[%s Performance: %s \033[0m",
		level.Color,
		pad(level.Grade, 3),
	)
}
