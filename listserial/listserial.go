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
		for name := range ports {
			fmt.Println(name)
		}
	case !*names && *paths:
		for _, path := range ports {
			fmt.Println(path)
		}
	case *names && *paths:
		s := fmt.Sprint(ports)
		s = dry.StringReplaceMulti(s, "map[", "", " ", "\n", ":", "\t-> ", "]", "")
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
		fmt.Fprintln(w, s)
		w.Flush()
	}
}
