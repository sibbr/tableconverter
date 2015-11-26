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

// Melt will change format to wide -> long
func Melt(input io.Reader, output io.Writer, fixed []string, sep string) error {

	dados := csv.NewReader(input)
	if sep == "tab" {
		dados.Comma = '\t'
	} else {
		dados.Comma = rune(sep[0])
	}
	dados.FieldsPerRecord = -1
	dados.LazyQuotes = true

	labels, err := dados.Read()

	if err != nil {
		return err
	}
	// data cleaning, removing all leading and trailing white space
	for k, v := range labels {
		labels[k] = strings.TrimSpace(v)
	}
	for k, v := range fixed {
		fixed[k] = strings.TrimSpace(v)
	}

	// stop if duplicate labels are found
	found := map[string]int{}
	anyDuplicate := []string{}
	for _, v := range labels {
		if _, ok := found[v]; ok {
			anyDuplicate = append(anyDuplicate, v)
		}
		found[v]++
	}
	if len(anyDuplicate) > 0 {
		return &reshapeError{"Duplicated column names: " + strings.Join(anyDuplicate, ", ")}
	}

	writeMeasurementData := csv.NewWriter(output)
	outputLabels := []string{"eventid"}
	outputLabels = append(outputLabels, fixed...)
	outputLabels = append(outputLabels, "measurementType", "measurementValue")
	if writeMeasurementData.Write(outputLabels) != nil {
		return err
	}

	fixedPos := []int{}
	for k, v := range labels {
		if indexContains(v, &fixed) > -1 {
			fixedPos = append(fixedPos, k)
		}
	}
	if len(fixedPos) < len(fixed) {
		return &reshapeError{"Fixed column not found in dataset"}
	}

	// for each line do a rotation and write, no waste in memory
	// number of resulting lines = (ncol - fixed) * nrow
	// fixed are columns like eventid to control the rotation of data
	for eventid := 1; ; eventid++ {

		line, err := dados.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		for elem := 0; elem < len(line); elem++ {
			if contains(labels[elem], &fixed) {
				continue
			}
			outputLine := []string{strconv.Itoa(eventid)}
			for _, v := range fixedPos {
				outputLine = append(outputLine, line[v])
			}
			outputLine = append(outputLine, labels[elem], line[elem])
			if writeMeasurementData.Write(outputLine) != nil {
				return err
			}
		}
	}
	writeMeasurementData.Flush()

	return nil
}

func contains(element string, elements *[]string) bool {
	for _, v := range *elements {
		if element == v {
			return true
		}
	}
	return false
}

func indexContains(element string, elements *[]string) int {
	for k, v := range *elements {
		if element == v {
			return k
		}
	}
	return -1
}

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
