// +build darwin

package goserial

import (
	"path/filepath"
	"strings"
)

func listPorts() map[string]string {
	matches, err := filepath.Glob("/dev/cu.*")
	if err != nil {
		panic(err)
	}
	results := make(map[string]string)
	for _, path := range matches {
		name := strings.TrimPrefix(path, "/dev/cu.")
		results[name] = path
	}
	return results
}
