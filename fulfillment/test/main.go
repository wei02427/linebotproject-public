package main

import (
	"log"
	"net/http"

	"project.com/fulfillment"
)

func main() {
	http.HandleFunc("/", fulfillment.Fulfillment)

	log.Fatal(http.ListenAndServe(":3000", nil))

}
