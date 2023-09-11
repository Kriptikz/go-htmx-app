package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = "10000"

func pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello World!")
	case "/ping":
		pong(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	log.Println("Starting server on port " + port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}
