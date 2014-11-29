package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ungerik/go-dry"
	"github.com/ungerik/goserial"
)

var (
	names = flag.Bool("names", true, "Print names of serial ports")
	paths = flag.Bool("paths", true, "Print paths of serial ports")
)

func main() {
	flag.Parse()
	ports := goserial.ListPorts()
	switch {
	case *names && !*paths:
		for _, name := range dry.StringMapSortedKeys(ports) {
			fmt.Println(name)
		}
	case !*names && *paths:
		for _, name := range dry.StringMapSortedKeys(ports) {
			fmt.Println(ports[name])
		}
	case *names && *paths:
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
		for _, name := range dry.StringMapSortedKeys(ports) {
			fmt.Fprintf(w, "%s\t-> %s\n", name, ports[name])
		}
		w.Flush()
	}
}
