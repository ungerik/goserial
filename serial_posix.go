// +build !windows

package goserial

// #include <termios.h>
// #include <unistd.h>
import "C"

// TODO: Maybe change to using syscall package + ioctl instead of cgo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
	//"unsafe"
)

func openPort(name string, conf *Config) (rwc io.ReadWriteCloser, err error) {
	f, err := os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	fd := C.int(f.Fd())
	if C.isatty(fd) != 1 {
		return nil, errors.New("File is not a tty")
	}

	var st C.struct_termios
	_, err = C.tcgetattr(fd, &st)
	if err != nil {
		return nil, err
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

	_, err = C.cfsetispeed(&st, speed)
	if err != nil {
		return nil, err
	}
	_, err = C.cfsetospeed(&st, speed)
	if err != nil {
		return nil, err
	}

	// Select local mode
	st.c_cflag |= C.CLOCAL | C.CREAD

	// Select stop bits
	switch conf.StopBits {
	case StopBits1:
		st.c_cflag &^= C.CSTOPB
	case StopBits2:
		st.c_cflag |= C.CSTOPB
	default:
		panic(conf.StopBits)
	}

	// Select character size
	st.c_cflag &^= C.CSIZE
	switch conf.Size {
	case Byte5:
		st.c_cflag |= C.CS5
	case Byte6:
		st.c_cflag |= C.CS6
	case Byte7:
		st.c_cflag |= C.CS7
	case Byte8:
		st.c_cflag |= C.CS8
	default:
		panic(conf.Size)
	}

	// Select parity mode
	switch conf.Parity {
	case ParityNone:
		st.c_cflag &^= C.PARENB
	case ParityEven:
		st.c_cflag |= C.PARENB
		st.c_cflag &^= C.PARODD
	case ParityOdd:
		st.c_cflag |= C.PARENB
		st.c_cflag |= C.PARODD
	default:
		panic(conf.Parity)
	}

	// Select CRLF translation
	if conf.CRLFTranslate {
		st.c_iflag |= C.ICRNL
	} else {
		st.c_iflag &^= C.ICRNL
	}

	// Select raw mode
	st.c_lflag &^= C.ICANON | C.ECHO | C.ECHOE | C.ISIG
	st.c_oflag &^= C.OPOST

	if conf.Timeout != 0 {
		st.c_cc[C.VMIN] = 0
		st.c_cc[C.VTIME] = C.cc_t(conf.Timeout / (time.Second / 10))
	}

	_, err = C.tcsetattr(fd, C.TCSANOW, &st)
	if err != nil {
		return nil, err
	}

	//fmt.Println("Tweaking", name)
	r1, _, e := syscall.Syscall(syscall.SYS_FCNTL,
		uintptr(f.Fd()),
		uintptr(syscall.F_SETFL),
		uintptr(0))
	if e != 0 || r1 != 0 {
		s := fmt.Sprint("Clearing NONBLOCK syscall error:", e, r1)
		return nil, errors.New(s)
	}

	/*
				r1, _, e = syscall.Syscall(syscall.SYS_IOCTL,
			                uintptr(f.Fd()),
			                uintptr(0x80045402), // IOSSIOSPEED
			                uintptr(unsafe.Pointer(&baud)));
			        if e != 0 || r1 != 0 {
			                s := fmt.Sprint("Baudrate syscall error:", e, r1)
		                        return nil, os.NewError(s)
				}
	*/

	return f, nil
}

func (conn *Connection) Drain() error {
	fd := conn.ReadWriteCloser.(*os.File).Fd()
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
