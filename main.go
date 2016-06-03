package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/urfave/negroni"
	"gopkg.in/gorp.v1"
)

// Templates var holds a cached version of every template
var Templates = template.Must(template.ParseFiles(
	"templates/index.html",
	"templates/upload.html",
	"templates/challenges.html",
	"templates/case.html",
	"templates/login.html",
	"templates/account.html",
	"templates/navbar.html",
	"templates/register.html",
	"templates/dashboard.html"))

var local = false

// var store = sessions.NewCookieStore([]byte("SECRET-CODE-TO-REPLACE"))
var store = sessions.NewCookieStore([]byte("SECRET-CODE-TO-REPLACE"))

const localString string = "root:root@tcp(localhost:8889)/test"
const serverString string = "root:root@/test"

var db *sql.DB
var dbmap *gorp.DbMap

func initDB() {
	gob.Register(&User{})

	// Set up the database connection
	var connectionString = ""
	if local {
		connectionString = localString
	} else {
		connectionString = serverString
	}
	db, _ = sql.Open("mysql", connectionString)

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	dbmap.AddTableWithName(BinaryCase{}, "cases").SetKeys(true, "ID")
	dbmap.AddTableWithName(User{}, "accounts").SetKeys(true, "ID")
	dbmap.CreateTablesIfNotExists()
}

// File struct is used to hold information about a given file on the server
type File struct {
	ID   string
	Path string
	Flag string
}

// Page struct holds information needed to Display a page
type Page struct {
	Title        string
	Body         string
	Files        []File
	Message      string
	Error        string
	Cases        []BinaryCase
	CurrentUser  User
	UserLoggedIn bool
}

// BinaryCase struct holds information about a case in the database
type BinaryCase struct {
	ID          int64  `db:"id"`
	Title       string `db:"title"`
	Summary     string `db:"summary"`
	FileFor     string `db:"file_for"`
	FileAgainst string `db:"file_against"`
	Date        string `db:"date_created"`
	Archived    string `db:"archived"`
	Decision    string `db:"decision"`
}

// User struct contains information about the current user
type User struct {
	ID        int64  `db:"id"`
	Username  string `db:"username"`
	Nickname  string `db:"nickname"`
	Secret    []byte `db:"hash"`
	Email     string `db:"email"`
	Score     int    `db:"score"`
	Suspended bool   `db:"suspended"`
}

// Display function shows a given template with the given data displayed
func Display(w http.ResponseWriter, name string, data interface{}) {
	err := Templates.ExecuteTemplate(w, name+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// CheckError function determines if there was an error and displays a message if there was
func CheckError(w http.ResponseWriter, err error, msg string) {
	if err != nil {
		log.Println(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetCases function searches the test database and cases table with the given search string
func GetCases(w http.ResponseWriter, r *http.Request, queryString string) []BinaryCase {
	b := make([]BinaryCase, 0, 0)

	_, err := dbmap.Select(&b, queryString)
	CheckError(w, err, "dbmap Select error")

	return b
}

// GetCase function searches the database for a case with the given case ID
func GetCase(w http.ResponseWriter, r *http.Request, caseID string) BinaryCase {
	c := BinaryCase{}

	// Select all from cases table
	err := dbmap.SelectOne(&c, "SELECT * FROM cases WHERE id = ?", caseID)
	CheckError(w, err, "gorp SelectOne error")

	return c
}

// VerifyDatabase function pings the database to make the application can connect to the database
func VerifyDatabase(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if err := db.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	next(w, r)
}

func main() {
	// Check for program flags
	flag.BoolVar(&local, "local", false, "Defines if the environment is local or not")
	flag.Parse()

	initDB()
	defer db.Close()

	// Set up Gorilla Mux router
	r := NewRouter()
	s := http.StripPrefix("/css/", http.FileServer(http.Dir("./css/")))

	r.PathPrefix("/css/").Handler(s)
	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))

	// Set up and run negroni
	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(VerifyDatabase))
	// store := cookiestore.New([]byte("SECRET-CODE-TO-REPLACE"))
	// n.Use(sessions.Sessions("lawlipops", store))
	// TODO: Fix

	n.UseHandler(r)
	n.Run(":8000")
}
