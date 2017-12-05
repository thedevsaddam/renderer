package renderer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type user struct {
	Name string
	Age  int
}

// common code macros
func checkNil(t *testing.T, err error) {
	if err != nil {
		t.Error("error should be nil")
	}
}

func checkNotNil(t *testing.T, err error) {
	if err == nil {
		t.Error("error should not be nil")
	}
}

func checkStatusOK(t *testing.T, code int) {
	if code != http.StatusOK {
		t.Error("http status code should be 200")
	}
}

func checkContentType(t *testing.T, got, want string) {
	if got != want {
		t.Errorf("content type missmatch. got: %s want: %s", got, want)
	}
}

func checkBody(t *testing.T, got, want string) {
	if got != want {
		t.Errorf("unexpected body: got %v want %v", got, want)
	}
}

func Test_DisableCharset(t *testing.T) {
	r := New()
	r.DisableCharset(true)
	if !r.opts.DisableCharset {
		t.Error("EnableCharset failed")
	}

	if r.opts.ContentJSON != fmt.Sprintf("%s; charset=%s", ContentJSON, defaultCharSet) {
		t.Log(r.opts.ContentJSON)
		t.Error("charset Failed")
	}

	r.DisableCharset(false)
}

func Test_DisableCharset_new(t *testing.T) {
	r := New(Options{DisableCharset: true})
	if !r.opts.DisableCharset {
		t.Error("EnableCharset on new instance failed")
	}
}

func Test_JSONIndent(t *testing.T) {
	r := New()
	r.JSONIndent(true)
	if !r.opts.JSONIndent {
		t.Error("JSONIndent failed")
	}
}

func Test_XMLIndent(t *testing.T) {
	r := New()
	r.XMLIndent(true)
	if !r.opts.XMLIndent {
		t.Error("XMLIndent failed")
	}
}

func Test_Charset(t *testing.T) {
	r := New()
	r.Charset("UTF-8")
	if r.opts.Charset != "UTF-8" {
		t.Error("Charset failed")
	}
}

func Test_EscapeHTML(t *testing.T) {
	r := New()
	r.EscapeHTML(true)
	if !r.opts.UnEscapeHTML {
		t.Error("EscapeHTML failed")
	}
}

func Test_Delims(t *testing.T) {
	r := New()
	r.Delims("[[", "]]")
	if r.opts.LeftDelim != "[[" && r.opts.RightDelim != "]]" {
		t.Error("Delims failed")
	}
}

func Test_NoContent(t *testing.T) {
	r := New()

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.NoContent(w)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/no-content", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)

	if res.Code != http.StatusNoContent {
		t.Error("error code missmatch")
	}
}

func Test_Render(t *testing.T) {
	r := New()

	var err error
	usr := user{
		"John Doe",
		30,
	}
	input, _ := json.Marshal(usr)
	expected := `{"Name":"John Doe","Age":30}`

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set(ContentType, ContentJSON)
		err = r.Render(w, http.StatusOK, input)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/render", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentJSON)
	checkBody(t, res.Body.String(), expected)
}

func Test_String(t *testing.T) {
	r := New()
	var err error

	data := "Simple string"
	expected := `Simple string`

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.String(w, http.StatusOK, data)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/render", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentText+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_json(t *testing.T) {
	r := New()
	var err error
	var bs []byte
	user := struct {
		Name string
		Age  int
	}{
		"John Doe",
		30,
	}
	expected := `{"Name":"John Doe","Age":30}`

	bs, err = r.json(user)

	checkNil(t, err)
	checkBody(t, string(bs), expected)
}

func Test_json_indent_unescapehtml(t *testing.T) {
	r := New(Options{
		JSONIndent:   true,
		UnEscapeHTML: true,
	})
	var err error
	var bs []byte
	usr := user{
		"John Doe",
		30,
	}
	expected := "{\n \"Name\": \"John Doe\",\n \"Age\": 30\n}"
	bs, err = r.json(usr)

	checkNil(t, err)
	checkBody(t, string(bs), expected)
}

func Test_JSON_prefix(t *testing.T) {
	r := New(
		Options{
			JSONPrefix: "\n",
		},
	)
	var err error

	usr := user{"John Doe", 30}
	expected := "\n{\"Name\":\"John Doe\",\"Age\":30}"

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.JSON(w, http.StatusOK, usr)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/json", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentJSON+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_JSONP(t *testing.T) {
	r := New(
		Options{
			JSONPrefix: "\n",
		},
	)
	var err error

	usr := user{"John Doe", 30}
	expected := "jsonp({\"Name\":\"John Doe\",\"Age\":30});"

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.JSONP(w, http.StatusOK, "jsonp", usr)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jsonp", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentJSONP+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_JSONP_without_callback(t *testing.T) {
	r := New(
		Options{
			JSONPrefix: "\n",
		},
	)
	var err error

	usr := user{"John Doe", 30}

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.JSONP(w, http.StatusOK, "", usr)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jsonp", nil)
	h.ServeHTTP(res, req)

	checkNotNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentJSONP+"; charset="+defaultCharSet)
}

func Test_XML(t *testing.T) {
	r := New()
	var err error

	usr := user{"John Doe", 30}
	expected := `<?xml version="1.0" encoding="ISO-8859-1" ?>\n<user><Name>John Doe</Name><Age>30</Age></user>`

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.XML(w, http.StatusOK, usr)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/xml", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentXML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_XML_indent_prefix(t *testing.T) {
	r := New(
		Options{
			XMLIndent: true,
			XMLPrefix: " ",
		},
	)
	var err error

	usr := user{"John Doe", 30}
	expected := " <user>\n <Name>John Doe</Name>\n <Age>30</Age>\n</user>"

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.XML(w, http.StatusOK, usr)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/xml", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentXML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_YAML(t *testing.T) {
	r := New()
	var err error

	usr := user{"John Doe", 30}
	expected := "name: John Doe\nage: 30\n"

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.YAML(w, http.StatusOK, usr)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/yaml", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentYAML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_HTMLString(t *testing.T) {
	r := New()
	var err error

	data := "<h1>Hello John</h1>"
	expected := `<h1>Hello John</h1>`

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.HTMLString(w, http.StatusOK, data)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/htmlstring", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_HTML(t *testing.T) {
	var err error
	dir := "htmls"
	perm := os.ModePerm
	//create tmp html template directory for parsing
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, perm)
	}
	defer os.RemoveAll(dir)
	header := `{{define "header"}}<head><title>Header</title></head>{{end}}`
	ioutil.WriteFile(dir+"/header.tmpl", []byte(header), perm)
	index := `{{define "homePage"}}<html>{{template "header"}}home</html>{{end}}`
	ioutil.WriteFile(dir+"/index.tmpl", []byte(index), perm)
	r := New(
		Options{
			ParseGlobPattern: dir + "/*.tmpl",
		},
	)

	expected := `<html><head><title>Header</title></head>home</html>`

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.HTML(w, http.StatusOK, "homePage", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/html", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_HTML_without_name_and_debug(t *testing.T) {
	var err error
	dir := "htmls"
	perm := os.ModePerm
	//create tmp html template directory for parsing
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, perm)
	}
	defer os.RemoveAll(dir)
	header := `{{define "header"}}<head><title>Header</title></head>{{end}}`
	ioutil.WriteFile(dir+"/header.tmpl", []byte(header), perm)
	index := `{{define "homePage"}}<html>{{template "header"}}home</html>{{end}}`
	ioutil.WriteFile(dir+"/index.tmpl", []byte(index), perm)
	r := New(
		Options{
			ParseGlobPattern: dir + "/*.tmpl",
			Debug:            true,
		},
	)

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.HTML(w, http.StatusOK, "", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/html", nil)
	h.ServeHTTP(res, req)

	checkNotNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
}

func Test_HTML_invalid_name(t *testing.T) {
	var err error
	dir := "htmls"
	perm := os.ModePerm
	//create tmp html template directory for parsing
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, perm)
	}
	defer os.RemoveAll(dir)
	header := `{{define "header"}}<head><title>Header</title></head>{{end}}`
	ioutil.WriteFile(dir+"/header.tmpl", []byte(header), perm)
	index := `{{define "homePage"}}<html>{{template "header"}}home</html>{{end}}`
	ioutil.WriteFile(dir+"/index.tmpl", []byte(index), perm)
	r := New(
		Options{
			ParseGlobPattern: dir + "/*.tmpl",
			Debug:            true,
		},
	)

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.HTML(w, http.StatusOK, "about", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/html", nil)
	h.ServeHTTP(res, req)

	checkNotNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
}

func Test_Template(t *testing.T) {
	var err error
	dir := "templates"
	perm := os.ModePerm
	//create tmp html template directory for parsing
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, perm)
	}
	defer os.RemoveAll(dir)
	index := `{{ define "title" }}An example layout{{ end }}{{ define "content" }}<h1>Hello {{.name | toUpper }}</h1>{{ end }}`
	ioutil.WriteFile(dir+"/index.tmpl", []byte(index), perm)
	layout := `<html><head><title>{{ template "title" . }}</title></head><body>{{ template "content" . }}</body></html>`
	ioutil.WriteFile(dir+"/layout.tmpl", []byte(layout), perm)
	r := New()

	expected := `<html><head><title>An example layout</title></head><body><h1>Hello JOHN DOE</h1></body></html>`

	toUpper := func(s string) string {
		return strings.ToUpper(s)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.FuncMap(template.FuncMap{
			"toUpper": toUpper,
		})
		err = r.Template(w, http.StatusOK, []string{"templates/layout.tmpl", "templates/index.tmpl"}, map[string]string{"name": "john doe"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/template", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_View(t *testing.T) {
	var err error
	dir := "view"
	perm := os.ModePerm
	//create tmp html template directory for parsing
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, perm)
	}
	defer os.RemoveAll(dir)
	home := `{{define "title"}}Home{{end}}{{define "content"}}<h3>Home page</h3><ul><li><a href="/">Home</a></li><li><a href="/about">About Me</a></li></ul><p>Lorem ipsum dolor sit amet</p>{{end}}`
	ioutil.WriteFile(dir+"/home.tpl", []byte(home), perm)
	about := `{{define "title"}}About Me{{end}}{{define "content"}}<h2>This is About me page.</h2><ul>Lorem ipsum dolor sit amet, consectetur adipisicing elit,</ul><p><a href="/">Home</a></p>{{end}}`
	ioutil.WriteFile(dir+"/about.tpl", []byte(about), perm)
	base := `<html><head><title>{{block "title" .}} {{end}}</title></head><body>{{ template "content" . }}</body></html>`
	ioutil.WriteFile(dir+"/base.lout", []byte(base), perm)

	r := New(
		Options{
			TemplateDir: "view",
			Debug:       true,
		},
	)

	expected := `<html><head><title>Home</title></head><body><h3>Home page</h3><ul><li><a href="/">Home</a></li><li><a href="/about">About Me</a></li></ul><p>Lorem ipsum dolor sit amet</p></body></html>`

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.View(w, http.StatusOK, "home", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/template", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
	checkBody(t, res.Body.String(), expected)
}

func Test_View_invalid_name(t *testing.T) {
	var err error
	dir := "view"
	perm := os.ModePerm
	//create tmp html template directory for parsing
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, perm)
	}
	defer os.RemoveAll(dir)
	home := `{{define "title"}}Home{{end}}{{define "content"}}<h3>Home page</h3><ul><li><a href="/">Home</a></li><li><a href="/about">About Me</a></li></ul><p>Lorem ipsum dolor sit amet</p>{{end}}`
	ioutil.WriteFile(dir+"/home.tpl", []byte(home), perm)
	base := `<html><head><title>{{block "title" .}} {{end}}</title></head><body>{{ template "content" . }}</body></html>`
	ioutil.WriteFile(dir+"/base.lout", []byte(base), perm)

	r := New(
		Options{
			TemplateDir: "view",
			Debug:       true,
		},
	)

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.View(w, http.StatusOK, "invalid template", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/template", nil)
	h.ServeHTTP(res, req)

	checkNotNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), ContentHTML+"; charset="+defaultCharSet)
}

func Test_Binary_inline(t *testing.T) {
	var err error
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		file := strings.NewReader("This is a long binary data")
		err = r.Binary(w, http.StatusOK, file, "abc.txt", true)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bin", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), r.opts.ContentBinary)
	checkBody(t, res.Body.String(), "This is a long binary data")
}

func Test_Binary_attachment(t *testing.T) {
	var err error
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		file := strings.NewReader("This is a long binary data")
		err = r.Binary(w, http.StatusOK, file, "abc.txt", false)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bin", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), r.opts.ContentBinary)
	checkBody(t, res.Body.String(), "This is a long binary data")
}

func Test_File_inline(t *testing.T) {
	var err error
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		file := strings.NewReader("This is a long binary data")
		err = r.File(w, http.StatusOK, file, "abc.txt", true)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file-inline", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkBody(t, res.Body.String(), "This is a long binary data")
}

func Test_File_attachment(t *testing.T) {
	var err error
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		file := strings.NewReader("This is a long binary data")
		err = r.File(w, http.StatusOK, file, "abc.txt", false)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file-attachment", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkBody(t, res.Body.String(), "This is a long binary data")
}

func Test_File_view(t *testing.T) {
	var err error
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.FileView(w, http.StatusOK, "README.md", "README.md")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file-attachment", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), r.opts.ContentText)
}

func Test_File_download(t *testing.T) {
	var err error
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err = r.FileDownload(w, http.StatusOK, "README.md", "README")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file-attachment", nil)
	h.ServeHTTP(res, req)

	checkNil(t, err)
	checkStatusOK(t, res.Code)
	checkContentType(t, res.HeaderMap.Get(ContentType), r.opts.ContentText)
}

func Benchmark_NoContent(b *testing.B) {
	r := New()
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for n := 0; n < b.N; n++ {
			r.NoContent(w)
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(res, req)
}

func Benchmark_String(b *testing.B) {
	r := New()
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for n := 0; n < b.N; n++ {
			r.String(w, http.StatusOK, "Hello John")
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(res, req)
}

func Benchmark_JSON(b *testing.B) {
	r := New()
	v := map[string]string{"name": "john doe"}
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for n := 0; n < b.N; n++ {
			r.JSON(w, http.StatusOK, v)
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(res, req)
}

func Benchmark_JSONP(b *testing.B) {
	r := New()
	v := map[string]string{"name": "john doe"}
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for n := 0; n < b.N; n++ {
			r.JSONP(w, http.StatusOK, "callback", v)
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(res, req)
}

func Benchmark_XML(b *testing.B) {
	r := New()
	v := map[string]string{"name": "john doe"}
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for n := 0; n < b.N; n++ {
			r.XML(w, http.StatusOK, v)
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(res, req)
}

func Benchmark_YAML(b *testing.B) {
	r := New()
	v := map[string]string{"name": "john doe"}
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for n := 0; n < b.N; n++ {
			r.YAML(w, http.StatusOK, v)
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(res, req)
}
