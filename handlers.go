package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

const emptyFileString = "(empty)"
const post = "POST"

// RootHandler function creates the home page view
func RootHandler(w http.ResponseWriter, r *http.Request) {
	var p = Page{}
	errorString := ""

	// if r.Method == post {
	// 	if r.FormValue("submitNewCase") != "" {
	// 		c := BinaryCase{}
	// 		// Validate form
	// 		if r.FormValue("title") != "" {
	// 			c.Title = r.FormValue("title")
	// 		} else {
	// 			errorString = "Error: Please enter a title"
	// 		}
	//
	// 		if r.FormValue("summary") != "" {
	// 			c.Summary = r.FormValue("summary")
	// 		} else {
	// 			errorString = "Error: Please enter a summary"
	// 		}
	//
	// 		if r.FormValue("file-for") != "" {
	// 			c.FileFor = r.FormValue("file-for")
	// 		} else {
	// 			// CHANGE TO THROW Error
	// 			c.FileFor = emptyFileString
	// 		}
	//
	// 		if r.FormValue("file-against") != "" {
	// 			c.FileAgainst = r.FormValue("file-against")
	// 		} else {
	// 			// CHANGE TO THROW Error
	// 			c.FileAgainst = emptyFileString
	// 		}
	//
	// 		// Update cases database
	// 		if errorString == "" {
	// 			// if r.FormValue("file-for") != "" && r.FormValue("file-against") != "" {
	// 			// 	// TODO Upload files
	// 			// 	c.FileFor = fileUploadHandler(w, r, "file-for")
	// 			// 	c.FileAgainst = fileUploadHandler(w, r, "file-against")
	// 			// }
	//
	// 			c.FileFor = fileUploadHandler(w, r, "file-for")
	// 			c.FileAgainst = fileUploadHandler(w, r, "file-against")
	//
	// 			c.Date = time.Now().UTC().Format("2006-01-02")
	// 			err := dbmap.Insert(&c)
	// 			CheckError(w, err, "Insert error")
	// 			p.Message = "New case added successfuly!"
	// 		}
	// 	} else if r.FormValue("affirm") != "" {
	// 		id, _ := strconv.ParseInt(strings.Split(r.FormValue("affirm"), "-")[1], 10, 64)
	// 		//dbmap.Exec("UPDATE cases SET final_decision=? WHERE id=?", 1, id)
	// 		SetFinalDecision(id, 1)
	// 	} else if r.FormValue("reverse") != "" {
	// 		id, _ := strconv.ParseInt(strings.Split(r.FormValue("reverse"), "-")[1], 10, 64)
	// 		//dbmap.Exec("UPDATE cases SET final_decision=? WHERE id=?", 2, id)
	// 		SetFinalDecision(id, 2)
	// 	} else if r.FormValue("delete") != "" {
	// 		id, _ := strconv.ParseInt(strings.Split(r.FormValue("delete"), "-")[1], 10, 64)
	// 		dbmap.Exec("DELETE FROM cases WHERE id=?", id)
	// 	} else if r.FormValue("save") != "" {
	// 		id, _ := strconv.ParseInt(strings.Split(r.FormValue("save"), "-")[1], 10, 64)
	//
	// 		forString := fileUploadHandler(w, r, "file-for")
	// 		againstString := fileUploadHandler(w, r, "file-against")
	//
	// 		if forString != emptyFileString {
	// 			if againstString != emptyFileString {
	// 				dbmap.Exec("UPDATE cases SET title=?, summary=?, file_for=?, file_against=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), forString, againstString, id)
	// 			} else {
	// 				dbmap.Exec("UPDATE cases SET title=?, summary=?, file_for=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), forString, id)
	// 			}
	// 		} else if againstString != emptyFileString {
	// 			dbmap.Exec("UPDATE cases SET title=?, summary=?, file_against=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), againstString, id)
	// 		} else {
	// 			dbmap.Exec("UPDATE cases SET title=?, summary=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), id)
	// 		}
	// 	}
	// }

	// c := GetCases(w, r, "SELECT * FROM cases WHERE archived = 0")
	c := GetCases(w, r, "SELECT * FROM cases")
	ch := GetChallenges(w, r, "SELECT * FROM challenges")

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

		// Get updated account information from the server
		dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		session.Values["currentUser"] = p.CurrentUser
		session.Save(r, w)
	}
	p.Cases = append(p.Cases, c...)
	p.Challenges = append(p.Challenges, ch...)
	p.Error = errorString
	//
	// if p.CurrentUser.Email == "chris@test.com" {
	// 	Display(w, "dashboard", p)
	// } else {
	Display(w, "index", p)
	// }
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	var p = Page{}
	errorString := ""

	session, _ := store.Get(r, "lawlipops")

	val := session.Values["userLoggedIn"]
	loggedIn, _ := val.(bool)

	val = session.Values["currentUser"]
	currentUser := &User{}
	currentUser, _ = val.(*User)

	p.UserLoggedIn = loggedIn
	if loggedIn {
		p.CurrentUser = *currentUser

		// Get updated account information from the server
		dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		session.Values["currentUser"] = p.CurrentUser
		session.Save(r, w)
	}

	if p.CurrentUser.Email == "chris@test.com" {
		if r.Method == post {
			if r.FormValue("submitNewCase") != "" {
				c := BinaryCase{}
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
					// if r.FormValue("file-for") != "" && r.FormValue("file-against") != "" {
					// 	// TODO Upload files
					// 	c.FileFor = fileUploadHandler(w, r, "file-for")
					// 	c.FileAgainst = fileUploadHandler(w, r, "file-against")
					// }

					c.FileFor = fileUploadHandler(w, r, "file-for")
					c.FileAgainst = fileUploadHandler(w, r, "file-against")

					c.Date = time.Now().UTC().Format("2006-01-02")
					err := dbmap.Insert(&c)
					CheckError(w, err, "Insert error")
					p.Message = "New case added successfuly!"
				}
			} else if r.FormValue("affirm") != "" {
				id, _ := strconv.ParseInt(strings.Split(r.FormValue("affirm"), "-")[1], 10, 64)
				//dbmap.Exec("UPDATE cases SET final_decision=? WHERE id=?", 1, id)
				SetFinalDecision(id, 1)
			} else if r.FormValue("reverse") != "" {
				id, _ := strconv.ParseInt(strings.Split(r.FormValue("reverse"), "-")[1], 10, 64)
				//dbmap.Exec("UPDATE cases SET final_decision=? WHERE id=?", 2, id)
				SetFinalDecision(id, 2)
			} else if r.FormValue("delete") != "" {
				id, _ := strconv.ParseInt(strings.Split(r.FormValue("delete"), "-")[1], 10, 64)
				dbmap.Exec("DELETE FROM cases WHERE id=?", id)
			} else if r.FormValue("save") != "" {
				id, _ := strconv.ParseInt(strings.Split(r.FormValue("save"), "-")[1], 10, 64)

				forString := fileUploadHandler(w, r, "file-for")
				againstString := fileUploadHandler(w, r, "file-against")

				if forString != emptyFileString {
					if againstString != emptyFileString {
						dbmap.Exec("UPDATE cases SET title=?, summary=?, file_for=?, file_against=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), forString, againstString, id)
					} else {
						dbmap.Exec("UPDATE cases SET title=?, summary=?, file_for=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), forString, id)
					}
				} else if againstString != emptyFileString {
					dbmap.Exec("UPDATE cases SET title=?, summary=?, file_against=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), againstString, id)
				} else {
					dbmap.Exec("UPDATE cases SET title=?, summary=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), id)
				}
			}
		}

		c := GetCases(w, r, "SELECT * FROM cases")
		ch := GetChallenges(w, r, "SELECT * FROM challenges")

		p.Cases = append(p.Cases, c...)
		p.Challenges = append(p.Challenges, ch...)
		p.Error = errorString

		Display(w, "dashboard", p)
	} else {
		http.Redirect(w, r, "/", http.StatusForbidden)
	}
}

func DashboardChallengesHandler(w http.ResponseWriter, r *http.Request) {
	var p = Page{}
	errorString := ""

	if r.FormValue("submitNewChallenge") != "" {
		c := Challenge{}
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

		// Update challenges database
		if errorString == "" {
			// if r.FormValue("file-for") != "" && r.FormValue("file-against") != "" {
			// 	// TODO Upload files
			// 	c.FileFor = fileUploadHandler(w, r, "file-for")
			// 	c.FileAgainst = fileUploadHandler(w, r, "file-against")
			// }

			c.Date = time.Now().UTC().Format("2006-01-02")
			err := dbmap.Insert(&c)
			CheckError(w, err, "Insert error")
			p.Message = "New case added successfuly!"
			log.Println(p.Message)
		}
	} else if r.FormValue("delete") != "" {
		log.Println(r.FormValue("delete"))
		id, _ := strconv.ParseInt(strings.Split(r.FormValue("delete"), "-")[1], 10, 64)
		dbmap.Exec("DELETE FROM challenges WHERE id=?", id)
		log.Println("delete")
	} else if r.FormValue("save") != "" {
		id, _ := strconv.ParseInt(strings.Split(r.FormValue("save"), "-")[1], 10, 64)
		dbmap.Exec("UPDATE challenges SET title=?, summary=? WHERE id=?", r.FormValue("title"), r.FormValue("summary"), id)
		log.Println("save")
	}
	log.Println(errorString)
	log.Println("exit")
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}

func fileUploadHandler(w http.ResponseWriter, r *http.Request, fileKey string) string {
	file, header, err := r.FormFile(fileKey)

	if err == nil {
		CheckError(w, err, "")
		defer file.Close()

		out, err := os.Create("files/" + header.Filename)
		CheckError(w, err, "")
		defer out.Close()

		_, err = io.Copy(out, file)
		CheckError(w, err, "")

		return header.Filename
	}

	return emptyFileString
}

// ChallengesHandler function creates a view for the challenge with the given id
func ChallengesHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/challenges/"):]
	body := "Nothing to see here"
	p := Page{Title: title, Body: body, Files: nil, Message: ""}
	Display(w, "challenges", p)
}

// // UploadHandler function creates a view of uploaded files and handles the upload of files
// func UploadHandler(w http.ResponseWriter, r *http.Request) {
// 	p := Page{Title: "Upload File", Body: "", Files: nil, Message: ""}
//
// 	//TODO: Update db operations to use gorp
//
// 	varname := ""
// 	rows, err := db.Query("SELECT * FROM data WHERE doc LIKE '%" + varname + "%'")
// 	CheckError(w, err, "Rows error")
//
// 	columns, err := rows.Columns()
// 	CheckError(w, err, "Error getting columns")
//
// 	values := make([]sql.RawBytes, len(columns))
//
// 	scanArgs := make([]interface{}, len(values))
// 	for i := range values {
// 		scanArgs[i] = &values[i]
// 	}
//
// 	for rows.Next() {
// 		err = rows.Scan(scanArgs...)
// 		CheckError(w, err, "")
//
// 		f := File{ID: "", Path: "", Flag: ""}
// 		f.ID = string(values[0])
// 		f.Path = string(values[1])
// 		f.Flag = string(values[2])
// 		p.Files = append(p.Files, f)
// 	}
//
// 	switch r.Method {
// 	// GET Displays the upload form
// 	case "GET":
// 		p.Message = "Choose files to upload"
// 		Display(w, "upload", p)
//
// 	// POST takes the uploaded file(s) saves it to disk, and updates the database
// 	case post:
// 		// parse the multipart form in the request
// 		err := r.ParseMultipartForm(100000)
// 		CheckError(w, err, "")
//
// 		// get a ref to the parsed multipart form
// 		m := r.MultipartForm
//
// 		// get the *fileheaders
// 		files := m.File["myfiles"]
// 		for i := 0; i < len(files); i++ {
// 			// for each fileheader, get a handle to the actual file
// 			file, err := files[i].Open()
// 			defer file.Close()
// 			CheckError(w, err, "")
//
// 			// create destination file making sure the path is writeable.
// 			dst, err := os.Create("files/" + files[i].Filename)
// 			CheckError(w, err, "Unable to create the file for writing. Check your write access privilege")
//
// 			defer dst.Close()
// 			// CheckError(w, err, "")
//
// 			// copy the uploaded file to the destination file
// 			if _, er := io.Copy(dst, file); err != nil {
// 				http.Error(w, er.Error(), http.StatusInternalServerError)
// 				return
// 			}
//
// 			// Insert into DB
// 			stmtIns, err := db.Prepare("INSERT INTO `data` (`id`, `doc`, `flag`) VALUES (NULL, ?, ?);")
// 			CheckError(w, err, "")
// 			defer stmtIns.Close()
//
// 			_, err = stmtIns.Exec("files/"+files[i].Filename, 0)
// 			CheckError(w, err, "")
// 		}
//
// 		// Display success message.
// 		p.Message = "Upload successful"
// 		Display(w, "upload", p)
// 	default:
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

// CaseHandler function creates a template for the case with the given id
func CaseHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{}

	session, err := store.Get(r, "lawlipops")
	CheckError(w, err, "")

	val := session.Values["userLoggedIn"]
	loggedIn, ok := val.(bool)
	if !ok {
		log.Println("Error getting userLoggedIn value")
	}

	if loggedIn {
		val = session.Values["currentUser"]
		currentUser := &User{}
		currentUser, ok = val.(*User)
		if !ok {
			log.Println("Error getting current user")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		p.CurrentUser = *currentUser

		// Get updated account information from the server
		dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		session.Values["currentUser"] = p.CurrentUser
		session.Save(r, w)
	}

	p.UserLoggedIn = loggedIn

	errorString := ""

	if r.Method == post {
		if r.FormValue("affirm") != "" {
			if p.UserLoggedIn {
				id, _ := strconv.ParseInt(strings.Split(r.FormValue("affirm"), "-")[1], 10, 64)
				var existingVote string
				dbmap.SelectOne(&existingVote, "SELECT id FROM `votes` WHERE account_id=? AND case_id=?", p.CurrentUser.ID, id)

				if existingVote == "" {
					dbmap.Exec("INSERT INTO `votes` (id, account_id, case_id, user_decision, final_decision) VALUES (NULL,?,?,?,?)", p.CurrentUser.ID, id, 1, 0)
					p.Message = "Vote registered successfuly"
				} else {
					dbmap.Exec("UPDATE `votes` SET user_decision=? WHERE account_id=? AND case_id=?", 1, p.CurrentUser.ID, id)
					p.Message = "Vote updated successfuly"
				}
			} else {
				errorString = "You must be logged in to vote"
			}
		} else if r.FormValue("reverse") != "" {
			if p.UserLoggedIn {
				id, _ := strconv.ParseInt(strings.Split(r.FormValue("reverse"), "-")[1], 10, 64)
				var existingVote string
				dbmap.SelectOne(&existingVote, "SELECT id FROM `votes` WHERE account_id=? AND case_id=?", p.CurrentUser.ID, id)

				if existingVote == "" {
					dbmap.Exec("INSERT INTO `votes` (id, account_id, case_id, user_decision, final_decision) VALUES (NULL,?,?,?,?)", p.CurrentUser.ID, id, 2, 0)
					p.Message = "Vote registered successfuly"
				} else {
					dbmap.Exec("UPDATE `votes` SET user_decision=? WHERE account_id=? AND case_id=?", 2, p.CurrentUser.ID, id)
					p.Message = "Vote updated successfuly"
				}
			} else {
				errorString = "You must be logged in to vote"
			}
		}
	}

	// vars := mux.Vars(r)
	// caseID := vars["id"]
	caseID := mux.Vars(r)["id"]
	caseToDisplay := GetCase(w, r, caseID)

	p.Cases = append(p.Cases, caseToDisplay)
	p.Error = errorString

	Display(w, "case", p)
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

	if r.FormValue("login") != "" {
		var user User
		err := dbmap.SelectOne(&user, "select * from accounts where email=?", string(r.FormValue("email")))
		if err != nil {
			// errorString = err.Error()
			errorString = "Email or password not recognized"
		} else if user.Email == "" {
			errorString = "No such user found with Email: " + r.FormValue("email")
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

	p := Page{}
	p.UserLoggedIn = loggedIn

	if loggedIn {
		val = session.Values["currentUser"]
		currentUser := &User{}
		currentUser, ok = val.(*User)
		if !ok {
			log.Println("Error getting current user")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		p.CurrentUser = *currentUser
		// Get updated account information from the server
		dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		session.Values["currentUser"] = p.CurrentUser
		session.Save(r, w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	errorString := ""

	switch r.Method {
	case "GET":
	// Display account information
	// Display(w, "account", p)
	case post:
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

		if r.FormValue("submit-username") != "" {
			p.CurrentUser.Username = r.FormValue("username")
			_, err := dbmap.Exec("UPDATE accounts SET username=? WHERE id=?", p.CurrentUser.Username, p.CurrentUser.ID)
			if err != nil {
				errorString = "Error updating database"
			}
			session.Values["currentUser"] = p.CurrentUser
			session.Save(r, w)
			p.Message = "Username update successful"
		}

		// if r.FormValue("refresh") != "" {
		// 	dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		// 	session.Values["currentUser"] = p.CurrentUser
		// 	session.Save(r, w)
		// 	p.Message = "Account information has been updated"
		// }

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

	p := Page{}

	errorString := ""

	if r.Method == post {
		user := User{ID: -1, Username: "", Secret: nil, Email: "", Score: -1, Suspended: false}
		_ = dbmap.SelectOne(&user, "select * from accounts where email=?", r.FormValue("email"))
		if user.ID != -1 {
			errorString = "Email already associated with an account"
		} else {
			secret, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
			// TODO: Get seperate username or parse from email
			if r.FormValue("username") != "" {
				user = User{-1, r.FormValue("username"), secret, r.FormValue("email"), 0, false, "-1"}
			} else {
				user = User{-1, r.FormValue("email"), secret, r.FormValue("email"), 0, false, "-1"}
			}
			if err := dbmap.Insert(&user); err != nil {
				errorString = err.Error()
			} else {
				session.Values["userLoggedIn"] = true
				session.Values["currentUser"] = user
				err := session.Save(r, w)
				CheckError(w, err, "err")

				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}

	p.Error = errorString
	Display(w, "register", p)
}

// SearchHandler func
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "lawlipops")

	val := session.Values["userLoggedIn"]
	loggedIn, ok := val.(bool)
	if !ok {
		log.Println("Error getting userLoggedIn value")
	}

	p := Page{}
	p.UserLoggedIn = loggedIn

	if loggedIn {
		val = session.Values["currentUser"]
		currentUser := &User{}
		currentUser, ok = val.(*User)
		if !ok {
			log.Println("Error getting current user")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		p.CurrentUser = *currentUser
		// Get updated account information from the server
		dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		session.Values["currentUser"] = p.CurrentUser
		session.Save(r, w)
	}

	errorString := ""

	queryString := r.FormValue("query")
	/** LEAVE THESE COMMENTS **/
	// c := GetCases(w, r, "SELECT * FROM cases WHERE title RLIKE '"+queryString+"' OR summary RLIKE '"+queryString+"'")
	// c := GetCases(w, r, "SELECT * FROM cases WHERE title REGEXP '"+queryString+"' OR summary REGEXP '"+queryString+"'")

	c := GetCases(w, r, "SELECT * FROM cases WHERE title LIKE '%"+queryString+"%' OR summary LIKE '%"+queryString+"%'")

	p.Cases = append(p.Cases, c...)
	p.Error = errorString

	Display(w, "search", p)
}

// HandleFacebookLogin func
func HandleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		log.Fatal("Parse: ", err)
	}
	parameters := url.Values{}
	parameters.Add("client_id", oauthConf.ClientID)
	parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
	parameters.Add("redirect_uri", oauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateString)
	Url.RawQuery = parameters.Encode()
	url := Url.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleFacebookCallback func
func HandleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")

	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
		url.QueryEscape(token.AccessToken))
	if err != nil {
		fmt.Printf("Get: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ReadAll: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Printf("parseResponseBody: %s\n", string(response))

	type login struct {
		name string
		id   string
	}

	// res, _ := fb.Get("/me", fb.Params{
	// 	"fields":       "first_name",
	// 	"access_token": "a-valid-access-token",
	// })
	//
	// log.Printf("%+v", res["email"])

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// ChallengeHandler function creates a template for the case with the given id
func ChallengeHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{}

	session, err := store.Get(r, "lawlipops")
	CheckError(w, err, "")

	val := session.Values["userLoggedIn"]
	loggedIn, ok := val.(bool)
	if !ok {
		log.Println("Error getting userLoggedIn value")
	}

	if loggedIn {
		val = session.Values["currentUser"]
		currentUser := &User{}
		currentUser, ok = val.(*User)
		if !ok {
			log.Println("Error getting current user")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		p.CurrentUser = *currentUser

		// Get updated account information from the server
		dbmap.SelectOne(&p.CurrentUser, "SELECT * FROM accounts WHERE id=?", p.CurrentUser.ID)
		session.Values["currentUser"] = p.CurrentUser
		session.Save(r, w)
	}

	p.UserLoggedIn = loggedIn

	errorString := ""

	if r.Method == post {
		// if r.FormValue("affirm") != "" {
		// if p.UserLoggedIn {
		// 	id, _ := strconv.ParseInt(strings.Split(r.FormValue("affirm"), "-")[1], 10, 64)
		// 	var existingVote string
		// 	dbmap.SelectOne(&existingVote, "SELECT id FROM `votes` WHERE account_id=? AND case_id=?", p.CurrentUser.ID, id)
		//
		// 	if existingVote == "" {
		// 		dbmap.Exec("INSERT INTO `votes` (id, account_id, case_id, user_decision, final_decision) VALUES (NULL,?,?,?,?)", p.CurrentUser.ID, id, 1, 0)
		// 		p.Message = "Vote registered successfuly"
		// 	} else {
		// 		dbmap.Exec("UPDATE `votes` SET user_decision=? WHERE account_id=? AND case_id=?", 1, p.CurrentUser.ID, id)
		// 		p.Message = "Vote updated successfuly"
		// 	}
		// } else {
		// 	errorString = "You must be logged in to vote"
		// }
		// } else if r.FormValue("reverse") != "" {
		// if p.UserLoggedIn {
		// 	id, _ := strconv.ParseInt(strings.Split(r.FormValue("reverse"), "-")[1], 10, 64)
		// 	var existingVote string
		// 	dbmap.SelectOne(&existingVote, "SELECT id FROM `votes` WHERE account_id=? AND case_id=?", p.CurrentUser.ID, id)
		//
		// 	if existingVote == "" {
		// 		dbmap.Exec("INSERT INTO `votes` (id, account_id, case_id, user_decision, final_decision) VALUES (NULL,?,?,?,?)", p.CurrentUser.ID, id, 2, 0)
		// 		p.Message = "Vote registered successfuly"
		// 	} else {
		// 		dbmap.Exec("UPDATE `votes` SET user_decision=? WHERE account_id=? AND case_id=?", 2, p.CurrentUser.ID, id)
		// 		p.Message = "Vote updated successfuly"
		// 	}
		// } else {
		// 	errorString = "You must be logged in to vote"
		// }
		// }
	}

	challengeID := mux.Vars(r)["id"]
	challengeToDisplay := GetChallenge(w, r, challengeID)

	p.Challenges = append(p.Challenges, challengeToDisplay)
	p.Error = errorString

	Display(w, "challenge", p)
}
