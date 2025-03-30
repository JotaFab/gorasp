package main

import (
	"log"
	"net/http"
	"time"
)

func main() {

	myHandler = mmuxHandler()

	// Configure the server
	s := &http.Server{
		Addr:           ":5000",
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Serving on: ", s.Addr)
	log.Fatal(s.ListenAndServe())
}

