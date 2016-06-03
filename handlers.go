package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
)

const emptyFileString = "(empty)"

// POST Const = "POST"
const POST = "POST"

// RootHandler function creates the home page view
func RootHandler(w http.ResponseWriter, r *http.Request) {
	var p = Page{}
	errorString := ""

	if r.Method == POST {
		if r.FormValue("submitNewCase") != "" {
			c := BinaryCase{}
			// TODO:
			// Validate form
			if r.FormValue("title") != "" {
				c.Title = r.FormValue("title")
			} else {
				errorString = "Error: Please enter a title"
			}

			if r.FormValue("summary") != "" {
				c.Summary = r.FormValue("summary")
			} else {
				errorString = "Error: Please enter a summary"
			}

			if r.FormValue("file-for") != "" {
				c.FileFor = r.FormValue("file-for")
			} else {
				// CHANGE TO THROW Error
				c.FileFor = emptyFileString
			}

			if r.FormValue("file-against") != "" {
				c.FileAgainst = r.FormValue("file-against")
			} else {
				// CHANGE TO THROW Error
				c.FileAgainst = emptyFileString
			}

			// Update cases database
			if errorString == "" {
				c.Date = time.Now().UTC().Format("2006-01-02")
				err := dbmap.Insert(&c)
				CheckError(w, err, "Insert error")
				p.Message = "New case added successfuly!"
			}
		} else if r.FormValue("affirm") != "" {
			id, _ := strconv.ParseInt(strings.Split(r.FormValue("affirm"), "-")[1], 10, 64)
			dbmap.Exec("UPDATE cases SET decision=? WHERE id=?", 1, id)
		} else if r.FormValue("reverse") != "" {
			id, _ := strconv.ParseInt(strings.Split(r.FormValue("reverse"), "-")[1], 10, 64)
			dbmap.Exec("UPDATE cases SET decision=? WHERE id=?", 2, id)
		} else if r.FormValue("delete") != "" {
			id, _ := strconv.ParseInt(strings.Split(r.FormValue("delete"), "-")[1], 10, 64)
			dbmap.Exec("DELETE FROM cases WHERE id=?", id)
		} else if r.FormValue("save") != "" {
			id, _ := strconv.ParseInt(strings.Split(r.FormValue("save"), "-")[1], 10, 64)
			fileFor := ""
			fileAgainst := ""
			if r.FormValue("file-for") != "" {
				fileFor = r.FormValue("file-for")
			} else {
				fileFor = emptyFileString
			}

			if r.FormValue("file-against") != "" {
				fileAgainst = r.FormValue("file-against")
			} else {

				fileAgainst = emptyFileString
			}

			dbmap.Exec("UPDATE cases SET title=?, summary=?, file_for=?, file_against=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), fileFor, fileAgainst, id)
		}
	}

	c := GetCases(w, r, "SELECT * FROM cases WHERE archived = 0")

	session, _ := store.Get(r, "lawlipops")

	val := session.Values["userLoggedIn"]
	loggedIn, _ := val.(bool)
	// if !ok {
	// 	loggedIn = false
	// }
	// loggedIn, _ := strconv.ParseBool(loggedInString)

	val = session.Values["currentUser"]
	currentUser := &User{}
	currentUser, _ = val.(*User)
	// if !ok {
	// 	// Blind panic
	// 	// log.Fatal("Error getting user")
	//
	// }

	p.UserLoggedIn = loggedIn
	if loggedIn {
		p.CurrentUser = *currentUser
	}
	p.Cases = append(p.Cases, c...)
	p.Error = errorString

	if p.CurrentUser.Username == "chris@test.com" {
		Display(w, "dashboard", p)
	} else {
		Display(w, "index", p)
	}
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
	case POST:
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

	val := session.Values["userLoggedIn"]
	loggedIn, _ := val.(bool)
	// if !ok {
	// }

	val = session.Values["currentUser"]
	currentUser := &User{}
	currentUser, _ = val.(*User)
	// if !ok {
	// 	// Blind panic
	// 	// log.Fatal("Error getting user")
	//
	// }

	var p = Page{}
	// p := Page{Title: "", Body: "", Files: nil, Message: "", Error: "", Cases: nil, CurrentUser: nil, UserLoggedIn: false}
	p.UserLoggedIn = loggedIn
	if loggedIn {
		p.CurrentUser = *currentUser
	}

	errorString := ""
	// Set up bcrypt hash
	// if r.FormValue("register") != "" {
	// 	user := User{ID: -1, Username: "", Nickname: "", Secret: nil, Email: "", Score: -1, Suspended: false}
	// 	_ = dbmap.SelectOne(&user, "select * from accounts where username=?", r.FormValue("username"))
	// 	if user.ID != -1 {
	// 		errorString = "Username already taken"
	// 	} else {
	// 		secret, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
	// 		// TODO: Get seperate username or parse from email
	// 		user = User{-1, r.FormValue("username"), r.FormValue("username"), secret, r.FormValue("username"), 0, false}
	// 		if err := dbmap.Insert(&user); err != nil {
	// 			errorString = err.Error()
	// 		} else {
	// 			session.Values["userLoggedIn"] = true
	// 			session.Values["currentUser"] = user
	// 			err := session.Save(r, w)
	// 			CheckError(w, err, "err")
	// 			// session.Values["userLoggedIn"]
	// 			http.Redirect(w, r, "/", http.StatusFound)
	// 			return
	// 		}
	// 	}
	/*} else */ if r.FormValue("login") != "" {
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
				session.Values["userLoggedIn"] = true
				session.Values["currentUser"] = user
				err := session.Save(r, w)
				CheckError(w, err, "err")
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}

	// session.Values["test"] = "good"
	// session.Save(r, w)
	p.Error = errorString
	Display(w, "login", p)
}

// LogoutHandler function handles the log out loic
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "lawlipops")
	// TODO:
	// Set UserLoggedIn to false
	session.Values["userLoggedIn"] = false
	// Set CurrentUser to nil
	session.Values["currentUser"] = nil
	err := session.Save(r, w)
	CheckError(w, err, "err")
	http.Redirect(w, r, "/", http.StatusFound)
}

// AccountHandler function displays a users account page
func AccountHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "lawlipops")

	val := session.Values["userLoggedIn"]
	loggedIn, ok := val.(bool)
	if !ok {
		log.Println("Error getting userLoggedIn value")
	}

	val = session.Values["currentUser"]
	currentUser := &User{}
	currentUser, ok = val.(*User)
	if !ok {
		log.Println("Error getting current user")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	p := Page{}
	p.CurrentUser = *currentUser
	p.UserLoggedIn = loggedIn

	if !loggedIn {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	errorString := ""

	switch r.Method {
	case "GET":
	// Display account information
	// Display(w, "account", p)
	case POST:
		if r.FormValue("submit-password") != "" {
			if err := bcrypt.CompareHashAndPassword(p.CurrentUser.Secret, []byte(r.FormValue("currentPassword"))); err != nil {
				errorString = "Current password field is incorrect"
			} else {
				if r.FormValue("newPassword1") == r.FormValue("newPassword2") {
					// Successful
					secret, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("newPassword2")), bcrypt.DefaultCost)
					p.CurrentUser.Secret = secret
					// if _, err := dbmap.Update(p.CurrentUser); err != nil {
					// 	errorString = "Error updating database"
					// }
					_, err := dbmap.Exec("UPDATE accounts SET hash=? WHERE id=?", secret, p.CurrentUser.ID)
					if err != nil {
						errorString = "Error updating database"
					}
					session.Values["currentUser"] = p.CurrentUser
					session.Save(r, w)
					p.Message = "Password update successful"
				} else {
					errorString = "Passwords do not match"
				}
			}
		}

		if r.FormValue("submit-nickname") != "" {
			p.CurrentUser.Nickname = r.FormValue("nickname")
			_, err := dbmap.Exec("UPDATE accounts SET nickname=? WHERE id=?", p.CurrentUser.Nickname, p.CurrentUser.ID)
			if err != nil {
				errorString = "Error updating database"
			}
			session.Values["currentUser"] = p.CurrentUser
			session.Save(r, w)
			p.Message = "Nickname update successful"
		}

		// http.Redirect(w, r, "/account", http.StatusFound)
		// return
		// END POST
	default:
		http.Redirect(w, r, "/", http.StatusForbidden)
	}

	// if errorString != "" {
	p.Error = errorString
	// } else {
	// 	p.Error = ""
	// }
	Display(w, "account", p)
}

// RegisterHandler function handles the displaying and logic of the register form
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "lawlipops")

	val := session.Values["userLoggedIn"]
	loggedIn, ok := val.(bool)
	if !ok {
		log.Println("Error getting userLoggedIn value")
	}

	val = session.Values["currentUser"]
	currentUser := &User{}
	currentUser, ok = val.(*User)
	if !ok {
		log.Println("Error getting current user")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	p := Page{}
	p.CurrentUser = *currentUser
	p.UserLoggedIn = loggedIn

	if !loggedIn {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	errorString := ""

	if r.FormValue("submit") != "" {
		user := User{ID: -1, Username: "", Nickname: "", Secret: nil, Email: "", Score: -1, Suspended: false}
		_ = dbmap.SelectOne(&user, "select * from accounts where username=?", r.FormValue("username"))
		if user.ID != -1 {
			errorString = "Username already taken"
		} else {
			secret, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
			// TODO: Get seperate username or parse from email
			if r.FormValue("nickname") != "" {
				user = User{-1, r.FormValue("username"), r.FormValue("nickname"), secret, r.FormValue("email"), 0, false}
			} else {
				user = User{-1, r.FormValue("username"), r.FormValue("username"), secret, r.FormValue("email"), 0, false}
			}
			if err := dbmap.Insert(&user); err != nil {
				errorString = err.Error()
			} else {
				session.Values["userLoggedIn"] = true
				session.Values["currentUser"] = user
				err := session.Save(r, w)
				CheckError(w, err, "err")
				// session.Values["userLoggedIn"]
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}

	p.Error = errorString
	Display(w, "register", p)
}
