package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	esp ESP32
	web http.Server
}

var server *Server

func StartServer() (s *Server) {
	s = &Server{}
	s.esp = NewESP32()

	s.web = http.Server{
		Addr:    ":5000",
		Handler: muxHandler(),
	}

	err := s.esp.Init()
	if err != nil {
		log.Fatalf("Error al inicializar la conexión serial: %v", err)
	}
	defer func() {
		if esp, ok := e.(*esp32); ok {
			esp.StopReading() // Detener la goroutine de lectura al finalizar
			if esp.port != nil {
				esp.port.Close() // Cerrar el puerto serial
			}
		}
	}()

	go func() {
		// Bucle para enviar el historial a los clientes conectados
		for {
			time.Sleep(1 * time.Second) // Ajusta el intervalo según sea necesario
			historial := s.esp.GetBufferHistorial()
			if len(historial) > 0 {
				lastMessage := historial[len(historial)-1]
				broadcastMessage(lastMessage)
			}
		}
	}()

	fmt.Printf("Servidor WebSocket escuchando en %s\n", s.web.Addr)

	log.Fatal(s.web.ListenAndServe())
	return s
}

func main() {
	// Start the HTTP server
	server = StartServer()
}
