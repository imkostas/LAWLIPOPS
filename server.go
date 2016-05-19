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

type File struct {
	ID   int
	Path string
	Flag int
}

type page struct {
	Title   string
	Body    string
	Files   []*File
	Message string
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
	p := page{Title: title, Body: body, Files: nil, Message: nil}
	display(w, "challenges", p)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Open Database connection
	// db, err := sql.Open("mysql", "root:root@/tcp(52.20.186.36:3306)/test?tls=skip-verify$autocommit=true")
	db, err := sql.Open("mysql", "root:root@/test")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := page{Title: "Upload File", Body: nil, Files: nil, Message: nil}

	rows, err := db.Query("SELECT * FROM data")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			fmt.Println(columns[i], ": ", value)
		}
	}

	switch r.Method {
	//GET displays the upload form.
	case "GET":
		p.Message = "Choose files to upload"
		display(w, "upload", p)

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
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
