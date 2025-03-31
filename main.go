package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/tarm/serial"
)

const (
	serialPort = "/dev/ttyUSB0"
	baudRate   = 115200 // Ajusta según tu dispositivo
)

var (
	ser *serial.Port
	mu  sync.Mutex
	outputBuffer []string
)

func initSerial() error {
	c := &serial.Config{Name: serialPort, Baud: baudRate}
	var err error
	ser, err = serial.OpenPort(c)
	if err != nil {
		return fmt.Errorf("error al abrir el puerto serial %s: %w", serialPort, err)
	}
	fmt.Printf("Conectado al puerto serial: %s\n", serialPort)
	go readSerial()
	return nil
}

func readSerial() {
	reader := bufio.NewReader(ser)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error al leer del puerto serial: %v\n", err)
			return
		}
		mu.Lock()
		outputBuffer = append(outputBuffer, line)
		mu.Unlock()
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al cargar la plantilla: %v", err), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al ejecutar la plantilla: %v", err), http.StatusInternalServerError)
	}
}

func sendCommandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	command := r.FormValue("command") + "\n" // Añade un salto de línea
	mu.Lock()
	if ser != nil {
		_, err := ser.Write([]byte(command))
		mu.Unlock()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error al escribir en el puerto serial: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Comando enviado: %s", command)
		w.WriteHeader(http.StatusOK)
		return
	}
	mu.Unlock()
	http.Error(w, "No hay conexión serial activa", http.StatusInternalServerError)
}

func getOutputHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	output := make([]string, len(outputBuffer))
	copy(output, outputBuffer)
	outputBuffer = nil // Limpia el buffer
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string][]string{"output": output})
	if err != nil {
		log.Printf("Error al codificar la salida JSON: %v", err)
	}
}

func main() {
	err := initSerial()
	if err != nil {
		log.Fatalf("Error al inicializar el serial: %v", err)
		os.Exit(1)
	}
	defer ser.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/send_command", sendCommandHandler)
	http.HandleFunc("/get_output", getOutputHandler)

	fmt.Println("Servidor web escuchando en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}