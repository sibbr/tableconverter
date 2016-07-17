package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type appHandler struct {
	*Context
	H func(*Context, http.ResponseWriter, *http.Request) error
}

func (ch appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := ch.H(ch.Context, w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func recvFile(c *Context, w http.ResponseWriter, r *http.Request) error {

	fr, _, err := r.FormFile(c.File)
	if err != nil {
		return err
	}

	invertedForm := make(map[string]string)
	for key, value := range r.MultipartForm.Value {
		invertedForm[value[0]] = key
	}

	tmpFile, err := ioutil.TempFile(".", "tableconverter")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, fr)
	if err != nil {
		return err
	}

	return nil
}
