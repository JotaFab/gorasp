package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/tarm/serial"
)

var (
	portName      = "/dev/ttyUSB0" // Change this to match your ESP32 serial port
	baudRate      = 115200
	serialPort    *serial.Port
	messageBuffer []string
	bufferMutex   sync.Mutex
)

// Initialize the serial connection
func initSerial() error {
	config := &serial.Config{Name: portName, Baud: baudRate}
	var err error
	serialPort, err = serial.OpenPort(config)
	if err != nil {
		return err
	}

	// Start reading from serial in a goroutine
	go readSerial()
	return nil
}

// Reads data from the serial port and stores it in a buffer
func readSerial() {
	buf := make([]byte, 128)
	for {
		n, err := serialPort.Read(buf)
		if err != nil {
			log.Println("Error reading from serial:", err)
			continue
		}
		if n > 0 {
			msg := string(buf[:n])
			log.Println("Received:", msg)
			bufferMutex.Lock()
			messageBuffer = append(messageBuffer, msg)
			bufferMutex.Unlock()
		}
	}
}

// Send a command to the ESP32
func sendCommand(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if serialPort != nil {
		_, err := serialPort.Write([]byte(data.Command + "\n"))
		if err != nil {
			http.Error(w, "Failed to send command", http.StatusInternalServerError)
			return
		}
		fmt.Println("Sent:", data.Command)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "sent", "command": data.Command})
	} else {
		http.Error(w, "Serial port not open", http.StatusInternalServerError)
	}
}

// Fetch received messages
func receiveMessages(w http.ResponseWriter, r *http.Request) {
	bufferMutex.Lock()
	messages := messageBuffer
	messageBuffer = []string{}
	bufferMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"messages": messages})
}

func main() {
	if err := initSerial(); err != nil {
		log.Fatal("Failed to open serial port:", err)
	}

	http.HandleFunc("/send", sendCommand)
	http.HandleFunc("/receive", receiveMessages)

	fmt.Println("Server started on port 5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
