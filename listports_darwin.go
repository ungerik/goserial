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
	// "path/filepath"
	"fmt"
	"strings"
	"unsafe"
)

var (
	kIOSerialBSDTypeKey   = C.CFStringCreateWithCStringNoCopy(nil, C.CString(C.kIOSerialBSDTypeKey), C.kCFStringEncodingASCII, nil)
	kIOSerialBSDRS232Type = C.CFStringCreateWithCStringNoCopy(nil, C.CString(C.kIOSerialBSDRS232Type), C.kCFStringEncodingASCII, nil)
	kIOCalloutDeviceKey   = C.CFStringCreateWithCStringNoCopy(nil, C.CString(C.kIOCalloutDeviceKey), C.kCFStringEncodingASCII, nil)
)

func listPorts() map[string]string {
	results := make(map[string]string)

	var masterPort C.mach_port_t
	kernResult := C.IOMasterPort(C.MACH_PORT_NULL, (*C.mach_port_t)(unsafe.Pointer(&masterPort)))
	if kernResult != C.KERN_SUCCESS {
		panic(fmt.Errorf("IOMasterPort error %v", kernResult))
	}

	classesToMatch := C.IOServiceMatching(C.CString(C.kIOSerialBSDServiceValue))
	if classesToMatch == nil {
		panic(fmt.Errorf("IOServiceMatching returned nil"))
	}

	C.CFDictionarySetValue(
		classesToMatch,
		unsafe.Pointer(kIOSerialBSDTypeKey),
		unsafe.Pointer(kIOSerialBSDRS232Type))

	var matchingServices C.io_iterator_t
	kernResult = C.IOServiceGetMatchingServices(masterPort, classesToMatch, &matchingServices)
	if kernResult != C.KERN_SUCCESS {
		panic(fmt.Errorf("IOServiceGetMatchingServices error %v", kernResult))
	}

	for serialService := C.IOIteratorNext(matchingServices); serialService != 0; serialService = C.IOIteratorNext(matchingServices) {

		deviceFilePathAsCFString := C.IORegistryEntryCreateCFProperty(
			C.io_registry_entry_t(serialService),
			kIOCalloutDeviceKey,
			C.kCFAllocatorDefault,
			0)

		if deviceFilePathAsCFString != nil {

			var deviceFilePath [1024]C.char

			result := C.CFStringGetCString(
				(*C.struct___CFString)(deviceFilePathAsCFString),
				&deviceFilePath[0],
				C.CFIndex(len(deviceFilePath)),
				C.kCFStringEncodingASCII)

			C.CFRelease(deviceFilePathAsCFString)

			if result != 0 {
				path := C.GoString(&deviceFilePath[0])
				name := strings.TrimPrefix(path, "/dev/cu.")
				results[name] = path
			}
		}

		C.IOObjectRelease(serialService)
	}

	return results
}

// func listPorts() map[string]string {
// 	matches, err := filepath.Glob("/dev/cu.*")
// 	if err != nil {
// 		panic(err)
// 	}
// 	results := make(map[string]string)
// 	for _, path := range matches {
// 		name := strings.TrimPrefix(path, "/dev/cu.")
// 		results[name] = path
// 	}
// 	return results
// }
