package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	fb "github.com/huandu/facebook"
	"golang.org/x/oauth2"
	oauth2fb "golang.org/x/oauth2/facebook"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/urfave/negroni"
	"gopkg.in/gorp.v1"
)

// Templates var holds a cached version of every template
var Templates = template.Must(template.ParseFiles(
	"templates/index.html",
	"templates/challenge.html",
	"templates/case.html",
	"templates/login.html",
	"templates/account.html",
	"templates/navbar.html",
	"templates/register.html",
	"templates/dashboard.html",
	"templates/search.html"))

var local = false

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
	dbmap.AddTableWithName(Challenge{}, "challenges").SetKeys(true, "ID")
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(true, "ID")
	dbmap.AddTableWithName(User{}, "accounts").SetKeys(true, "ID")
	dbmap.CreateTablesIfNotExists()
}

var (
	oauthConf = &oauth2.Config{
		ClientID:     "1227119183985031",                 // Hide in OS.env vars if possible
		ClientSecret: "37a5b222db8324ddf14c0b0ea990eb71", // Hide in OS.env vars if possible
		RedirectURL:  "http://localhost:8000/facebookCallback",
		Scopes:       []string{"public_profile", "email"},
		Endpoint:     oauth2fb.Endpoint,
	}
	oauthStateString = RandomString(32)

	token, err = oauthConf.Exchange(oauth2.NoContext, "code")

	client = oauthConf.Client(oauth2.NoContext, token)

	session = &fb.Session{
		Version:    "v2.4",
		HttpClient: client,
	}

	res, _ = session.Get("/me", nil)
)

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
	Challenges   []Challenge
	CurrentUser  User
	UserLoggedIn bool
}

// BinaryCase struct holds information about a case in the database
type BinaryCase struct {
	ID            int64  `db:"id"`
	Title         string `db:"title"`
	Summary       string `db:"summary"`
	FileFor       string `db:"file_for"`
	FileAgainst   string `db:"file_against"`
	Date          string `db:"date_created"`
	Archived      string `db:"archived"`
	FinalDecision string `db:"final_decision"`
}

//TODO: Put in database
type Challenge struct {
	ID       int64  `db:"id"`
	Title    string `db:"title"`
	Summary  string `db:"summary"`
	Date     string `db:"date_created"`
	EndDate  string `db:"end_date"`
	Archived string `db:"archived"`
	Owner    string //User   //string `db:"owner"`
	Reward   string `db:"reward"`
	Awardee  string //User   //string `db:"awardee"`
	Comments string //[]Comment
}

//TODO: Put in database
type Comment struct {
	ID     int64  `db:"id"`
	Title  string `db:"title"`
	Body   string `db:"body"`
	Video  string `db:"video"`
	Author string `db:"author"`
	Votes  string `db:"votes"`
}

// User struct contains information about the current user
type User struct {
	ID         int64  `db:"id"`
	Username   string `db:"username"`
	Secret     []byte `db:"hash"`
	Email      string `db:"email"`
	Score      int    `db:"score"`
	Suspended  bool   `db:"suspended"`
	FacebookID string `db:"facebook_id"`
}

// Vote struct contrains info about what decision a user made about a case
type Vote struct {
	ID            int64 `db:"id"`
	AccountID     int64 `db:"account_id"`
	CaseID        int64 `db:"case_id"`
	UserDecision  int   `db:"user_decision"`
	FinalDecision int   `db:"final_decision"`
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
	err := dbmap.SelectOne(&c, "SELECT * FROM cases WHERE id=?", caseID)
	CheckError(w, err, "gorp SelectOne error")

	return c
}

func GetChallenges(w http.ResponseWriter, r *http.Request, queryString string) []Challenge {
	b := make([]Challenge, 0, 0)

	_, err := dbmap.Select(&b, queryString)
	CheckError(w, err, "dbmap Select error")

	return b
}

// GetChallenge function searches the database for a case with the given challenge ID
func GetChallenge(w http.ResponseWriter, r *http.Request, challengeID string) Challenge {
	c := Challenge{}

	// Select all from cases table
	err := dbmap.SelectOne(&c, "SELECT * FROM challenges WHERE id=?", challengeID)
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

// SetFinalDecision function takes the id and decision variable and updates the database
// decision: 1 is affirm, 2 is reverse
func SetFinalDecision(id int64, decision int) {
	// 1) Set case final_decision param to decision value
	// 2) Set case to archived
	dbmap.Exec("UPDATE cases SET final_decision=?, archived=1 WHERE id=?", decision, id)
	// 3) Find all accountIDs of users who answered correctly
	var ids []string
	dbmap.Select(&ids, "SELECT account_id FROM votes WHERE case_id=? AND user_decision=?", id, decision)
	// 4) Add points to each user who asnwered correctly
	for _, id := range ids {
		dbmap.Exec("UPDATE accounts SET score=score+? WHERE id=?", 10, id)
	}

	//TODO:
	// Set all votes to have the correct final_decision
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
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
	r.PathPrefix("/logo/").Handler(http.StripPrefix("/logo/", http.FileServer(http.Dir("./logo/"))))
	r.PathPrefix("/font-awesome-4.6.3/").Handler(http.StripPrefix("/font-awesome-4.6.3/", http.FileServer(http.Dir("./font-awesome-4.6.3/"))))

	// Set up and run negroni
	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(VerifyDatabase))

	n.UseHandler(r)
	n.Run(":8000")
}
