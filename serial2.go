package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	// Configuración del puerto serie
	config := &serial.Config{
		Name:    "/dev/ttyUSB0", // Cambia si tu ESP32 está en otro puerto
		Baud:    115200,         // Baudrate detectado
		Timeout: time.Second,    // Timeout de 1 segundo
	}

	// Abre el puerto serie
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	// Escribir datos al ESP32
	_, err = port.Write([]byte("Hola ESP32!\n"))
	if err != nil {
		log.Fatal(err)
	}

	// Leer respuesta del ESP32
	buf := make([]byte, 128)
	n, err := port.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ESP32 dice: %s\n", buf[:n])
}
