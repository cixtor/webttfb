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

	fmt.Printf("@ Testing domain '%s'\n", tester.domain)
	fmt.Printf("  Status: Connection Time, First Byte Time, Total Time\n")

	for _, data := range results {
		fmt.Printf("- Testing server '%s' -> %s, %s, %s %s\n",
			data.Output.ServerID,
			data.Output.ConnectTime,
			data.Output.FirstbyteTime,
			data.Output.TotalTime,
			data.Output.ServerTitle)
	}

	fmt.Println("  Finished")
}
