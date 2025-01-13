package main

import (
	"embed"
	"log"
	"net/http"
)

//go:embed fonts/*
var fonts embed.FS

//go:embed images/*
var img embed.FS

func main() {
	http.Handle("/", HandlerRoot())

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
