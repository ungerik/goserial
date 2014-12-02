// +build windows

package goserial

import "C"

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
	return conn.file.Close()
}
