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
		fmt.Println("Sorting options: status, conn, ttfb, ttl")
		fmt.Println()
		fmt.Println("Usage:")
		flag.PrintDefaults()
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

		fmt.Printf("@ Testing domain '%s'\n", tester.Domain)
		fmt.Printf("%s", output) /* convert []byte to string */
		fmt.Printf("  Finished\n")
		return
	}

	var icon string
	tester.Analyze()

	fmt.Println("@ Testing domain [" + tester.Domain + "]")
	fmt.Println("# Status:  Conn   TTFB   TTL    Location")
	fmt.Println()

	for _, data := range tester.Report(*sorting) {
		if data.Status == 1 {
			icon = "\033[0;32m\u2714\033[0m"
		} else {
			icon = "\033[0;31m\u2718\033[0m"
		}

		fmt.Print(icon)
		fmt.Print("\x20")
		fmt.Print("\033[0;2m" + data.Output.ServerID + "\033[0m")
		fmt.Print("\x20\x20")
		fmt.Print(Colorize("conn", data.Output.ConnectTime))
		fmt.Print("\x20\x20")
		fmt.Print(Colorize("ttfb", data.Output.FirstbyteTime))
		fmt.Print("\x20\x20")
		fmt.Print(Colorize("ttl", data.Output.TotalTime))
		fmt.Print("\x20\x20")
		fmt.Print(data.Output.ServerTitle)
		fmt.Println()
	}

	for _, message := range tester.ErrorMessages() {
		fmt.Println("\033[0;91m\u2718\033[0m " + message.Error())
	}

	fmt.Print("\x20\x20")
	fmt.Print("Average")
	fmt.Print("\x20\x20")
	fmt.Printf("%.3f", tester.Average(connectionTime))
	fmt.Print("\x20\x20")
	fmt.Printf("%.3f", tester.Average(timeToFirstByte))
	fmt.Print("\x20\x20")
	fmt.Printf("%.3f", tester.Average(totalTime))
	fmt.Println()

	fmt.Println()
	fmt.Println("* Time is in seconds")
	fmt.Println("* Conn — Connection Time")
	fmt.Println("* TTFB — Time To First Byte")
	fmt.Println("* TTL  — Total Time")
}
