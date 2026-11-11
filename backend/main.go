package main

import (
	"log"
	"net/http"
	"time"

	"webmail/backend/internal/config"
	"webmail/backend/internal/server"
)

func main() {
	cfg := config.Load()
	srv := server.New(cfg)
	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("webmail: listening on %s", cfg.HTTPAddr)
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
