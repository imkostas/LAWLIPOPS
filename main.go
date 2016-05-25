package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Templates var holds a cached version of every template
var Templates = template.Must(template.ParseFiles(
	"templates/index.html",
	"templates/upload.html",
	"templates/challenges.html",
	"templates/case.html"))

// File struct is used to hold information about a given file on the server
type File struct {
	ID   string
	Path string
	Flag string
}

// Page struct holds information needed to Display a page
type Page struct {
	Title   string
	Body    string
	Files   []File
	Message string
}

// BinaryCase struct holds information about a case in the database
type BinaryCase struct {
	ID          string
	Title       string
	Summary     string
	FileFor     string
	FileAgainst string
	Date        string
	Archived    string
	Decision    string
}

// Display function shows a given template with the given data displayed
func Display(w http.ResponseWriter, name string, data interface{}) {
	err := Templates.ExecuteTemplate(w, name+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// t, _ := template.ParseFiles("challenges.html")
	// t.Execute(w, p)
}

// CheckError function determines if there was an error and displays a message if there was
func CheckError(w http.ResponseWriter, err error, msg string) {
	if err != nil {
		log.Println(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetCases function searches the test database and cases table with the given search string
func GetCases(w http.ResponseWriter, r *http.Request, searchString string) []BinaryCase {
	// Select all from cases table
	// db, err := sql.Open("mysql", "root:root@/test")
	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/test")
	CheckError(w, err, "Can't open db connection")

	defer db.Close()

	err = db.Ping()
	// CheckError(w, err, "Ping Error")
	if err != nil {
		panic(err.Error())
	}

	b := make([]BinaryCase, 0, 0)

	//	rows, err := db.Query("SELECT * FROM cases WHERE title LIKE '%" + searchString + "%'")
	rows, err := db.Query(searchString)
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

		c := BinaryCase{ID: "", Title: "", Summary: "", FileFor: "", FileAgainst: "", Date: "", Archived: "", Decision: ""}
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

// GetCase function searches the database for a case with the given case ID
func GetCase(w http.ResponseWriter, r *http.Request, caseID string) BinaryCase {
	// Select all from cases table
	// db, err := sql.Open("mysql", "root:root@/test")
	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/test")
	CheckError(w, err, "Can't open db connection")

	defer db.Close()

	stmntOut, err := db.Prepare("SELECT * FROM cases WHERE id = ?")
	CheckError(w, err, "")
	defer stmntOut.Close()

	c := BinaryCase{ID: "", Title: "", Summary: "", FileFor: "", FileAgainst: "", Date: "", Archived: "", Decision: ""}
	err = stmntOut.QueryRow(caseID).Scan(&c.ID,
		&c.Title,
		&c.Summary,
		&c.FileFor,
		&c.FileAgainst,
		&c.Date,
		&c.Archived,
		&c.Decision)

	CheckError(w, err, "Query Row error")

	return c
}

func main() {

	// r := mux.NewRouter()
	// r.HandleFunc("/", RootHandler)
	// r.HandleFunc("/challenges/", ChallengesHandler)
	// r.HandleFunc("/upload/", UploadHandler)
	// r.HandleFunc("/cases/{id}", CaseHandler)
	//
	// http.Handle("/", r)

	// http.HandleFunc("/", RootHandler)
	// http.HandleFunc("/challenges/", ChallengesHandler)
	// http.HandleFunc("/upload/", UploadHandler)
	// http.HandleFunc("/cases/", CaseHandler)
	// http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	r := NewRouter()
	s := http.StripPrefix("/css/", http.FileServer(http.Dir("./css/")))
	r.PathPrefix("/css/").Handler(s)
	log.Fatal(http.ListenAndServe(":8000", r))
	// err := http.ListenAndServe(":8000", r)
	// if err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
}
