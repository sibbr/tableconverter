package main

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
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
	"sync"
	"time"
)

// Labels placeholder for labels
type Labels struct {
	Value string
	ID    int
}

var templates = template.Must(template.ParseFiles("labels.html"))
var cookieDuration = 60 * 60 // cookie active time in seconds

// Publicador hold information about the publisher converting tables at moment
type Publicador struct {
	fp      multipart.File
	form    *multipart.Form
	sep     string
	cookie  http.Cookie
	created time.Time
}

// Publicadores list of publishers
var Publicadores = map[string]Publicador{}
var mutex = &sync.RWMutex{}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		publicador := Publicador{}
		var err error

		publicador.fp, _, err = r.FormFile("uploadFile")
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		publicador.sep = r.FormValue("separator")
		labels := getLabels(publicador.fp, r.FormValue("separator"))
		nlabels := make([]Labels, len(labels))
		for k, v := range labels {
			nlabels[k].Value = v
			nlabels[k].ID = k
		}

		// cookie creation
		rawCookie := make([]byte, 12)
		rand.Read(rawCookie)
		cookieName := hex.EncodeToString(rawCookie)
		publicador.cookie = http.Cookie{Name: "sibbr-tableconverter", Value: cookieName, MaxAge: cookieDuration, HttpOnly: true}
		http.SetCookie(w, &publicador.cookie)

		publicador.created = time.Now()
		publicador.form = r.MultipartForm
		mutex.Lock()
		Publicadores[cookieName] = publicador
		mutex.Unlock()

		renderTemplate(w, "labels", &nlabels)

	} else if r.Method == "GET" {

		cookie, err := r.Cookie("sibbr-tableconverter")
		if err != nil {
			http.Redirect(w, r, "http://"+r.Host, 302)
			return
		}

		var publicador Publicador

		// check if cookie is alive on server side
		mutex.RLock()
		if _, ok := Publicadores[cookie.Value]; ok {
			publicador = Publicadores[cookie.Value]
		} else {
			http.Redirect(w, r, "http://"+r.Host, 302)
			return
		}
		mutex.RUnlock()

		// Parse form values
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		// Inverting form values (maps golang)
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

		// Seeking fp cause of getLabels used early
		_, err = publicador.fp.Seek(0, 0)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		// First run of Melt looking for errors
		// FIXME: buffer output to escape second Melt call
		err = Melt(publicador.fp, ioutil.Discard, ordenados, publicador.sep)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		// Rewind because of the first Melt call
		_, err = publicador.fp.Seek(0, 0)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		// set cookie to delete, processing done
		publicador.cookie.MaxAge = -1
		http.SetCookie(w, &publicador.cookie)

		w.Header().Set("Content-Disposition", "attachment; filename=converted.csv")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

		Melt(publicador.fp, w, ordenados, publicador.sep)

		// delete publisher
		mutex.Lock()
		Publicadores[cookie.Value].fp.Close()
		Publicadores[cookie.Value].form.RemoveAll()
		delete(Publicadores, cookie.Value)
		mutex.Unlock()

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
		fmt.Fprint(w, "Page not found")
	}
	io.Copy(w, strings.NewReader(string(content)))
}

func main() {
	// goroutine to remove expired info (invalid cookie, files) from server
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()
			for k, v := range Publicadores {
				if v.created.Add(time.Duration(cookieDuration) * time.Second).After(time.Now()) {
					Publicadores[k].fp.Close()
					Publicadores[k].form.RemoveAll()
					delete(Publicadores, k)
				}
			}
			mutex.Unlock()
		}
	}()

	fsCSS := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", fsCSS))

	fsIMG := http.FileServer(http.Dir("img"))
	http.Handle("/img/", http.StripPrefix("/img/", fsIMG))

	fsJS := http.FileServer(http.Dir("js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsJS))

	http.HandleFunc("/upload", upload)
	http.HandleFunc("/", home)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
