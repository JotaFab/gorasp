package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
)

var (
	portName      string // Adjust based on your setup
	baudRate      = 115200
	messageBuffer []string
	bufferMutex   sync.Mutex
	tmpl          = template.Must(template.ParseFiles("index.html"))
	myHandler     *http.ServeMux
	// Initialize serialHandler and myHandler
	serialHandler = &SerialHandler{baudRate: baudRate}
)

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

	if serialHandler != nil && serialHandler.port != nil {
		err := serialHandler.SendCommand(data.Command)
		if err != nil {
			http.Error(w, "Failed to send command", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "sent", "command": data.Command})
	} else {
		http.Error(w, "Serial port not open", http.StatusInternalServerError)
	}
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
		log.Println("WebSocket upgrade error:", err) // More specific error message
		return
	}
	defer conn.Close()

	// Send the entire messageBuffer to the new client
	bufferMutex.Lock()
	for _, msg := range messageBuffer {
		err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("WebSocket write error:", err)
			bufferMutex.Unlock()
			return
		}
	}
	bufferMutex.Unlock()

	for {
		bufferMutex.Lock()
		if len(messageBuffer) > 0 {
			// Send only the new messages
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

func findSerialPort() (string, error) {
	var candidates []string

	if runtime.GOOS == "linux" {
		candidates = []string{"/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyACM0", "/dev/ttyACM1"}
	} else {
		return "", fmt.Errorf("unsupported operating system")
	}

	for _, port := range candidates {
		config := &serial.Config{Name: port, Baud: baudRate}
		s, err := serial.OpenPort(config)
		if err == nil {
			s.Close()
			fmt.Println("Found serial port:", port)
			return port, nil
		}
		fmt.Println("Attempted", port, "Error:", err)
	}

	return "", fmt.Errorf("no serial port found")
}

func connectSerialHandler(w http.ResponseWriter, r *http.Request) {
	foundPort, err := findSerialPort()
	if err != nil {
		log.Println("Error finding serial port:", err)
		http.Error(w, "Failed to find serial port", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		return
	}

	err = serialHandler.Init(foundPort, baudRate)
	if err != nil {
		log.Println("Error initializing serial port:", err)
		http.Error(w, "Failed to initialize serial port", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		return
	}

	portName = foundPort // Update global portName
	fmt.Println("Serial port connected on", portName)
	go serialHandler.ReadSerial(&messageBuffer, &bufferMutex)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "port": portName})
	return
}

func scanAPsHandler(w http.ResponseWriter, r *http.Request) {
	
	ssids, err := AutomateAPScan(serialHandler, &messageBuffer, &bufferMutex)
	if err != nil {
		log.Printf("Error during AP scan automation: %v", err)                                    // Detailed log
		http.Error(w, fmt.Sprintf("Failed to scan APs: %v", err), http.StatusInternalServerError) // Include error in response
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "aps": ssids})
}

func mmuxHandler() *http.ServeMux {

	myHandler = http.NewServeMux()
	myHandler.HandleFunc("/", servePage)
	myHandler.HandleFunc("/send", sendCommand)
	myHandler.HandleFunc("/ws", serveWs)
	myHandler.HandleFunc("/connect_serial", connectSerialHandler)
	myHandler.HandleFunc("/scan_aps", scanAPsHandler) // Register the new handler

	return myHandler
}
