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

	if rate == 0 {
		return
	}

	var stop byte
	switch c.StopBits {
	case StopBits1:
		stop = 0
	case StopBits2:
		stop = syscall.CSTOPB
	default:
		panic(c.StopBits)
	}

	var size byte
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
		panic(c.Size)
	}

	if size == 0 {
		return
	}

	var parity byte
	switch c.Parity {
	case ParityNone:
		parity = 0
	case ParityEven:
		parity = 2
	case ParityOdd:
		parity = 1
	default:
		panic(c.Parity)
	}

	f, err := os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil && f != nil {
			f.Close()
		}
	}()

	fd := f.Fd()
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  stop | size | parity | syscall.CREAD | syscall.CLOCAL | rate,
		Cc:     [32]uint8{syscall.VMIN: 1},
		Ispeed: rate,
		Ospeed: rate,
	}

	if _, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	); errno != 0 {
		return nil, errno
	}

	if err = syscall.SetNonblock(int(fd), false); err != nil {
		return
	}

	return f, nil
}
