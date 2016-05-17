package main

import (
	"io"
	"net/http"
)

func hello(w http.ResponseWriter, h *http.Request) {
	io.WriteString(w, "Hello world!")
}

func main() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8000", nil)
}
