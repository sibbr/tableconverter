package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	context := &Context{":8080", "file"}
	router.Handle("/file", appHandler{context, recvFile})
	log.Fatal(http.ListenAndServe(context.Bind, router))
}
