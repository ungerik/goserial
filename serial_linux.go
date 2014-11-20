// +build linux,!cgo

package goserial

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

func openPort(name string, c *Config) (rwc io.ReadWriteCloser, err error) {

	var rate uint32
	switch c.Baud {
	case 115200:
		rate = syscall.B115200
	case 57600:
		rate = syscall.B57600
	case 38400:
		rate = syscall.B38400
	case 19200:
		rate = syscall.B19200
	case 9600:
		rate = syscall.B9600
	case 4800:
		rate = syscall.B4800
	default:
		return nil, fmt.Errorf("Unknown baud rate %v", c.Baud)
	}

	var stop uint32
	switch c.StopBits {
	case StopBits1:
		stop = 0
	case StopBits2:
		stop = syscall.CSTOPB
	default:
		panic("should not happen if Config.check() was called before")
	}

	var size uint32
	switch c.Size {
	case Byte5:
		size = syscall.CS5
	case Byte6:
		size = syscall.CS6
	case Byte7:
		size = syscall.CS7
	case Byte8:
		size = syscall.CS8
	default:
		panic("should not happen if Config.check() was called before")
	}

	var parity uint32
	switch c.Parity {
	case ParityNone:
		parity = 0
	case ParityEven:
		parity = 2
	case ParityOdd:
		parity = 1
	default:
		panic("should not happen if Config.check() was called before")
	}

	f, err := os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return nil, err
	}

	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  stop | size | parity | syscall.CREAD | syscall.CLOCAL | rate,
		Cc:     [32]uint8{syscall.VMIN: 1},
		Ispeed: rate,
		Ospeed: rate,
	}

	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(f.Fd()),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	)
	if errno != 0 {
		return nil, errno
	}

	err = syscall.SetNonblock(int(f.Fd()), false)
	if err != nil {
		return nil, err
	}

	return f, nil
}
