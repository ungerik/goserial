// +build windows

package goserial

// include <Windows.h>
import "C"

import (
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"
)

func listPorts() map[string]string {
	var buffer [1024]byte
	for i := 0; i < 256; i++ {
		name := C.fmt.Sprintf("COM%d", i)
		lpDeviceName := C.LPCTSTR(C.CString(name))
		lpTargetPath := C.LPTSTR(unsafe.Pointer(&buffer))
		result := C.QueryDosDevice(lpDeviceName, lpTargetPath, len(buffer))
		if result == 0 {
			results[name] = name
		}
	}
	return results
}
