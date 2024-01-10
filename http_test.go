package fs

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticContentHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/http.go", nil)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile("./http.go")
	if err != nil {
		t.Fatal(err)
	}

	h := StaticContentHandler(contents, "text/plain")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if ctype := res.Header().Get("Content-Type"); ctype != "text/x-go; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ctype, "text/x-go; charset=utf-8")
	}

	body := res.Body.String()
	if !strings.HasPrefix(body, "package fs") {
		t.Errorf("handler returned wrong contents, got %v", body)
	}
}

func TestDirHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/http.go", nil)
	if err != nil {
		t.Fatal(err)
	}
	h := DirHandler("./", "")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if ctype := res.Header().Get("Content-Type"); ctype != "text/x-go; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ctype, "text/x-go; charset=utf-8")
	}

	body := res.Body.String()
	if !strings.HasPrefix(body, "package fs") {
		t.Errorf("handler returned wrong contents")
	}
}

// TestFaviconHandler will test the FaviconHandler which calls the StaticContentHandler too
func TestFaviconHandler(t *testing.T) {
	favPath := "./testfiles/old_iris_favicon.ico"

	req, err := http.NewRequest("GET", "/favicon.ico", nil)
	if err != nil {
		t.Fatal(err)
	}
	h := FaviconHandler(favPath)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if ctype := res.Header().Get("Content-Type"); ctype != "image/vnd.microsoft.icon; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ctype, "image/vnd.microsoft.icon; charset=utf-8")
	}

	body := res.Body.Bytes()
	favContents, err := os.ReadFile(favPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, favContents) {
		t.Errorf("handler returned wrong contents")
	}
}

// TestSendStaticFileHandler will test the SendStaticFileHandler which calls the StaticFileHandler too
func TestSendStaticFileHandler(t *testing.T) {
	sendFile := "./testfiles/first.zip"

	req, err := http.NewRequest("GET", "/first.zip", nil)
	if err != nil {
		t.Fatal(err)
	}
	h := SendStaticFileHandler(sendFile)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if ctype := res.Header().Get("Content-Type"); ctype != "application/zip; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ctype, "application/zip; charset=utf-8")
	}

	// get the filename only, no the abs path
	_, filename := filepath.Split(sendFile)

	if attachment := res.Header().Get(contentDisposition); attachment != "attachment;filename="+filename {
		t.Errorf("handler returned wrong attachment: got %v want %v",
			attachment, "attachment;filename="+filename)
	}

	body := res.Body.Bytes()
	fileContents, err := os.ReadFile(sendFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, fileContents) {
		t.Errorf("handler returned wrong contents")
	}
}
