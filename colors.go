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

// Paint builds the escape sequence to render the colors.
func Paint(c Colorizer, value float64) string {
	if value > c.danger() {
		return fmt.Sprintf("\033[48;5;009m%.3f\033[0m", value)
	}

	if value > c.warning() {
		return fmt.Sprintf("\033[48;5;011m%.3f\033[0m", value)
	}

	if value < c.success() {
		return fmt.Sprintf("\033[48;5;010m%.3f\033[0m", value)
	}

	return fmt.Sprintf("%.3f", value)
}

// Colorize returns the floating point with a background color.
func Colorize(group string, value float64) string {
	if value == 0.0 {
		// Do not colorize failed tests.
		return fmt.Sprintf("%.3f", value)
	}

	if group == "conn" {
		return Paint(ColorConn{}, value)
	}

	if group == "ttfb" {
		return Paint(ColorTTFB{}, value)
	}

	if group == "ttl" {
		return Paint(ColorTTL{}, value)
	}

	return fmt.Sprintf("%.3f", value)
}
