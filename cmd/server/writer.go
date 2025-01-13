package main

import "net/http"

type statusAndSizeWriter struct {
	http.ResponseWriter
	Status int
	Size   int
}

func (w *statusAndSizeWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusAndSizeWriter) Write(b []byte) (int, error) {
	written, err := w.ResponseWriter.Write(b)
	if err != nil {
		return written, err
	}
	w.Size += written
	return written, nil
}
