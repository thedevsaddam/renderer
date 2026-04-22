package renderer

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// ============================================================
// Core Interface
// ============================================================

type Response interface {
	Render(w http.ResponseWriter) error
}

// ============================================================
// Safe Writer Helper (production-safe core)
// ============================================================

func write(
	w http.ResponseWriter,
	status int,
	headers http.Header,
	contentType string,
	body func(io.Writer) error,
) error {

	// apply headers first
	for k, v := range headers {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.WriteHeader(status)

	// panic safety (prevents server crash)
	defer func() {
		if recover() != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}()

	return body(w)
}

// ============================================================
// JSON Response
// ============================================================

type JSONResponse[T any] struct {
	status int
	data   T
	header http.Header

	indent bool
}

func JSON[T any](status int, data T) *JSONResponse[T] {
	return &JSONResponse[T]{
		status: status,
		data:   data,
		header: make(http.Header),
	}
}

func (r *JSONResponse[T]) Indent() *JSONResponse[T] {
	r.indent = true
	return r
}

func (r *JSONResponse[T]) Header(key, value string) *JSONResponse[T] {
	r.header.Add(key, value)
	return r
}

func (r *JSONResponse[T]) Render(w http.ResponseWriter) error {
	return write(w, r.status, r.header, "application/json; charset=utf-8", func(out io.Writer) error {
		enc := json.NewEncoder(out)
		if r.indent {
			enc.SetIndent("", "  ")
		}
		return enc.Encode(r.data)
	})
}

// ============================================================
// TEXT Response
// ============================================================

type TextResponse struct {
	status int
	data   string
	header http.Header
}

func Text(status int, s string) *TextResponse {
	return &TextResponse{
		status: status,
		data:   s,
		header: make(http.Header),
	}
}

func (r *TextResponse) Header(key, value string) *TextResponse {
	r.header.Add(key, value)
	return r
}

func (r *TextResponse) Render(w http.ResponseWriter) error {
	return write(w, r.status, r.header, "text/plain; charset=utf-8", func(out io.Writer) error {
		_, err := out.Write([]byte(r.data))
		return err
	})
}

// ============================================================
// XML Response
// ============================================================

type XMLResponse[T any] struct {
	status int
	data   T
	header http.Header

	indent bool
}

func XML[T any](status int, data T) *XMLResponse[T] {
	return &XMLResponse[T]{
		status: status,
		data:   data,
		header: make(http.Header),
	}
}

func (r *XMLResponse[T]) Indent() *XMLResponse[T] {
	r.indent = true
	return r
}

func (r *XMLResponse[T]) Header(key, value string) *XMLResponse[T] {
	r.header.Add(key, value)
	return r
}

func (r *XMLResponse[T]) Render(w http.ResponseWriter) error {
	return write(w, r.status, r.header, "application/xml; charset=utf-8", func(out io.Writer) error {
		enc := xml.NewEncoder(out)

		if r.indent {
			enc.Indent("", "  ")
		}

		return enc.Encode(r.data)
	})
}

// ============================================================
// RAW BYTES Response
// ============================================================

type BytesResponse struct {
	status int
	data   []byte
	header http.Header
	ct     string
}

func Bytes(status int, data []byte, contentType string) *BytesResponse {
	return &BytesResponse{
		status: status,
		data:   data,
		header: make(http.Header),
		ct:     contentType,
	}
}

func (r *BytesResponse) Header(key, value string) *BytesResponse {
	r.header.Add(key, value)
	return r
}

func (r *BytesResponse) Render(w http.ResponseWriter) error {
	return write(w, r.status, r.header, r.ct, func(out io.Writer) error {
		_, err := out.Write(r.data)
		return err
	})
}

// ============================================================
// FILE Streaming Response (SAFE)
// ============================================================

type FileResponse struct {
	status int
	reader io.Reader
	name   string
	inline bool
	header http.Header
}

func File(status int, reader io.Reader, name string) *FileResponse {
	return &FileResponse{
		status: status,
		reader: reader,
		name:   name,
		header: make(http.Header),
	}
}

func (f *FileResponse) Inline() *FileResponse {
	f.inline = true
	return f
}

func (f *FileResponse) Header(key, value string) *FileResponse {
	f.header.Add(key, value)
	return f
}

func (f *FileResponse) Render(w http.ResponseWriter) error {
	mode := "attachment"
	if f.inline {
		mode = "inline"
	}

	f.header.Set("Content-Disposition", mode+"; filename="+f.name)
	f.header.Set("Content-Type", "application/octet-stream")

	return write(w, f.status, f.header, "", func(out io.Writer) error {
		_, err := io.Copy(out, f.reader)
		return err
	})
}

// ============================================================
// NO CONTENT
// ============================================================

type NoContentResponse struct {
	status int
}

func NoContent() *NoContentResponse {
	return &NoContentResponse{
		status: http.StatusNoContent,
	}
}

func (r *NoContentResponse) Render(w http.ResponseWriter) error {
	w.WriteHeader(r.status)
	return nil
}

// ============================================================
// Chi / stdlib Handler Adapter
// ============================================================

type HandlerFunc func(r *http.Request) (Response, error)

func Handle(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		resp, err := h(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp == nil {
			http.Error(w, "empty response", http.StatusInternalServerError)
			return
		}

		if err := resp.Render(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
