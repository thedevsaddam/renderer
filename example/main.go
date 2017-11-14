package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/thedevsaddam/renderer"
)

func toUpper(s string) string {
	return strings.ToUpper(s)
}

func main() {
	rnd := renderer.New(
		renderer.Options{
			ParseGlobPattern: "html/*.html",
			TemplateDir:      "view",
		},
	)

	mux := http.NewServeMux()

	usr := struct {
		Name string
		Age  int
	}{"John Doe", 30}

	// serving String as text/plain
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rnd.String(w, http.StatusOK, "Welcome to renderer!")
	})

	// serving success but no content
	mux.HandleFunc("/no-content", func(w http.ResponseWriter, r *http.Request) {
		rnd.NoContent(w)
	})

	// serving JSON
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		rnd.JSON(w, http.StatusOK, usr)
	})

	// serving JSONP
	mux.HandleFunc("/jsonp", func(w http.ResponseWriter, r *http.Request) {
		rnd.JSONP(w, http.StatusOK, "callback", usr)
	})

	// serving XML
	mux.HandleFunc("/xml", func(w http.ResponseWriter, r *http.Request) {
		rnd.XML(w, http.StatusOK, usr)
	})

	// serving YAML
	mux.HandleFunc("/yaml", func(w http.ResponseWriter, r *http.Request) {
		rnd.YAML(w, http.StatusOK, usr)
	})

	// serving File as arbitary binary data
	mux.HandleFunc("/binary", func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader
		reader, _ = os.Open("../README.md")
		rnd.Binary(w, http.StatusOK, reader, "readme.md", true)
	})

	// serving File as inline
	mux.HandleFunc("/file-inline", func(w http.ResponseWriter, r *http.Request) {
		rnd.FileView(w, http.StatusOK, "../README.md", "readme.md")
	})

	// serving File as attachment
	mux.HandleFunc("/file-download", func(w http.ResponseWriter, r *http.Request) {
		rnd.FileDownload(w, http.StatusOK, "../README.md", "readme.md")
	})

	// serving File from reader as inline
	mux.HandleFunc("/file-reader", func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader
		reader, _ = os.Open("../README.md")
		rnd.File(w, http.StatusOK, reader, "readme.md", true)
	})

	// serving custom response using render and chaining methods
	mux.HandleFunc("/render", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(renderer.ContentType, renderer.ContentText)
		rnd.Render(w, http.StatusOK, []byte("Send the message as text response"))
	})

	// When using the "Template" method you can simply pass the base layouts, templates
	// path as a slice of strings. You can parse template on the fly using
	// Template method. You can set a delimiter, and inject a FuncMap easily.
	mux.HandleFunc("/template", func(w http.ResponseWriter, r *http.Request) {
		tpls := []string{"template/layout.tmpl", "template/index.tmpl", "template/partial.tmpl"}
		rnd.FuncMap(template.FuncMap{
			"toUpper": toUpper,
		})
		err := rnd.Template(w, http.StatusOK, tpls, usr)
		if err != nil {
			log.Fatal(err) //respond with error page or message
		}
	})

	// When using "HTML" you can parse a template directory using a glob
	// pattern and then call the templates by their names.
	mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		err := rnd.HTML(w, http.StatusOK, "indexPage", usr)
		if err != nil {
			log.Fatal(err) //respond with error page or message
		}
	})

	// When using "View" for parsing template you can pass multiple layout and
	// templates. The template name will be the file name root.
	mux.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		err := rnd.View(w, http.StatusOK, "home", usr)
		if err != nil {
			log.Fatal(err) //respond with error page or message
		}
	})

	port := ":9000"
	log.Println("Listening on port", port)
	http.ListenAndServe(port, mux)
}
