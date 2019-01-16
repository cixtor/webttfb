package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const config string = ".webttfb.cfg"
const service string = "https://performance.sucuri.net/index.php?ajaxcall"
const timeToFirstByte string = "ttfb"
const connectionTime string = "conn"
const totalTime string = "ttl"

var domain = flag.String("d", "example.com", "Domain name to be tested")
var sorting = flag.String("s", "status", "Criteria to sort the results")
var private = flag.Bool("p", false, "Hide results from public stats")
var local = flag.Bool("l", false, "Run the tests with local resources")

func main() {
	flag.Usage = func() {
		fmt.Println("Website TTFB")
		fmt.Println("https://cixtor.com/")
		fmt.Println("https://performance.sucuri.net/")
		fmt.Println("https://github.com/cixtor/webttfb")
		fmt.Println("https://en.wikipedia.org/wiki/Time_To_First_Byte")
		fmt.Println()
		fmt.Println("Time To First Byte (TTFB) is a measurement used as an indication of the")
		fmt.Println("responsiveness of a webserver or other network resource. TTFB measures the")
		fmt.Println("duration from the user or client making an HTTP request to the first byte of the")
		fmt.Println("page being received by the client's browser. This time is made up of the socket")
		fmt.Println("connection time, the time taken to send the HTTP request, and the time taken to")
		fmt.Println("get the first byte of the page.")
		fmt.Println()
		fmt.Println("Sorting: status, conn, ttfb, ttl")
		fmt.Println()
		fmt.Println("Usage:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Abbrs:")
		fmt.Println("  Time is measured in seconds")
		fmt.Println("  Performance is based on TTL")
		fmt.Println("  Conn — Connection Time")
		fmt.Println("  TTFB — Time To First Byte")
		fmt.Println("  TTL  — Total Time")
		os.Exit(2)
	}

	flag.Parse()

	var err error
	var icon string
	var tester *TTFB
	var output []byte

	if tester, err = NewTTFB(*domain, *private); err != nil {
		fmt.Println(err)
		return
	}

	if *local {
		if output, err = tester.LocalTest(); err != nil {
			fmt.Println(err)
			return
		}

		/* convert []byte to string */
		fmt.Printf("%s\n", output)
		return
	}

	tester.Analyze()

	fmt.Println("    ┌─────────┬───────┬───────┬───────┬────────────────────┐")
	fmt.Println("    │ Server  │ Conn  │ TTFB  │ TTL   │ Location           │")
	fmt.Println("┌───┼─────────┼───────┼───────┼───────┼────────────────────┤")

	for _, data := range tester.Report(*sorting) {
		if data.Status == 1 {
			icon = "\033[0;32m\u2714\033[0m"
		} else {
			icon = "\033[0;31m\u2718\033[0m"
		}

		fmt.Printf(
			"│ %s │ \033[0;2m%s\033[0m │ %s │ %s │ %s │ %s │\n",
			icon,
			data.Output.ServerID,
			Colorize("conn", data.Output.ConnectTime),
			Colorize("ttfb", data.Output.FirstbyteTime),
			Colorize("ttl", data.Output.TotalTime),
			pad(data.Output.ServerTitle, 18),
		)
	}

	fmt.Println("└───┼─────────┼───────┼───────┼───────┼────────────────────┤")

	fmt.Printf(
		"    │ Average │ %.3f │ %.3f │ %.3f │ %s │\n",
		tester.Average(connectionTime),
		tester.Average(timeToFirstByte),
		tester.Average(totalTime),
		PerformanceGrade(tester),
	)

	fmt.Println("    └─────────┴───────┴───────┴───────┴────────────────────┘")

	for _, message := range tester.ErrorMessages() {
		fmt.Println("\033[0;94m\u2022\033[0m " + message.Error())
	}

	os.Exit(0)
}

func pad(text string, length int) string {
	largo := len(text)

	if largo > length {
		return text[0:length-1] + "…"
	}

	return text + strings.Repeat("\x20", length-largo)
}
