package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	err := os.MkdirAll("data", 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func pageExists(title string) bool {
	_, err := os.Stat("data/" + title + ".txt")
	return err == nil
}

var (
	//go:embed templates/*
	templateFiles embed.FS
	//go:embed static/*
	staticFiles embed.FS
)

var templates = template.Must(
	template.New("").
		Funcs(template.FuncMap{"mdToHTML": mdToHTML}).
		ParseFS(templateFiles, "templates/*.html"),
)

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9 ]+)$")

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, _ *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func staticHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, r)
	}
}

func indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	}
}

func main() {
	fmt.Println("Listening on http://localhost:8080")
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/static/", staticHandler())
	http.HandleFunc("/", indexHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
