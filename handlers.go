package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
)

// RootHandler function creates the home page view
func RootHandler(w http.ResponseWriter, r *http.Request) {
	c := GetCases(w, r, "SELECT * FROM cases WHERE archived = 0")

	session, _ := store.Get(r, "lawlipops")

	val := session.Values["userLoggedIn"]
	log.Println(val)
	loggedInString, ok := val.(string)
	if !ok {
		// loggedIn = false
	}
	loggedIn, _ := strconv.ParseBool(loggedInString)

	val = session.Values["currentUser"]
	currentUser, ok := val.(User)
	if !ok {
		// Blind panic
	}

	// val = session.Values["test"]
	// message, ok := val.(string)
	// if !ok {
	// 	// Panic
	// }
	// log.Println("msg: " + message)

	var p = Page{}
	// p := Page{Title: "", Body: "", Files: nil, Message: "", Error: "", Cases: nil, CurrentUser: nil, UserLoggedIn: false}

	p.UserLoggedIn = loggedIn
	p.CurrentUser = currentUser
	p.Cases = append(p.Cases, c...)
	// p.UserLoggedIn = false
	// log.Printf("%v", p)
	// log.Println(p.UserLoggedIn)
	Display(w, "index", p)
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
	p := Page{Title: "Upload File", Body: "", Files: nil, Message: ""}

	//TODO: Update db operations to use gorp

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
	// vars := mux.Vars(r)
	// caseID := vars["id"]
	caseID := mux.Vars(r)["id"]
	caseToDisplay := GetCase(w, r, caseID)

	Display(w, "case", caseToDisplay)
}

// LoginHandler function
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "lawlipops")
	CheckError(w, err, "")
	errorString := ""
	// Set up bcrypt hash
	if r.FormValue("register") != "" {
		user := User{ID: -1, Username: "", Secret: nil, Email: "", Score: -1, Suspended: false}
		_ = dbmap.SelectOne(&user, "select * from accounts where username=?", r.FormValue("username"))
		if user.ID != -1 {
			errorString = "Username already taken"
		} else {
			secret, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
			// TODO: Get seperate username or parse from email
			user = User{-1, r.FormValue("username"), secret, r.FormValue("username"), 0, false}
			if err := dbmap.Insert(&user); err != nil {
				errorString = err.Error()
			} else {
				session.Values["userLoggedIn"] = "true"
				session.Values["currentUser"] = user
				session.Save(r, w)
				// session.Values["userLoggedIn"]
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	} else if r.FormValue("login") != "" {
		// Validate account credentials
		// user, err := dbmap.Get(User{}, r.FormValue("username"))
		var user User
		err := dbmap.SelectOne(&user, "select * from accounts where username=?", string(r.FormValue("username")))
		if err != nil {
			// errorString = err.Error()
			errorString = "Username or password not recognized"
		} else if user.Username == "" {
			errorString = "No such user found with Username: " + r.FormValue("username")
		} else {
			if err := bcrypt.CompareHashAndPassword(user.Secret, []byte(r.FormValue("password"))); err != nil {
				errorString = err.Error()
			} else {
				// Login Successful
				// TODO: Set session vars
				session.Values["userLoggedIn"] = "true"
				session.Values["currentUser"] = user
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}

	// session.Values["test"] = "good"
	// session.Save(r, w)
	Display(w, "login", errorString)
}
