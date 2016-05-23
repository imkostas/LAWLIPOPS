package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var templates = template.Must(template.ParseFiles(
	"templates/index.html",
	"templates/upload.html",
	"templates/challenges.html",
	"templates/case.html"))

type file struct {
	ID   string
	Path string
	Flag string
}

type page struct {
	Title   string
	Body    string
	Files   []file
	Message string
}

type binaryCase struct {
	ID          string
	Title       string
	Summary     string
	FileFor     string
	FileAgainst string
	Date        string
	Archived    string
	Decision    string
}

func display(w http.ResponseWriter, name string, data interface{}) {
	err := templates.ExecuteTemplate(w, name+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// t, _ := template.ParseFiles("challenges.html")
	// t.Execute(w, p)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	c := getCases(w, r, "SELECT * FROM cases WHERE archived = 0")
	display(w, "index", c)
}

func challengesHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/challenges/"):]
	body := "Nothing to see here"
	p := page{Title: title, Body: body, Files: nil, Message: ""}
	display(w, "challenges", p)
}

func checkError(w http.ResponseWriter, err error, msg string) {
	if err != nil {
		log.Println(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Open Database connection
	// db, err := sql.Open("mysql", "root:root@/tcp(52.20.186.36:80)/test?tls=skip-verify$autocommit=true")
	// db, err := sql.Open("mysql", "root:root@/tcp(52.20.186.36)/test")
	db, err := sql.Open("mysql", "root:root@/test")
	checkError(w, err, "Can't open db connection")

	defer db.Close()
	err = db.Ping()
	checkError(w, err, "Ping Error")

	p := page{Title: "Upload File", Body: "", Files: nil, Message: ""}

	// varname := "lab"
	varname := ""
	rows, err := db.Query("SELECT * FROM data WHERE doc LIKE '%" + varname + "%'")
	checkError(w, err, "")

	columns, err := rows.Columns()
	checkError(w, err, "Error getting columns")

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		checkError(w, err, "")

		// var value string
		// for i, col := range values {
		// 	if col == nil {
		// 		value = "NULL"
		// 	} else {
		// 		value = string(col)
		// 	}
		// 	fmt.Println(columns[i], ": ", value)
		// }
		f := file{ID: "", Path: "", Flag: ""}
		f.ID = string(values[0])
		f.Path = string(values[1])
		f.Flag = string(values[2])
		p.Files = append(p.Files, f)
	}

	switch r.Method {
	// GET displays the upload form
	case "GET":
		p.Message = "Choose files to upload"
		display(w, "upload", p)

	// POST takes the uploaded file(s) saves it to disk, and updates the database
	case "POST":
		// parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		checkError(w, err, "")

		// get a ref to the parsed multipart form
		m := r.MultipartForm

		// get the *fileheaders
		files := m.File["myfiles"]
		for i, _ := range files {
			// for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			checkError(w, err, "")

			// create destination file making sure the path is writeable.
			dst, err := os.Create("files/" + files[i].Filename)
			// if err != nil {
			// 	fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
			// 	return
			// }
			checkError(w, err, "")

			defer dst.Close()
			checkError(w, err, "")

			// copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Insert into DB
			stmtIns, err := db.Prepare("INSERT INTO `data` (`id`, `doc`, `flag`) VALUES (NULL, ?, ?);")
			checkError(w, err, "")
			defer stmtIns.Close()

			_, err = stmtIns.Exec("files/"+files[i].Filename, 0)
			checkError(w, err, "")
		}

		// display success message.
		p.Message = "Upload successful"
		display(w, "upload", p)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getCases(w http.ResponseWriter, r *http.Request, searchString string) []binaryCase {
	// Select all from cases table
	db, err := sql.Open("mysql", "root:root@/test")
	checkError(w, err, "Can't open db connection")

	defer db.Close()
	err = db.Ping()
	checkError(w, err, "Ping Error")

	b := make([]binaryCase, 0, 0)

	//	rows, err := db.Query("SELECT * FROM cases WHERE title LIKE '%" + searchString + "%'")
	rows, err := db.Query(searchString)
	checkError(w, err, "")

	columns, err := rows.Columns()
	checkError(w, err, "Error getting columns")

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		checkError(w, err, "")

		c := binaryCase{ID: "", Title: "", Summary: "", FileFor: "", FileAgainst: "", Date: "", Archived: "", Decision: ""}
		c.ID = string(values[0])
		c.Title = string(values[1])
		c.Summary = string(values[2])
		c.FileFor = string(values[3])
		c.FileAgainst = string(values[4])
		c.Date = string(values[5])
		c.Archived = string(values[6])
		c.Decision = string(values[7])
		b = append(b, c)
	}

	return b
}

func getCase(w http.ResponseWriter, r *http.Request, caseID string) binaryCase {
	// Select all from cases table
	db, err := sql.Open("mysql", "root:root@/test")
	checkError(w, err, "Can't open db connection")

	defer db.Close()

	stmntOut, err := db.Prepare("SELECT * FROM cases WHERE id = ?")
	checkError(w, err, "")
	defer stmntOut.Close()

	c := binaryCase{ID: "", Title: "", Summary: "", FileFor: "", FileAgainst: "", Date: "", Archived: "", Decision: ""}
	err = stmntOut.QueryRow(caseID).Scan(&c.ID,
		&c.Title,
		&c.Summary,
		&c.FileFor,
		&c.FileAgainst,
		&c.Date,
		&c.Archived,
		&c.Decision)

	checkError(w, err, "")

	return c
}

func caseHandler(w http.ResponseWriter, r *http.Request) {
	caseID := r.URL.Path[len("/cases/"):]
	caseToDisplay := getCase(w, r, caseID)

	display(w, "case", caseToDisplay)
}

func main() {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/challenges/", challengesHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/cases/", caseHandler)
	// http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
