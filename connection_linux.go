// +build linux

package serial

// #include <unistd.h>
// #include <fcntl.h>
// #include <sys/ioctl.h>
// #include <termios.h>
import "C"

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

var bauds = map[Baud]C.speed_t{
	2400:   C.B2400,
	4800:   C.B4800,
	9600:   C.B9600,
	19200:  C.B19200,
	38400:  C.B38400,
	57600:  C.B57600,
	115200: C.B115200,
	230400: C.B230400,
}

var byteSizes = map[ByteSize]C.tcflag_t{
	ByteSize5: C.CS5,
	ByteSize6: C.CS6,
	ByteSize7: C.CS7,
	ByteSize8: C.CS8,
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

// https://github.com/ynezz/librs232/blob/master/src/rs232_posix.c
// http://sigrok.org/wiki/Libserialport

func openPort(name string, baud Baud, byteSize ByteSize, parity ParityMode, stopBits StopBits, readTimeout time.Duration) (conn *Connection, err error) {
	var file *os.File
	file, err = os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			fmt.Println("DEBUG: Error in openPort(), closing file.")
			file.Close()
		}
	}()

	fd := C.int(file.Fd())

	if C.isatty(fd) != 1 {
		err = errors.New("File is not a tty")
		return
	}

	// // Note that open() follows POSIX semantics: multiple open() calls to
	// // the same file will succeed unless the TIOCEXCL ioctl is issued.
	// // This will prevent additional opens except by root-owned processes.
	// // See tty(4) ("man 4 tty") and ioctl(2) ("man 2 ioctl") for details.
	r0, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), C.TIOCEXCL, 0)
	if r0 != 0 {
		err = fmt.Errorf("Error setting TIOCEXCL: %s", errno)
		return
	}

	// Clear the O_NONBLOCK flag so subsequent I/O will block
	flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), C.F_GETFL, 0)
	if errno != 0 {
		err = fmt.Errorf("Error clearing O_NONBLOCK: %s", errno)
		return
	}

	flags &^= C.O_NONBLOCK

	r0, _, errno = syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), C.F_SETFL, flags)
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

	defer func() {
		C.tcsetattr(fd, C.TCSANOW, &origTermiosSettings)
	}()

	// The serial port attributes such as timeouts and baud rate are set by
	// modifying the termios structure and then calling tcsetattr to
	// cause the changes to take effect. Note that the
	// changes will not take effect without the tcsetattr() call.
	// See tcsetattr(4) ("man 4 tcsetattr") for details.
	// termios := origTermiosSettings

	var termios C.struct_termios

	// IGNPAR: ignore bytes with parity errors
	termios.c_iflag = C.IGNPAR

	// Select local mode
	termios.c_cflag = C.CLOCAL | C.CREAD

	// // Sets the terminal to something like the "raw" mode of the old Version 7 terminal driver:
	// // input is available character by character, echoing is disabled,
	// // and all special processing of terminal input and output characters is disabled.
	// C.cfmakeraw(&termios)

	// // Turn off all fancy termios tricks, give us a raw channel
	// termios.c_iflag &^= (C.IGNBRK | C.BRKINT | C.PARMRK | C.ISTRIP | C.INLCR | C.IGNCR | C.ICRNL | C.IMAXBEL)
	// termios.c_lflag &^= (C.ISIG | C.ICANON | C.ECHO | C.IEXTEN)
	// termios.c_lflag &^= (C.ICANON | C.ECHO | C.ECHOE | C.ECHOK | C.ECHONL | C.ISIG | C.IEXTEN)

	// // Select raw mode
	// termios.c_lflag &^= C.ICANON | C.ECHO | C.ECHOE | C.ISIG
	// termios.c_oflag &^= C.OPOST

	// // Disable flow control
	// termios.c_cflag &^= C.CRTSCTS
	// termios.c_iflag &^= (C.IXON | C.IXOFF | C.IXANY)

	// See http://www.unixwiz.net/techtips/termios-vmin-vtime.html
	termios.c_cc[C.VMIN] = 0
	termios.c_cc[C.VTIME] = C.cc_t(readTimeout / (time.Second / 10))

	speed, ok := bauds[baud]
	if !ok {
		err = fmt.Errorf("Unknown baud rate %d", baud)
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

	switch stopBits {
	case StopBits1:
		termios.c_cflag &^= C.CSTOPB
	case StopBits2:
		termios.c_cflag |= C.CSTOPB
	default:
		err = fmt.Errorf("Bad number of stop bits: %d", stopBits)
		return
	}

	size, ok := byteSizes[byteSize]
	if !ok {
		err = fmt.Errorf("Bad byte size: %d", byteSize)
		return
	}
	termios.c_cflag &^= C.CSIZE
	termios.c_cflag |= size

	// Select parity mode
	switch parity {
	case ParityModeNone:
		termios.c_cflag &^= C.PARENB
	case ParityModeEven:
		termios.c_cflag |= C.PARENB
		termios.c_cflag &^= C.PARODD
	case ParityModeOdd:
		termios.c_cflag |= C.PARENB
		termios.c_cflag |= C.PARODD
	default:
		err = errors.New("goserial config: bad parity")
		return
	}

	r, err = C.tcflush(fd, C.TCIOFLUSH)
	if r != 0 {
		err = fmt.Errorf("Error flushing connection: %s", err)
		return
	}

	r, err = C.tcsetattr(fd, C.TCSANOW, &termios)
	if r != 0 {
		err = fmt.Errorf("Error setting serial options: %s", err)
		return
	}

	return &Connection{file: file}, nil
}
