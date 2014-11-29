package main

import (
	"flag"
	"fmt"
	"os"
	// "sort"
	"text/tabwriter"

	"github.com/ungerik/go-dry"
	"github.com/ungerik/goserial"
)

var (
	names = flag.Bool("names", true, "Print names of serial ports")
	paths = flag.Bool("paths", true, "Print paths of serial ports")
)

func main() {
	// todo extract as unit test:
	// l := dry.StringNumberGroupPostfixSorter{"usb1", "usb24", "usb2", "usb12", "usb0", "usb13", "usb000", "usb010"}
	// sort.Sort(l)
	// for _, s := range l {
	// 	fmt.Println(s)
	// }
	// fmt.Println("\n")

	flag.Parse()
	switch {
	case *names && !*paths:
		for _, short := range goserial.ListPortsShort() {
			fmt.Println(short)
		}
	case !*names && *paths:
		for _, long := range goserial.ListPorts() {
			fmt.Println(long)
		}
	case *names && *paths:
		ports := goserial.ListPortsShortLong()
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
		for _, name := range dry.StringMapGroupedNumberPostfixSortedKeys(ports) {
			fmt.Fprintf(w, "%s\t-> %s\n", name, ports[name])
		}
		w.Flush()
	}
}
