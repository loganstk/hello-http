package main

import (
	"net/http"

	"com.github/loganstk/hello-http/handler"
)

func main() {

	http.HandleFunc("POST /vendor/{vendorId}/point", handler.PostHandler)

	http.ListenAndServe(":8080", nil)
}
