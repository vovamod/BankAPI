package main

import (
	"fmt"
	"net/http"
)

func main() {
	// set-up server + handler
	server := &http.Server{
		Addr:    ":25565",
		Handler: http.HandlerFunc(basicHandler),
	}
	// Robust log of failure (server like)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to start server. Reason: ", err)
	}
}

// Routes handler + Requests here
func basicHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Received call at: " + r.URL.Path))
	if err != nil {
		return
	}
}
