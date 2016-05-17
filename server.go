package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type page struct {
	Title string
	Body  string
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>%s</h1>", "Welcome to LAWLIPOPS!")
}

func challengesHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/challenges/"):]
	body := "Nothing to see here"
	p := page{Title: title, Body: body}
	t, _ := template.ParseFiles("challenges.html")
	t.Execute(w, p)
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/challenges/", challengesHandler)
	http.ListenAndServe(":8000", nil)
}
