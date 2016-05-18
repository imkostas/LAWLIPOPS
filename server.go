package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

var templates = template.Must(template.ParseFiles("templates/index.html",
	"templates/upload.html",
	"templates/challenges.html"))

type page struct {
	Title string
	Body  string
}

func display(w http.ResponseWriter, name string, data interface{}) {
	templates.ExecuteTemplate(w, name+".html", data)

	// t, _ := template.ParseFiles("challenges.html")
	// t.Execute(w, p)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	display(w, "index", nil)
}

func challengesHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/challenges/"):]
	body := "Nothing to see here"
	p := page{Title: title, Body: body}
	display(w, "challenges", p)
}

// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("method: ", r.Method)
// 	if r.Method == "GET" {
// 		display(w, "upload", nil)
// 	} else {
// 		r.ParseForm()
//
// 		fmt.Println("file: ", r.Form["file"]) // file is from the name of the input
//
// 	}
// }

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "upload", "Choose files")

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		// // file, header, err := r.FormFile("file")

		//parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		files := m.File["myfiles"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			dst, err := os.Create("files/" + files[i].Filename)
			if err != nil {
				fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
				return
			}
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		//display success message.
		display(w, "upload", "Upload successful!")
		// http.Redirect(w, r, "/upload/", http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/challenges/", challengesHandler)
	http.HandleFunc("/upload/", uploadHandler)
	// http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
