// +build !windows

package goserial

// #include <termios.h>
// #include <unistd.h>
import "C"

// TODO: Maybe change to using syscall package + ioctl instead of cgo

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

type Connection struct {
	*os.File
}

func openPort(name string, conf *Config) (conn *Connection, err error) {
	var f *os.File
	f, err = os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	fd := C.int(f.Fd())
	if C.isatty(fd) != 1 {
		err = errors.New("File is not a tty")
		return
	}

	var termios C.struct_termios
	_, err = C.tcgetattr(fd, &termios)
	if err != nil {
		return
	}
	var speed C.speed_t
	switch conf.Baud {
	case 115200:
		speed = C.B115200
	case 57600:
		speed = C.B57600
	case 38400:
		speed = C.B38400
	case 19200:
		speed = C.B19200
	case 9600:
		speed = C.B9600
	case 4800:
		speed = C.B4800
	default:
		return nil, fmt.Errorf("Unknown baud rate %v", conf.Baud)
	}

	_, err = C.cfsetispeed(&termios, speed)
	if err != nil {
		return
	}
	_, err = C.cfsetospeed(&termios, speed)
	if err != nil {
		return
	}

	// Select local mode
	termios.c_cflag |= C.CLOCAL | C.CREAD

	// Select stop bits
	switch conf.StopBits {
	case StopBits1:
		termios.c_cflag &^= C.CSTOPB
	case StopBits2:
		termios.c_cflag |= C.CSTOPB
	default:
		panic(conf.StopBits)
	}

	// Select character size
	termios.c_cflag &^= C.CSIZE
	switch conf.Size {
	case Byte5:
		termios.c_cflag |= C.CS5
	case Byte6:
		termios.c_cflag |= C.CS6
	case Byte7:
		termios.c_cflag |= C.CS7
	case Byte8:
		termios.c_cflag |= C.CS8
	default:
		panic(conf.Size)
	}

	// Select parity mode
	switch conf.Parity {
	case ParityNone:
		termios.c_cflag &^= C.PARENB
	case ParityEven:
		termios.c_cflag |= C.PARENB
		termios.c_cflag &^= C.PARODD
	case ParityOdd:
		termios.c_cflag |= C.PARENB
		termios.c_cflag |= C.PARODD
	default:
		panic(conf.Parity)
	}

	// Select CRLF translation
	if conf.CRLFTranslate {
		termios.c_iflag |= C.ICRNL
	} else {
		termios.c_iflag &^= C.ICRNL
	}

	// Select raw mode
	termios.c_lflag &^= C.ICANON | C.ECHO | C.ECHOE | C.ISIG
	termios.c_oflag &^= C.OPOST

	// if conf.Timeout != 0 {
	// 	termios.c_cc[C.VMIN] = 0
	// 	termios.c_cc[C.VTIME] = C.cc_t(conf.Timeout / (time.Second / 10))
	// }

	_, err = C.tcsetattr(fd, C.TCSANOW, &termios)
	if err != nil {
		return
	}

	// flags, _, _ := syscall.Syscall(syscall.SYS_FCNTL, uintptr(f.Fd()), uintptr(syscall.F_GETFL), 0)
	// if int(flags) == -1 {
	// 	err = errors.New("syscall.F_GETFL failed")
	// 	return
	// }
	// flags |= syscall.O_NONBLOCK
	// flags &^= syscall.O_NONBLOCK
	// r1, _, _ := syscall.Syscall(syscall.SYS_FCNTL, uintptr(f.Fd()), uintptr(syscall.F_SETFL), flags)
	// if int(r1) == -1 {
	// 	err = errors.New("syscall.F_SETFL failed")
	// 	return
	// }

	return &Connection{f}, nil
}

func (conn *Connection) Drain() error {
	fd := conn.Fd()
	err := syscall.Errno(C.tcdrain(C.int(fd)))
	if err != 0 {
		return err
	}
	return nil

	// var options C.struct_termios
	// _, err := C.tcgetattr(C.int(f.Fd()), &options)
	// if err != nil {
	// 	return err
	// }
	// _, err = C.tcsetattr(C.int(f.Fd()), C.TCSAFLUSH, &options)
	// return err
}
