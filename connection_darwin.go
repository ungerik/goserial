// +build darwin

package goserial

// #cgo LDFLAGS: -framework IOKit
// #cgo LDFLAGS: -framework CoreFoundation
// #include <stdio.h>
// #include <string.h>
// #include <unistd.h>
// #include <fcntl.h>
// #include <sys/ioctl.h>
// #include <errno.h>
// #include <paths.h>
// #include <termios.h>
// #include <sysexits.h>
// #include <sys/param.h>
// #include <sys/select.h>
// #include <sys/time.h>
// #include <time.h>
// #include <CoreFoundation/CoreFoundation.h> 
// #include <IOKit/IOKitLib.h>
// #include <IOKit/serial/IOSerialKeys.h>
// #include <IOKit/IOBSD.h>
import "C"

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

var bauds = map[int]C.speed_t{
	2400:   C.B2400,
	4800:   C.B4800,
	9600:   C.B9600,
	19200:  C.B19200,
	38400:  C.B38400,
	57600:  C.B57600,
	115200: C.B115200,
	230400: C.B230400,
}

type Connection struct {
	file       *os.File
	readMutex  sync.Mutex
	writeMutex sync.Mutex
}

func (conn *Connection) Read(buf []byte) (int, error) {
	conn.readMutex.Lock()
	defer conn.readMutex.Unlock()
	return conn.file.Read(buf)
}

func (conn *Connection) Write(buf []byte) (int, error) {
	conn.writeMutex.Lock()
	defer conn.writeMutex.Unlock()
	return conn.file.Write(buf)
}

func (conn *Connection) Close() error {
	fmt.Println("DEBUG: Closing", conn.file)
	return conn.file.Close()
}

func (conn *Connection) Drain() error {
	_, err := C.tcdrain(C.int(conn.file.Fd()))
	return err
}

// See https://developer.apple.com/library/mac/documentation/DeviceDrivers/Conceptual/WorkingWSerial/WWSerial_SerialDevs/SerialDevices.html
func openPort(name string, conf *Config) (conn *Connection, err error) {
	conn = new(Connection)
	conn.file, err = os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			fmt.Println("DEBUG: Error in openPort(), closing file.")
			conn.file.Close()
			conn = nil
		}
	}()

	fd := C.int(conn.file.Fd())

	if C.isatty(fd) != 1 {
		err = errors.New("File is not a tty")
		return
	}

	// Note that open() follows POSIX semantics: multiple open() calls to
	// the same file will succeed unless the TIOCEXCL ioctl is issued.
	// This will prevent additional opens except by root-owned processes.
	// See tty(4) ("man 4 tty") and ioctl(2) ("man 2 ioctl") for details.
	r0, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), C.TIOCEXCL, 0)
	if r0 != 0 {
		err = fmt.Errorf("Error setting TIOCEXCL: %s", errno)
		return
	}

	// Clear the O_NONBLOCK flag so subsequent I/O will block
	// See fcntl(2) ("man 2 fcntl") for details.
	r0, _, errno = syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), C.F_SETFL, 0)
	if r0 != 0 {
		err = fmt.Errorf("Error clearing O_NONBLOCK: %s", errno)
		return
	}

	// Get the current options and save them so we can restore the
	// default settings later.
	var r C.int
	var origTermiosSettings C.struct_termios
	r, err = C.tcgetattr(fd, &origTermiosSettings)
	if r != 0 {
		err = fmt.Errorf("Error getting serial attributes: %s", err)
		return
	}

	// The serial port attributes such as timeouts and baud rate are set by
	// modifying the termios structure and then calling tcsetattr to
	// cause the changes to take effect. Note that the
	// changes will not take effect without the tcsetattr() call.
	// See tcsetattr(4) ("man 4 tcsetattr") for details.
	termios := origTermiosSettings

	// Sets the terminal to something like the "raw" mode of the old Version 7 terminal driver:
	// input is available character by character, echoing is disabled,
	// and all special processing of terminal input and output characters is disabled. 
	C.cfmakeraw(&termios)

	// See http://www.unixwiz.net/techtips/termios-vmin-vtime.html
	termios.c_cc[C.VMIN] = 0
	termios.c_cc[C.VTIME] = C.cc_t(conf.Timeout / (time.Second / 10))

	speed, ok := bauds[conf.Baud]
	if !ok {
		err = fmt.Errorf("Unknown baud rate %d", conf.Baud)
		return
	}
	r, err = C.cfsetispeed(&termios, speed)
	if r != 0 {
		err = fmt.Errorf("Error setting input speed: %d (%s)", speed, err)
		return
	}
	r, err = C.cfsetospeed(&termios, speed)
	if r != 0 {
		err = fmt.Errorf("Error setting output speed: %d (%s)", speed, err)
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

	r, err = C.tcsetattr(fd, C.TCSANOW, &termios)
	if r != 0 {
		err = fmt.Errorf("Error setting serial options: %s", err)
		return
	}

	return conn, nil
}
