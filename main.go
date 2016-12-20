package main

import (
	"flag"
	"fmt"
	"os"
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

	tester, err := NewTTFB(*domain, *private)

	if err != nil {
		fmt.Println(err)
		return
	}

	if *local {
		output, err := tester.LocalTest()

		if err != nil {
			fmt.Println(err)
			return
		}

		/* convert []byte to string */
		fmt.Printf("%s", output)
		return
	}

	var icon string

	fmt.Println("@ Testing domain [" + tester.Domain + "]")
	fmt.Println("  Status:  Conn   TTFB   TTL    Location")

	tester.Analyze()

	for _, data := range tester.Report(*sorting) {
		if data.Status == 1 {
			icon = "\033[0;32m\u2714\033[0m"
		} else {
			icon = "\033[0;31m\u2718\033[0m"
		}

		fmt.Print(icon)
		fmt.Printf(" \033[0;2m%s\033[0m", data.Output.ServerID)
		fmt.Printf("  %s", Colorize("conn", data.Output.ConnectTime))
		fmt.Printf("  %s", Colorize("ttfb", data.Output.FirstbyteTime))
		fmt.Printf("  %s", Colorize("ttl", data.Output.TotalTime))
		fmt.Printf("  %s", data.Output.ServerTitle)
		fmt.Println()
	}

	for _, message := range tester.ErrorMessages() {
		fmt.Println("\033[0;94m\u2022\033[0m " + message.Error())
	}

	fmt.Print("  Average")
	fmt.Printf("  %.3f", tester.Average(connectionTime))
	fmt.Printf("  %.3f", tester.Average(timeToFirstByte))
	fmt.Printf("  %.3f", tester.Average(totalTime))
	fmt.Printf("  %s\n", PerformanceGrade(tester))
	fmt.Println("  Finished")

	os.Exit(0)
}
