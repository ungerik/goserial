// +build linux,!cgo

package serial

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"syscall"
// 	"unsafe"
// )

// func openPort(name string, c *Config) (rwc io.ReadWriteCloser, err error) {

// 	var rate uint32
// 	switch c.Baud {
// 	case 115200:
// 		rate = syscall.B115200
// 	case 57600:
// 		rate = syscall.B57600
// 	case 38400:
// 		rate = syscall.B38400
// 	case 19200:
// 		rate = syscall.B19200
// 	case 9600:
// 		rate = syscall.B9600
// 	case 4800:
// 		rate = syscall.B4800
// 	default:
// 		return nil, fmt.Errorf("Unknown baud rate %v", c.Baud)
// 	}

// 	var stop uint32
// 	switch c.StopBits {
// 	case StopBits1:
// 		stop = 0
// 	case StopBits2:
// 		stop = syscall.CSTOPB
// 	default:
// 		panic("should not happen if Config.check() was called before")
// 	}

// 	var size uint32
// 	switch c.Size {
// 	case Byte5:
// 		size = syscall.CS5
// 	case Byte6:
// 		size = syscall.CS6
// 	case Byte7:
// 		size = syscall.CS7
// 	case Byte8:
// 		size = syscall.CS8
// 	default:
// 		panic("should not happen if Config.check() was called before")
// 	}

// 	var parity uint32
// 	switch c.Parity {
// 	case ParityNone:
// 		parity = 0
// 	case ParityEven:
// 		parity = 2
// 	case ParityOdd:
// 		parity = 1
// 	default:
// 		panic("should not happen if Config.check() was called before")
// 	}

// 	f, err := os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0600)
// 	if err != nil {
// 		return nil, err
// 	}

// 	t := syscall.Termios{
// 		Iflag:  syscall.IGNPAR,
// 		Cflag:  stop | size | parity | syscall.CREAD | syscall.CLOCAL | rate,
// 		Cc:     [32]uint8{syscall.VMIN: 1},
// 		Ispeed: rate,
// 		Ospeed: rate,
// 	}

// 	_, _, errno := syscall.Syscall6(
// 		syscall.SYS_IOCTL,
// 		uintptr(f.Fd()),
// 		uintptr(syscall.TCSETS),
// 		uintptr(unsafe.Pointer(&t)),
// 		0,
// 		0,
// 		0,
// 	)
// 	if errno != 0 {
// 		f.Close()
// 		return nil, errno
// 	}

// 	err = syscall.SetNonblock(int(f.Fd()), false)
// 	if err != nil {
// 		f.Close()
// 		return nil, err
// 	}

// 	return f, nil
// }

// possix

// // +build !windows,!darwin

// package serial

// // #include <unistd.h>
// // #include <termios.h>
// import "C"

// // TODO: Maybe change to using syscall package + ioctl instead of cgo

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"sync"
// 	"syscall"
// )

// type Connection struct {
// 	file       *os.File
// 	readMutex  sync.Mutex
// 	writeMutex sync.Mutex
// }

// func (conn *Connection) Read(buf []byte) (int, error) {
// 	conn.readMutex.Lock()
// 	defer conn.readMutex.Unlock()
// 	return conn.file.Read(buf)
// }

// func (conn *Connection) Write(buf []byte) (int, error) {
// 	conn.writeMutex.Lock()
// 	defer conn.writeMutex.Unlock()
// 	return conn.file.Write(buf)
// }

// func (conn *Connection) Close() error {
// 	return conn.file.Close()
// }

// func (conn *Connection) Drain() error {
// 	_, err := C.tcdrain(C.int(conn.file.Fd()))
// 	return err
// }

// func openPort(name string, conf *Config) (conn *Connection, err error) {
// 	var f *os.File
// 	f, err = os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer func() {
// 		if err != nil {
// 			f.Close()
// 		}
// 	}()

// 	fd := C.int(f.Fd())
// 	if C.isatty(fd) != 1 {
// 		err = errors.New("File is not a tty")
// 		return
// 	}

// 	var termios C.struct_termios
// 	_, err = C.tcgetattr(fd, &termios)
// 	if err != nil {
// 		return
// 	}
// 	var speed C.speed_t
// 	switch conf.Baud {
// 	case 115200:
// 		speed = C.B115200
// 	case 57600:
// 		speed = C.B57600
// 	case 38400:
// 		speed = C.B38400
// 	case 19200:
// 		speed = C.B19200
// 	case 9600:
// 		speed = C.B9600
// 	case 4800:
// 		speed = C.B4800
// 	default:
// 		return nil, fmt.Errorf("Unknown baud rate %v", conf.Baud)
// 	}

// 	_, err = C.cfsetispeed(&termios, speed)
// 	if err != nil {
// 		return
// 	}
// 	_, err = C.cfsetospeed(&termios, speed)
// 	if err != nil {
// 		return
// 	}

// 	// Select local mode
// 	termios.c_cflag |= C.CLOCAL | C.CREAD

// 	// Select stop bits
// 	switch conf.StopBits {
// 	case StopBits1:
// 		termios.c_cflag &^= C.CSTOPB
// 	case StopBits2:
// 		termios.c_cflag |= C.CSTOPB
// 	default:
// 		panic(conf.StopBits)
// 	}

// 	// Select character size
// 	termios.c_cflag &^= C.CSIZE
// 	switch conf.Size {
// 	case Byte5:
// 		termios.c_cflag |= C.CS5
// 	case Byte6:
// 		termios.c_cflag |= C.CS6
// 	case Byte7:
// 		termios.c_cflag |= C.CS7
// 	case Byte8:
// 		termios.c_cflag |= C.CS8
// 	default:
// 		panic(conf.Size)
// 	}

// 	// Select parity mode
// 	switch conf.Parity {
// 	case ParityNone:
// 		termios.c_cflag &^= C.PARENB
// 	case ParityEven:
// 		termios.c_cflag |= C.PARENB
// 		termios.c_cflag &^= C.PARODD
// 	case ParityOdd:
// 		termios.c_cflag |= C.PARENB
// 		termios.c_cflag |= C.PARODD
// 	default:
// 		panic(conf.Parity)
// 	}

// 	// Select CRLF translation
// 	if conf.CRLFTranslate {
// 		termios.c_iflag |= C.ICRNL
// 	} else {
// 		termios.c_iflag &^= C.ICRNL
// 	}

// 	// Select raw mode
// 	termios.c_lflag &^= C.ICANON | C.ECHO | C.ECHOE | C.ISIG
// 	termios.c_oflag &^= C.OPOST

// 	// if conf.Timeout != 0 {
// 	// 	termios.c_cc[C.VMIN] = 0
// 	// 	termios.c_cc[C.VTIME] = C.cc_t(conf.Timeout / (time.Second / 10))
// 	// }

// 	_, err = C.tcsetattr(fd, C.TCSANOW, &termios)
// 	if err != nil {
// 		return
// 	}

// 	// flags, _, _ := syscall.Syscall(syscall.SYS_FCNTL, uintptr(f.Fd()), uintptr(syscall.F_GETFL), 0)
// 	// if int(flags) == -1 {
// 	// 	err = errors.New("syscall.F_GETFL failed")
// 	// 	return
// 	// }
// 	// flags |= syscall.O_NONBLOCK
// 	// flags &^= syscall.O_NONBLOCK
// 	// r1, _, _ := syscall.Syscall(syscall.SYS_FCNTL, uintptr(f.Fd()), uintptr(syscall.F_SETFL), flags)
// 	// if int(r1) == -1 {
// 	// 	err = errors.New("syscall.F_SETFL failed")
// 	// 	return
// 	// }

// 	return &Connection{file: f}, nil
// }
