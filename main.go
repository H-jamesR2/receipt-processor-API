package main

import (
	"fmt"
	"log"
	"net/http"
	"rcpt-proc-challenge-ans/controller"
)

func main() {
	http.HandleFunc("/receipts/process", controller.ProcessReceipt)
	http.HandleFunc("/receipts/", controller.GetReceipt)

	// Handle all other routes
    http.HandleFunc("/", controller.NotFoundHandler)

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
