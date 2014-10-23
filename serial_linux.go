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
		speed = syscall.B115200
	case 57600:
		speed = syscall.B57600
	case 38400:
		speed = syscall.B38400
	case 19200:
		speed = syscall.B19200
	case 9600:
		speed = syscall.B9600
	case 4800:
		speed = syscall.B4800
	default:
		return nil, fmt.Errorf("Unknown baud rate %v", c.Baud)
	}

	if rate == 0 {
		return
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
		Cflag:  syscall.CS8 | syscall.CREAD | syscall.CLOCAL | rate,
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
