// +build linux

package serial

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-dry"
)

const PREFIX = "/dev/tty"

// https://github.com/wjwwood/serial/blob/master/src/impl/list_ports/list_ports_linux.cc
func listPorts() map[string]string {
	results := make(map[string]string)
	if dry.FileExists("/dev/serial/by-id/") {
		files, err := dry.ListDir("/dev/serial/by-id/")
		if err != nil {
			panic(err)
		}
		for _, name := range files {
			path, err := os.Readlink("/dev/serial/by-id/" + name)
			if err != nil {
				panic(err)
			}
			path, _ = filepath.Abs("/dev/serial/by-id/" + path)
			results[name] = path
		}
	} else {
		// Older Linux don't have /dev/serial
		// use a more general approach:
		matches, err := filepath.Glob("/sys/class/tty/*/device/driver")
		if err != nil {
			panic(err)
		}
		for _, tty := range matches {
			name := strings.TrimSuffix(strings.TrimPrefix(tty, "/sys/class/tty/"), "/device/driver")
			path := "/dev/" + name
			results[name] = path
		}
	}
	return results
}

func isName(name string) bool {
	if strings.HasPrefix(name, PREFIX) {
		return true
	}
	for n := range listPorts() {
		if n == name {
			return true
		}
	}
	return false
}
