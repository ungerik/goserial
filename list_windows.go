// +build windows

package goserial

// #include <Windows.h>
import "C"

import (
	"fmt"
	"unsafe"
)

func listPorts() map[string]string {
	var buffer [1024]byte
	lpTargetPath := C.LPTSTR(unsafe.Pointer(&buffer))
	results := make(map[string]string)

	for i := 0; i < 256; i++ {
		name := fmt.Sprintf("COM%d", i)
		lpDeviceName := (*C.CHAR)(C.CString(name))
		n := C.QueryDosDevice(lpDeviceName, lpTargetPath, C.DWORD(len(buffer)))
		if n > 0 {
			results[name] = name
		}
	}

	return results
}
