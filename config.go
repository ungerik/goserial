package goserial

type ParityMode byte

const (
	ParityNone = ParityMode(iota)
	ParityEven
	ParityOdd
)

type ByteSize byte

const (
	Byte8 = ByteSize(iota)
	Byte5
	Byte6
	Byte7
)

type StopBits byte

const (
	StopBits1 = StopBits(iota)
	StopBits2
)

// Config contains the information needed to open a serial port.
//
// Currently few options are implemented, but more may be added in the
// future (patches welcome), so it is recommended that you create a
// new config addressing the fields by name rather than by order.
//
// For example:
//
//    c0 := &serial.Config{Name: "COM45", Baud: 115200}
// or
//    c1 := new(serial.Config)
//    c1.Name = "/dev/tty.usbserial"
//    c1.Baud = 115200
//
type Config struct {
	Name string
	Baud int

	Size     ByteSize
	Parity   ParityMode
	StopBits StopBits

	// RTSFlowControl bool
	// DTRFlowControl bool
	// XONFlowControl bool

	CRLFTranslate bool // Ignored on Windows.
	// TimeoutStuff int
}

func (c *Config) check() error {
	switch c.Size {
	case Byte5, Byte6, Byte7, Byte8:
	default:
		return ErrConfigByteSize
	}

	switch c.StopBits {
	case StopBits1, StopBits2:
	default:
		return ErrConfigStopBits
	}

	switch c.Parity {
	case ParityNone, ParityEven, ParityOdd:
	default:
		return ErrConfigParity
	}

	return nil
}
