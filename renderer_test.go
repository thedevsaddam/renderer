package renderer

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================
// JSON Tests
// ============================================================

func TestJSONRender(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := JSON(200, map[string]string{"hello": "world"})

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"hello":"world"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestJSONIndent(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := JSON(200, map[string]string{"a": "b"}).Indent()

	_ = resp.Render(rec)

	if !strings.Contains(rec.Body.String(), "\n") {
		t.Fatalf("expected indented JSON")
	}
}

// ============================================================
// TEXT Tests
// ============================================================

func TestTextRender(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := Text(200, "hello")

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.String() != "hello" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

// ============================================================
// XML Tests
// ============================================================

type testXML struct {
	Name string `xml:"name"`
}

func TestXMLRender(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := XML(200, testXML{Name: "test"})

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(rec.Body.String(), "<name>test</name>") {
		t.Fatalf("unexpected xml: %s", rec.Body.String())
	}
}

// ============================================================
// BYTES Tests
// ============================================================

func TestBytesRender(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := Bytes(200, []byte("raw-data"), "application/octet-stream")

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.String() != "raw-data" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

// ============================================================
// FILE Tests
// ============================================================

func TestFileRender(t *testing.T) {
	rec := httptest.NewRecorder()

	content := "file-content"
	reader := strings.NewReader(content)

	resp := File(200, reader, "test.txt")

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.String() != content {
		t.Fatalf("unexpected file body: %s", rec.Body.String())
	}

	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") {
		t.Fatalf("expected attachment disposition")
	}
}

func TestFileInline(t *testing.T) {
	rec := httptest.NewRecorder()

	reader := strings.NewReader("data")

	resp := File(200, reader, "a.txt").Inline()

	_ = resp.Render(rec)

	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "inline") {
		t.Fatalf("expected inline disposition")
	}
}

// ============================================================
// NO CONTENT
// ============================================================

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := NoContent()

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body")
	}
}

// ============================================================
// HEADER TESTS
// ============================================================

func TestCustomHeader(t *testing.T) {
	rec := httptest.NewRecorder()

	resp := JSON(200, map[string]string{"x": "y"}).
		Header("X-Test", "true")

	_ = resp.Render(rec)

	if rec.Header().Get("X-Test") != "true" {
		t.Fatalf("header not set")
	}
}

// ============================================================
// HANDLER TESTS
// ============================================================

func TestHandleSuccess(t *testing.T) {
	handler := Handle(func(r *http.Request) (Response, error) {
		return JSON(200, map[string]string{"ok": "yes"}), nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestHandleError(t *testing.T) {
	handler := Handle(func(r *http.Request) (Response, error) {
		return nil, io.EOF
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected 500 got %d", rec.Code)
	}
}

func TestHandleNilResponse(t *testing.T) {
	handler := Handle(func(r *http.Request) (Response, error) {
		return nil, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected 500 got %d", rec.Code)
	}
}

// ============================================================
// PANIC SAFETY TEST
// ============================================================

type panicResponse struct{}

func (p panicResponse) Render(w http.ResponseWriter) error {
	panic("boom")
}

func TestPanicSafety(t *testing.T) {
	handler := Handle(func(r *http.Request) (Response, error) {
		return panicResponse{}, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected 500 on panic got %d", rec.Code)
	}
}

// ============================================================
// LARGE PAYLOAD TEST
// ============================================================

func TestLargeJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	data := make(map[string]string)
	for i := 0; i < 1000; i++ {
		data[string(rune(i))] = "value"
	}

	resp := JSON(200, data)

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.Len() == 0 {
		t.Fatalf("expected non-empty body")
	}
}

// ============================================================
// STREAM TEST
// ============================================================

func TestFileStreaming(t *testing.T) {
	rec := httptest.NewRecorder()

	data := strings.Repeat("A", 1024*10)
	expectedLen := len(data)

	reader := strings.NewReader(data)

	resp := File(200, reader, "big.txt")

	err := resp.Render(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Body.Len() != expectedLen {
		t.Fatalf("stream mismatch: got %d want %d", rec.Body.Len(), expectedLen)
	}
}

func BenchmarkJSON(b *testing.B) {
	rec := httptest.NewRecorder()

	data := map[string]string{"hello": "world"}

	for i := 0; i < b.N; i++ {
		_ = JSON(200, data).Render(rec)
	}
}
