package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	tmpl        = template.Must(template.ParseFiles("public/index.html"))
	myHandler   *http.ServeMux
	serialMutex sync.Mutex
	e           ESP32                            // Instancia de ESP32
	clients     = make(map[*websocket.Conn]bool) // Clientes conectados
	clientsMu   sync.Mutex                       // Mutex para proteger el acceso a clients
)

func muxHandler() *http.ServeMux {

	myHandler = http.NewServeMux()
	myHandler.HandleFunc("/", servePage)
	myHandler.HandleFunc("/send", sendCommand)
	myHandler.HandleFunc("/ws", serveWs)
	myHandler.HandleFunc("/connect", handleConnect)
	// myHandler.HandleFunc("/scan_aps", scanAPsHandler) // Register the new handler

	return myHandler
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

	serialMutex.Lock()
	defer serialMutex.Unlock()

	if e != nil {
		err := e.SendCmd(data.Command)
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
	CheckOrigin: func(r *http.Request) bool {
		return true // Permitir todas las conexiones (solo para desarrollo)
	},
}

// Serve WebSocket endpoint
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	log.Println("Cliente conectado")

	// Enviar el historial inicial al cliente
	initialHistory := e.GetBufferHistorial()
	for _, msg := range initialHistory {
		if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Error al enviar el historial inicial:", err)
			break
		}
	}

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			log.Println("Cliente desconectado:", err)
			return
		}
		// Manejar los mensajes del cliente (comandos para el ESP32)
		if messageType == websocket.TextMessage {
			command := string(p)
			log.Printf("Recibido comando del cliente: %s", command)
			err := e.SendCmd(command)
			if err != nil {
				log.Printf("Error al enviar el comando al ESP32: %v", err)
				sendMessageToClient(ws, "Error al enviar el comando al ESP32: "+err.Error())
			} else {
				sendMessageToClient(ws, "Comando enviado: "+command)
			}
		}
	}
}

func broadcastMessage(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error al enviar mensaje al cliente: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func sendMessageToClient(client *websocket.Conn, message string) {
	err := client.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error al enviar mensaje al cliente: %v", err)
		client.Close()
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
	}
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	serialMutex.Lock()
	defer serialMutex.Unlock()

	if e == nil {
		e = NewESP32()
	}

	err := e.Init()
	if err != nil {
		log.Println("Error initializing serial port:", err)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Serial connected successfully"})
}
