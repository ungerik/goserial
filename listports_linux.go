// +build linux

package serial

import (
	"path/filepath"
	"strings"
)

const PREFIX = "/dev/ttyS"

func listPorts() map[string]string {
	// http://stackoverflow.com/questions/2530096/how-to-find-all-serial-devices-ttys-ttyusb-on-linux-without-opening-them
	matches, err := filepath.Glob("/sys/class/tty/*/device/driver") // ???
	if err != nil {
		panic(err)
	}
	results := make(map[string]string)
	for _, tty := range matches {
		name := strings.TrimSuffix(strings.TrimPrefix(tty, "/sys/class/tty/"), "/device/driver")
		path := "/dev/" + name
		results[name] = path
	}
	return results
}

func isName(name string) bool {
	return strings.HasPrefix(name, PREFIX)
}
