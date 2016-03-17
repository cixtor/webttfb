package main

import (
	"flag"
	"fmt"
)

func main() {
	tester, err := NewTTFB()

	if err != nil {
		fmt.Println(err)
		return
	}

	flag.Parse()

	results, err := tester.Report(flag.Arg(0))

	if err != nil {
		fmt.Println(err)
		return
	}

	var icon string

	fmt.Printf("@ Testing domain '%s'\n", tester.domain)
	fmt.Printf("  Status: Connection Time, First Byte Time, Total Time\n")

	for _, data := range results {
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

	fmt.Println("  Finished")
}
