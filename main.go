package main

import (
	"log"
	"net/http"
	"time"
)



func main() {
	serialHandler = &SerialHandler{baudRate: baudRate}
	s := &http.Server{
		Addr:           ":8080",
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	myHandler = mmuxHandler()

	log.Fatal(s.ListenAndServe())
}
