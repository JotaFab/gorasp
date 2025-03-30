package main

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/tarm/serial"
)

type SerialHandler struct {
	port     *serial.Port
	portName string
	baudRate int
	mutex    sync.Mutex
}

func (sh *SerialHandler) Init(portName string, baudRate int) error {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	config := &serial.Config{Name: portName, Baud: baudRate}
	var err error
	sh.port, err = serial.OpenPort(config)
	if err != nil {
		return err
	}

	sh.portName = portName
	sh.baudRate = baudRate
	return nil
}

func (sh *SerialHandler) SendCommand(command string) error {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if sh.port == nil {
		return fmt.Errorf("serial port not open")
	}

	_, err := sh.port.Write([]byte(command + "\n"))
	if err != nil {
		return err
	}

	fmt.Println("Sent:", command)
	return nil
}

func (sh *SerialHandler) Close() error {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if sh.port != nil {
		err := sh.port.Close()
		if err != nil {
			return err
		}
		sh.port = nil
	}
	return nil
}

func (sh *SerialHandler) ReadSerial(messageBuffer *[]string, bufferMutex *sync.Mutex) {
	buf := make([]byte, 128)
	var tempBuf []byte // Temporary buffer to accumulate data
	for {
		n, err := sh.port.Read(buf)
		if err != nil {
			log.Println("Error reading from serial:", err)
			continue
		}
		if n > 0 {
			tempBuf = append(tempBuf, buf[:n]...) // Append new data to temp buffer
			for {
				i := bytes.IndexByte(tempBuf, '\n') // Check for newline
				if i < 0 {
					break // No newline found, continue reading
				}
				lineBytes := tempBuf[:i] // Extract the complete line as bytes
				msg := string(lineBytes) // Convert bytes to string (handles UTF-8)
				tempBuf = tempBuf[i+1:]  // Remove the processed line

				// Clean up non-ASCII characters
				cleanedMsg := ""
				for _, r := range msg {
					if r >= ' ' && r <= '~' { // Basic ASCII range
						cleanedMsg += string(r)
					}
				}

				log.Println("Received:", cleanedMsg)
				bufferMutex.Lock()
				*messageBuffer = append(*messageBuffer, cleanedMsg) // Append to message buffer
				bufferMutex.Unlock()
			}
		}
	}
}
