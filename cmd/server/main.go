package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"
)

//go:embed fonts/*
var fonts embed.FS

//go:embed images/*
var img embed.FS

func main() {
	http.Handle("/", Logger(HandlerRoot()))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusAndSizeWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		fmt.Printf(
			"%d %s %s %s %s\n",
			sw.Status,
			r.Method,
			r.RequestURI,
			time.Since(start),
			humanizeBytes(int64(sw.Size)),
		)
	})
}

func humanizeBytes(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	unit := 0
	for size >= 1024 {
		size /= 1024
		unit++
	}
	return fmt.Sprintf("%d %s", size, units[unit])
}
