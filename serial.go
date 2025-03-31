package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

const (
	baudRate = 115200 // Puedes ajustar esto si es necesario
	timeout  = 5 * time.Second
)

type ESP32 interface {
	Init() error
	SendCmd(cmd string) error
	GetBufferHistorial() []string
}

type esp32 struct {
	port               serial.Port
	portName           string
	mu                 sync.Mutex
	magBufferHistorial []string
	stopReadSerial     chan bool
	isReading          bool
}

func NewESP32() ESP32 {
	return &esp32{
		stopReadSerial: make(chan bool),
	}
}

func (e *esp32) Init() (err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Si ya hay una conexión, intenta cerrarla
	if e.port != nil {
		if err := e.port.Close(); err != nil {
			log.Printf("Error al cerrar el puerto anterior: %v", err)
			return err
		}
		e.port = nil
	}
	e.portName, err = findSerialPort()
	if err != nil {
		log.Printf("Port Not Found : %v\n", err)
	}

	// Configurar y abrir el puerto serial
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	e.port, err = serial.Open(e.portName, mode)
	if err != nil {
		log.Printf("Error al abrir el puerto serial: %v", err)
		return err
	}

	// Iniciar la lectura del serial en una goroutine si no está ya en curso
	if !e.isReading {
		go e.readSerialToBuffer()
		e.isReading = true
	}

	log.Println("Conexión serial inicializada en:", e.portName)
	return nil
}

func (e *esp32) SendCmd(cmd string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.port == nil {
		return fmt.Errorf("puerto serial no inicializado")
	}

	// Agregar '\n' al comando si no lo tiene
	if !strings.HasSuffix(cmd, "\n") {
		cmd += "\n"
	}

	_, err := e.port.Write([]byte(cmd))
	if err != nil {
		log.Printf("Error al escribir en el puerto serial: %v", err)
		return err
	}

	log.Printf("Comando enviado: %s", cmd)
	return nil
}

func (e *esp32) readSerialToBuffer() {
	reader := bufio.NewReader(e.port)
	for {
		select {
		case <-e.stopReadSerial:
			log.Println("Deteniendo la lectura del serial...")
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Println("Fin de la conexión serial.")
					return
				}
				log.Printf("Error al leer del serial: %v", err)
				time.Sleep(time.Second) // Esperar antes de intentar leer de nuevo
				continue
			}

			e.mu.Lock()
			e.magBufferHistorial = append(e.magBufferHistorial, strings.TrimSpace(line))
			e.mu.Unlock()

			log.Printf("Recibido del serial: %s", strings.TrimSpace(line))
		}
	}
}

func (e *esp32) GetBufferHistorial() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.magBufferHistorial
}

// StopReading detiene la goroutine de lectura del serial.
func (e *esp32) StopReading() {
	if e.isReading {
		e.stopReadSerial <- true
		e.isReading = false
	}
}

func findSerialPort() (string, error) {
	for i := 0; i < 10; i++ {
		portName := fmt.Sprintf("/dev/ttyUSB%d", i)
		mode := &serial.Mode{
			BaudRate: baudRate,
		}
		port, err := serial.Open(portName, mode)
		if err == nil {
			// Close the port immediately after opening to check if it's available
			port.Close()
			log.Println("Puerto serial encontrado:", portName)
			return portName, nil
		}
	}
	return "", fmt.Errorf("no se encontró ningún puerto serial disponible en /dev/ttyUSB[0-9]")
}