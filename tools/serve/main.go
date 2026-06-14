package main

import (
	"log"
	"net/http"
)

func main() {
	port := ":5500"
	// Point to the base public directory so sub-folders resolve perfectly
	dir := "./backend/public"
	log.Printf("Serving static assets on http://localhost%s/admin/login.html", port)
	log.Fatal(http.ListenAndServe(port, http.FileServer(http.Dir(dir))))
}
