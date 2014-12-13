package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ungerik/go-dry"
	"github.com/ungerik/goserial"
)

var (
	port      string
	baud      int
	maxBytes  int
	quitAfter time.Duration
	timeout   time.Duration

	stop bool
)

func main() {
	flag.StringVar(&port, "port", "", "Serial port to connect to")
	flag.IntVar(&baud, "baud", 57600, "Speed of the connection")
	flag.IntVar(&maxBytes, "max", 100, "Quit program after this number of packets")
	flag.DurationVar(&quitAfter, "quitafter", time.Second*3, "Quit program after this duration")
	flag.DurationVar(&timeout, "timeout", time.Second/10, "Read timeout per packet")
	flag.Parse()

	if port == "" {
		if flag.NArg() > 0 {
			port = flag.Arg(0)
		} else {
			ports := serial.ListPorts()
			if len(ports) == 1 {
				port = ports[0]
			} else {
				fmt.Fprintln(os.Stderr, "Call with -port=PORT")
				flag.PrintDefaults()
				fmt.Fprintln(os.Stderr, "\nAvailable as PORT are:")
				for _, p := range ports {
					fmt.Fprintln(os.Stderr, "  ", p)
				}
				return
			}
		}
	}

	conn, err := serial.OpenDefault(port, serial.Baud(baud), timeout)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Opened serial port", port)
	defer func() {
		log.Println("Closed serial port", port, "with error", conn.Close())
	}()

	go func() {
		dry.WaitForStdin("Press any key to quit")
		stop = true
	}()

	time.AfterFunc(quitAfter, func() { stop = true })

	buf := make([]byte, 1)
	for i := 0; i < maxBytes && !stop; i++ {
		_, err = conn.Read(buf)
		if err == nil {

			log.Printf("%q %d", buf, buf[0])
		} else {
			log.Println(err)
		}
	}

}
