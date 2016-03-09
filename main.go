package main

import (
	"flag"
	"fmt"
	"os"
)

var domain = flag.String("d", "example.com", "Domain name to be tested")
var sorting = flag.String("s", "status", "Criteria to sort the results")
var private = flag.Bool("p", false, "Hide results from public stats")

func main() {
	tester, err := NewTTFB()

	if err != nil {
		fmt.Println(err)
		return
	}

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

	var icon string

	tester.Analyze(*domain, *private)

	fmt.Printf("@ Testing domain '%s'\n", tester.Domain)
	fmt.Printf("  Status: Connection Time, First Byte Time, Total Time\n")

	for _, data := range tester.Report(*sorting) {
		if data.Status == 1 {
			icon = "\033[0;92m\u2714\033[0m"
		} else {
			icon = "\033[0;91m\u2718\033[0m"
		}

		fmt.Printf("%s %s -> %s, %s, %s %s\n",
			icon,
			data.Output.ServerID,
			data.Output.ConnectTime,
			data.Output.FirstbyteTime,
			data.Output.TotalTime,
			data.Output.ServerTitle)
	}

	for _, message := range tester.Messages() {
		fmt.Println("\033[0;91m\u2718\033[0m " + message.Error())
	}

	fmt.Println("  Finished")
}
