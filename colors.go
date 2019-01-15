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
		return fmt.Sprintf("\033[48;5;009m%.3f\033[0m", value)
	}

	if value > c.warning() {
		return fmt.Sprintf("\033[48;5;226m%.3f\033[0m", value)
	}

	if value < c.success() {
		return fmt.Sprintf("\033[48;5;034m%.3f\033[0m", value)
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
	var g Grade
	var grade string
	var color string

	average := t.Average(totalTime)

	// Assign a better grade only if most tests succeeded.
	if len(t.Messages) > 4 || average <= 0 {
		grade = "F"
		color = "007"
	} else if average <= g.perfect() {
		grade = "A+"
		color = "014"
	} else if average <= g.excellent() {
		grade = "A"
		color = "034"
	} else if average <= g.good() {
		grade = "B"
		color = "226"
	} else if average <= g.bad() {
		grade = "C"
		color = "9"
	} else if average <= g.awful() {
		grade = "D"
		color = "09"
	} else if average <= g.worst() {
		grade = "E"
		color = "009"
	} else {
		grade = "~"
		color = "007"
	}

	return fmt.Sprintf("Performance: \033[48;5;%sm %s \033[0m", color, grade)
}
