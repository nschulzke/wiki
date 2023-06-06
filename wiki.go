package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Page struct {
	Title     string
	Body      []byte
	Backlinks []string
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".md"
	directory := filepath.Dir(filename)
	err := os.MkdirAll(directory, 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, p.Body, 0600)
}

type Breadcrumb struct {
	Display string
	Title   string
	Exists  bool
}

func (p *Page) Breadcrumbs() []Breadcrumb {
	if strings.Contains(p.Title, "/") == false {
		return nil
	}
	segments := strings.Split(p.Title, "/")[0 : len(strings.Split(p.Title, "/"))-1]
	var breadcrumbs []Breadcrumb
	for i, _ := range segments {
		display := segments[i]
		title := strings.Join(segments[0:i+1], "/")
		breadcrumbs = append(breadcrumbs, Breadcrumb{Title: title, Display: display, Exists: pageExists(title)})
	}
	return breadcrumbs
}

func loadPage(title string, withBacklinks bool) (*Page, error) {
	filename := "data/" + title + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var bls []string = nil
	if withBacklinks {
		bls, err = backlinks(title)
		if err != nil {
			return nil, err
		}
	}
	return &Page{Title: title, Body: body, Backlinks: bls}, nil
}

func listPages() ([]string, error) {
	var pages []string = nil
	err := filepath.WalkDir("data", func(path string, file os.DirEntry, err error) error {
		if file.IsDir() {
			return nil
		}
		bareName := strings.TrimSuffix(path, ".md")
		bareName = strings.TrimPrefix(bareName, "data/")
		pages = append(pages, bareName)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(pages, func(i, j int) bool {
		return strings.ToLower(pages[i]) < strings.ToLower(pages[j])
	})
	return pages, nil
}

func deletePage(title string) error {
	filename := "data/" + title + ".md"
	return os.Remove(filename)
}

func pageExists(title string) bool {
	_, err := os.Stat("data/" + title + ".md")
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

var validPath = regexp.MustCompile("^/(edit|save|view|delete)/([a-zA-Z0-9/ ]+)$")

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title, true)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, _ *http.Request, title string) {
	p, err := loadPage(title, true)
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

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	if r.Method == "POST" {
		err := deletePage(title)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		p, err := loadPage(title, true)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		renderTemplate(w, "confirm_delete", p)
	}
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
		pages, err := listPages()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = templates.ExecuteTemplate(w, "index.html", pages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func main() {
	fmt.Println("Listening on http://localhost:8080")
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/static/", staticHandler())
	http.HandleFunc("/", indexHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
