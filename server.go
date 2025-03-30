package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
)

var (
	portName      = "/dev/ttyUSB1" // Adjust based on your setup
	baudRate      = 115200
	serialPort    *serial.Port
	messageBuffer []string
	bufferMutex   sync.Mutex
	tmpl          = template.Must(template.ParseFiles("index.html"))
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

// Read from the serial port
func readSerial() {
	buf := make([]byte, 128)
	var tempBuf []byte // Temporary buffer to accumulate data
	for {
		n, err := serialPort.Read(buf)
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
				msg := string(tempBuf[:i]) // Extract the complete line
				tempBuf = tempBuf[i+1:]    // Remove the processed line
				log.Println("Received:", msg)
				bufferMutex.Lock()
				messageBuffer = append(messageBuffer, msg) // Append to message buffer
				bufferMutex.Unlock()
			}
		}
	}
}

// Serve the HTML page
func servePage(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

// Send command to ESP32
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

// Retrieve received messages
func receiveMessages(w http.ResponseWriter, r *http.Request) {
	bufferMutex.Lock()
	messages := messageBuffer
	messageBuffer = []string{}
	bufferMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"messages": messages})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Serve WebSocket endpoint
func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		bufferMutex.Lock()
		if len(messageBuffer) > 0 {
			for _, msg := range messageBuffer {
				err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("WebSocket write error:", err)
					bufferMutex.Unlock()
					return
				}
			}
			messageBuffer = []string{} // Clear the buffer after sending messages
		}
		bufferMutex.Unlock()
		time.Sleep(100 * time.Millisecond) // Check every 100ms
	}
}

func main() {
	if err := initSerial(); err != nil {
		log.Fatal("Failed to open serial port:", err)
	}

	http.HandleFunc("/", servePage)
	http.HandleFunc("/send", sendCommand)
	http.HandleFunc("/receive", receiveMessages)
	http.HandleFunc("/ws", serveWs) // Register WebSocket endpoint

	fmt.Println("Server started on port 5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
