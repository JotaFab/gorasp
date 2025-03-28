package main

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	// Configurar la conexión serial con el ESP32
	config := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	// Esperar a que el ESP32 esté listo
	time.Sleep(2 * time.Second)

	// Enviar comando AT para escanear dispositivos Bluetooth
	command := "AT+BLESCAN=1\r\n"
	_, err = port.Write([]byte(command))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Escaneando dispositivos Bluetooth...")

	// Leer la respuesta del ESP32
	reader := bufio.NewReader(port)
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(response)

		// Si el ESP32 deja de enviar datos, salir
		if response == "OK\r\n" {
			break
		}
	}
}
