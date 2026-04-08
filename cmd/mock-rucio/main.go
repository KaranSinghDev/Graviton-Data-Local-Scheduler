package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/KaranSinghDev/data-gravity-operator/internal/mockrucio"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on")
	flag.Parse()

	handler := mockrucio.NewHandler()
	log.Printf("mock-rucio listening on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
