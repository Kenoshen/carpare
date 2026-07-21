// Command carpare serves the Carpare car-comparison dashboard.
package main

import (
	"log"
	"net/http"

	"carpare/internal/db"
	"carpare/internal/handlers"
	"carpare/internal/templates"
)

const addr = "localhost:6767"

func main() {
	pages, err := templates.Load()
	if err != nil {
		log.Fatalf("loading templates: %v", err)
	}

	store, err := db.Open("data")
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}

	h := handlers.New(pages, store)

	mux := http.NewServeMux()
	mux.Handle("/", h.Routes())
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Printf("carpare listening on http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
