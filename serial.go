package main

import (
	"fmt"
	"go.bug.st/serial"
	"log"
)

func main() {
	// Abrir puerto serial (ajusta según el ESP32)
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyUSB0", mode)
	if err != nil {
		log.Fatal("Error abriendo el puerto:", err)
	}
	defer port.Close()

	// Enviar mensaje al ESP32
	msg := "Hola ESP32\n"
	n, err := port.Write([]byte(msg))
	if err != nil {
		log.Fatal("Error enviando datos:", err)
	}
	fmt.Printf("Mensaje enviado: %s (%d bytes)\n", msg, n)
}
