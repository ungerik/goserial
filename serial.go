package serial

import (
	"time"

	"github.com/ungerik/go-dry"
)

type Baud int

const (
	Baud2400   Baud = 2400
	Baud4800   Baud = 4800
	Baud9600   Baud = 9600
	Baud19200  Baud = 19200
	Baud38400  Baud = 38400
	Baud57600  Baud = 57600
	Baud115200 Baud = 115200
	Baud230400 Baud = 230400
)

type ParityMode int

const (
	ParityModeNone = 0
	ParityModeOdd  = 1
	ParityModeEven = 2
)

type ByteSize int

const (
	ByteSize5 = 5
	ByteSize6 = 6
	ByteSize7 = 7
	ByteSize8 = 8
)

type StopBits int

const (
	StopBits1 = 1
	StopBits2 = 2
)

func ListPorts() []string {
	return dry.StringMapGroupedNumberPostfixSortedValues(listPorts())
}

func ListPortsShort() []string {
	return dry.StringMapGroupedNumberPostfixSortedKeys(listPorts())
}

func ListPortsShortLong() map[string]string {
	return listPorts()
}

// IsName returns if name mathes the pattern for serial ports.
// This does not mean, that there is an actual serial port with
// this name currently active. Use IsPort to get this information.
func IsName(name string) bool {
	return isName(name)
}

// IsPort returns if port is one of the long or short names that
// ListPort returns.
func IsPort(port string) bool {
	for short, long := range ListPortsShortLong() {
		if port == short || port == long {
			return true
		}
	}
	return false
}

// OpenDefault calls Open with ByteSize8, ParityModeNone, StopBits1.
func OpenDefault(port string, baud Baud, readTimeout time.Duration) (*Connection, error) {
	return openPort(port, baud, ByteSize8, ParityModeNone, StopBits1, readTimeout)
}

func Open(port string, baud Baud, byteSize ByteSize, parity ParityMode, stopBits StopBits, readTimeout time.Duration) (*Connection, error) {
	return openPort(port, baud, byteSize, parity, stopBits, readTimeout)
}
