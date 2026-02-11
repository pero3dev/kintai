package integrationtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MultipartFile struct {
	FileName    string
	ContentType string
	Content     []byte
}

func (e *TestEnv) DoRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	e.Router.ServeHTTP(recorder, req)
	return recorder
}

func (e *TestEnv) DoJSON(
	t testing.TB,
	method string,
	path string,
	body any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	req, err := JSONRequest(method, path, body, headers)
	if err != nil {
		t.Fatalf("failed to build JSON request: %v", err)
	}
	return e.DoRequest(req)
}

func (e *TestEnv) DoMultipart(
	t testing.TB,
	method string,
	path string,
	fields map[string]string,
	files map[string]MultipartFile,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	req, err := MultipartRequest(method, path, fields, files, headers)
	if err != nil {
		t.Fatalf("failed to build multipart request: %v", err)
	}
	return e.DoRequest(req)
}

func (e *TestEnv) DoDownload(
	t testing.TB,
	method string,
	path string,
	headers map[string]string,
) ([]byte, http.Header, int) {
	t.Helper()

	req, err := http.NewRequest(method, normalizePath(path), nil)
	if err != nil {
		t.Fatalf("failed to build download request: %v", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp := e.DoRequest(req)
	return resp.Body.Bytes(), resp.Header(), resp.Code
}

func JSONRequest(method string, path string, body any, headers map[string]string) (*http.Request, error) {
	var reader io.Reader

	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal JSON body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, normalizePath(path), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func MultipartRequest(
	method string,
	path string,
	fields map[string]string,
	files map[string]MultipartFile,
	headers map[string]string,
) (*http.Request, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("write field %s: %w", key, err)
		}
	}

	for fieldName, file := range files {
		name := file.FileName
		if name == "" {
			name = "upload.bin"
		}

		part, err := writer.CreateFormFile(fieldName, name)
		if err != nil {
			return nil, fmt.Errorf("create form file %s: %w", fieldName, err)
		}
		if _, err := part.Write(file.Content); err != nil {
			return nil, fmt.Errorf("write form file %s: %w", fieldName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(method, normalizePath(path), &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}
