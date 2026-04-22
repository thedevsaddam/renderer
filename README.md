# renderer

A minimal, fast, and production-ready HTTP response renderer for Go.

It provides a clean abstraction over net/http responses with support for JSON, XML, Text, Bytes, Files, and a safe handler adapter compatible with Chi and standard library HTTP.

---

## ✨ Features

- ⚡ Zero-dependency core (stdlib only)
- 🧠 Clean Response abstraction
- 🔐 Panic-safe HTTP handler wrapper
- 📦 JSON / XML / Text / Bytes / File responses
- 📁 Streaming file support (no memory overload)
- 🔧 Header support on all responses
- 🧩 Chi / net/http compatible
- 🧪 Fully testable design

---

## 📦 Installation

```bash
go get github.com/thedevsaddam/renderer/v2
```

---

## 🚀 Quick Start

```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/thedevsaddam/renderer/v2"
)

func main() {
	r := chi.NewRouter()

	r.Get("/hello", renderer.Handle(func(r *http.Request) (renderer.Response, error) {
		return renderer.JSON(200, map[string]string{
			"message": "hello renderer v2",
		}), nil
	}))

	http.ListenAndServe(":8080", r)
}
```

---

## 📤 JSON Response

```go
renderer.JSON(200, map[string]string{
	"name": "Saddam",
})
```

### With indentation

```go
renderer.JSON(200, data).Indent()
```

### Add headers

```go
renderer.JSON(200, data).
	Header("X-App", "renderer")
```

---

## 📝 Text Response

```go
renderer.Text(200, "hello world")
```

---

## 📄 XML Response

```go
type User struct {
	Name string `xml:"name"`
}

renderer.XML(200, User{Name: "Saddam"})
```

---

## 📦 Raw Bytes

```go
renderer.Bytes(200, []byte("hello"), "application/octet-stream")
```

---

## 📁 File Streaming

### Download file

```go
file := strings.NewReader("file content")

renderer.File(200, file, "report.txt")
```

### Inline file

```go
renderer.File(200, file, "report.txt").Inline()
```

---

## ⚙️ Handler Adapter

```go
type HandlerFunc func(r *http.Request) (Response, error)
```

```go
r.Get("/ping", renderer.Handle(func(r *http.Request) (renderer.Response, error) {
	return renderer.JSON(200, map[string]any{
		"ping": "pong",
	}), nil
}))
```

---

## 💥 Panic Safety

All handlers are automatically protected from panics and return HTTP 500 safely.

---

## 🧠 Design Philosophy

- Explicit over magic
- Composition over framework lock-in
- Streaming-first design
- Minimal API surface

---

## 🧪 Testing

```bash
go test ./...
```

## ⚡ Benchmarks

```bash
go test ./... -bench=.
```

---

## 📄 License

MIT
