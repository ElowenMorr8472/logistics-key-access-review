package main

import (
	"log"
	"net/http"
	"os"

	"github.com/infrai-examples/logistics-key-audit/internal/logistics"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.Handle("POST /reviews", logistics.NewReviewHandler(logistics.NewClient(key)))
	log.Printf("logistics access review listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
