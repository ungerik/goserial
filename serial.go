/*
Goserial is a simple go package to allow you to read and write from
the serial port as a stream of bytes.

It aims to have the same API on all platforms, including windows.  As
an added bonus, the windows package does not use cgo, so you can cross
compile for windows from another platform.  Unfortunately goinstall
does not currently let you cross compile so you will have to do it
manually:

    GOOS=windows make clean install

Currently there is very little in the way of configurability.  You can
set the baud rate.  Then you can Read(), Write(), or Close() the
connection.  Read() will block until at least one byte is returned.
Write is the same.  There is currently no exposed way to set the
timeouts, though patches are welcome.

Currently ports are opened with 8 data bits, 1 stop bit, no parity, no hardware
flow control, and no software flow control by default.  This works fine for
many real devices and many faux serial devices including usb-to-serial
converters and bluetooth serial ports.

You may Read() and Write() simulantiously on the same connection (from
different goroutines).

Example usage:

    package main

    import (
        "github.com/tarm/goserial"
        "log"
    )

    func main() {
        c := &serial.Config{Name: "COM5", Baud: 115200}
        s, err := serial.OpenPort(c)
        if err != nil {
                log.Fatal(err)
        }

        n, err := s.Write([]byte("test"))
        if err != nil {
                log.Fatal(err)
        }

        buf := make([]byte, 128)
        n, err = s.Read(buf)
        if err != nil {
                log.Fatal(err)
        }
        log.Print("%q", buf[:n])
    }
*/
package goserial

import (
	"errors"
	"io"

	"github.com/ungerik/go-dry"
)

var (
	ErrConfigStopBits = errors.New("goserial config: bad number of stop bits")
	ErrConfigByteSize = errors.New("goserial config: bad byte size")
	ErrConfigParity   = errors.New("goserial config: bad parity")
)

func ListPortsShortLong() map[string]string {
	return listPorts()
}

func ListPorts() []string {
	return dry.StringMapGroupedNumberPostfixSortedValues(listPorts())
}

func ListPortsShort() []string {
	return dry.StringMapGroupedNumberPostfixSortedKeys(listPorts())
}

type Connection struct {
	io.ReadWriteCloser
}

// OpenPort opens a serial port with the specified configuration
func OpenPort(c *Config) (conn *Connection, err error) {
	if err := c.check(); err != nil {
		return nil, err
	}

	var connection Connection
	connection.ReadWriteCloser, err = openPort(c.Name, c)
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

// func Flush()
//  tcsetattr(fileDescriptor, TCSAFLUSH, &myNewTTYoptions);

// func SendBreak()

// func RegisterBreakHandler(func())
