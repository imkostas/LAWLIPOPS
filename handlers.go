package main

import (
	"database/sql"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// RootHandler function creates the home page view
func RootHandler(w http.ResponseWriter, r *http.Request) {
	c := GetCases(w, r, "SELECT * FROM cases WHERE archived = 0")
	Display(w, "index", c)
}

// ChallengesHandler function creates a view for the challenge with the given id
func ChallengesHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/challenges/"):]
	body := "Nothing to see here"
	p := Page{Title: title, Body: body, Files: nil, Message: ""}
	Display(w, "challenges", p)
}

// UploadHandler function creates a view of uploaded files and handles the upload of files
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Open Database connection
	var connectionString = ""
	if local {
		connectionString = localString
	} else {
		connectionString = serverString
	}
	db, err := sql.Open("mysql", connectionString)
	CheckError(w, err, "Can't open db connection")

	defer db.Close()
	err = db.Ping()
	CheckError(w, err, "Ping Error")

	p := Page{Title: "Upload File", Body: "", Files: nil, Message: ""}

	// varname := "lab"
	varname := ""
	rows, err := db.Query("SELECT * FROM data WHERE doc LIKE '%" + varname + "%'")
	CheckError(w, err, "Rows error")

	columns, err := rows.Columns()
	CheckError(w, err, "Error getting columns")

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		CheckError(w, err, "")

		// var value string
		// for i, col := range values {
		// 	if col == nil {
		// 		value = "NULL"
		// 	} else {
		// 		value = string(col)
		// 	}
		// 	fmt.Println(columns[i], ": ", value)
		// }
		f := File{ID: "", Path: "", Flag: ""}
		f.ID = string(values[0])
		f.Path = string(values[1])
		f.Flag = string(values[2])
		p.Files = append(p.Files, f)
	}

	switch r.Method {
	// GET Displays the upload form
	case "GET":
		p.Message = "Choose files to upload"
		Display(w, "upload", p)

	// POST takes the uploaded file(s) saves it to disk, and updates the database
	case "POST":
		// parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		CheckError(w, err, "")

		// get a ref to the parsed multipart form
		m := r.MultipartForm

		// get the *fileheaders
		files := m.File["myfiles"]
		// for i, _ := range files {
		for i := 0; i < len(files); i++ {
			// for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			CheckError(w, err, "")

			// create destination file making sure the path is writeable.
			dst, err := os.Create("files/" + files[i].Filename)
			CheckError(w, err, "Unable to create the file for writing. Check your write access privilege")

			defer dst.Close()
			CheckError(w, err, "")

			// copy the uploaded file to the destination file
			if _, er := io.Copy(dst, file); err != nil {
				http.Error(w, er.Error(), http.StatusInternalServerError)
				return
			}

			// Insert into DB
			stmtIns, err := db.Prepare("INSERT INTO `data` (`id`, `doc`, `flag`) VALUES (NULL, ?, ?);")
			CheckError(w, err, "")
			defer stmtIns.Close()

			_, err = stmtIns.Exec("files/"+files[i].Filename, 0)
			CheckError(w, err, "")
		}

		// Display success message.
		p.Message = "Upload successful"
		Display(w, "upload", p)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// CaseIndex function creates a template view for every case
func CaseIndex(w http.ResponseWriter, r *http.Request) {
	// c := GetCases(w, r, "SELECT * FROM cases WHERE archived = 0")
	// w.Write(c)
}

// CaseHandler function creates a template for the case with the given id
func CaseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// caseID := r.URL.Path[len("/cases/"):]
	caseID := vars["id"]
	// log.Println("caseID: " + caseID)
	caseToDisplay := GetCase(w, r, caseID)

	Display(w, "case", caseToDisplay)
}
