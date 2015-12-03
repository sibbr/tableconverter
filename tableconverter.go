package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type reshapeError struct {
	prob string
}

func (c *reshapeError) Error() string {
	return fmt.Sprintf("%s", c.prob)
}

// Labels placeholder for labels
type Labels struct {
	Value string
	ID    int
}

var templates = template.Must(template.ParseFiles("labels.html"))
var fp multipart.File
var sep string

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseMultipartForm(1000000 * 500)
		if err != nil { // 500mb
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		fp, _, err = r.FormFile("uploadFile")
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		defer fp.Close()

		sep = r.FormValue("separator")
		labels := getLabels(fp, r.FormValue("separator"))
		nlabels := make([]Labels, len(labels))
		for k, v := range labels {
			nlabels[k].Value = v
			nlabels[k].ID = k
		}

		renderTemplate(w, "labels", &nlabels)

	} else if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		fixed := []int{}
		reverseForm := map[int]string{}

		for k, v := range r.Form {
			numero, _ := strconv.Atoi(v[0])
			fixed = append(fixed, numero)
			reverseForm[numero] = k
		}

		sort.Ints(fixed)
		ordenados := []string{}

		for i := 0; i < len(fixed); i++ {
			ordenados = append(ordenados, reverseForm[fixed[i]])
		}

		_, err := fp.Seek(0, 0)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
		}

		err = Melt(fp, ioutil.Discard, ordenados, sep)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
		}

		_, err = fp.Seek(0, 0)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
		}

		w.Header().Set("Content-Disposition", "attachment; filename=converted.csv")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		Melt(fp, w, ordenados, sep)

	}

}

func renderTemplate(w http.ResponseWriter, tmpl string, p *[]Labels) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getLabels(input io.Reader, sep string) []string {
	dados := csv.NewReader(input)
	if sep == "tab" {
		dados.Comma = '\t'
	} else {
		dados.Comma = rune(sep[0])
	}
	dados.FieldsPerRecord = -1

	labels, err := dados.Read()
	if err != nil {
		return nil
	}

	return labels
}

func home(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("index.html")
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprint(w, "Not found")
	}
	io.Copy(w, strings.NewReader(string(content)))
}

func main() {

	fsCSS := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", fsCSS))

	fsIMG := http.FileServer(http.Dir("img"))
	http.Handle("/img/", http.StripPrefix("/img/", fsIMG))

	http.HandleFunc("/upload", upload)
	http.HandleFunc("/", home)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
