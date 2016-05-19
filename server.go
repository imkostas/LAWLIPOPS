package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
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

		// Open Database connection
		// db, err := sql.Open("mysql", "root:root@/tcp(52.20.186.36:3306)/test?tls=skip-verify$autocommit=true")
		db, err := sql.Open("mysql", "root:root@/test")
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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

			// Insert into DB
			stmtIns, err := db.Prepare("INSERT INTO `data` (`id`, `doc`, `flag`) VALUES (NULL, ?, ?);")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer stmtIns.Close()

			_, err = stmtIns.Exec("files/"+files[i].Filename, 0)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		//display success message.
		display(w, "upload", "Upload successful!")
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
