package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

// === shared helpers ===

func testClient(url string) *Client {
	return NewClient(url, false, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "tok"}))
}

func tmpFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	os.WriteFile(p, data, 0644)
	return p
}

func testServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func okJSON(w http.ResponseWriter, v interface{}) {
	jsonResponse(w, v)
}

func createdJSON(w http.ResponseWriter, v interface{}) {
	w.WriteHeader(201)
	jsonResponse(w, v)
}

func noContent(w http.ResponseWriter) {
	w.WriteHeader(204)
}

func badRequest(w http.ResponseWriter, msg string) {
	w.WriteHeader(400)
	fmt.Fprint(w, msg)
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(404)
}

func serverError(w http.ResponseWriter) {
	w.WriteHeader(500)
}

type driveStub struct {
	id, name, driveType string
}

var testDrive = driveStub{"d1!r1", "TestDrive", "personal"}

func handleDriveOK(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"id":"%s","name":"%s","driveType":"%s","root":{"webDavUrl":"http://%s/dav/spaces/%s"}}`,
		testDrive.id, testDrive.name, testDrive.driveType, r.Host, testDrive.id)
}

func handleValueList(w http.ResponseWriter, items string) {
	fmt.Fprintf(w, `{"value":%s}`, items)
}

func handleCreated(w http.ResponseWriter, r *http.Request, body interface{}) {
	var b map[string]interface{}
	if body != nil {
		json.NewDecoder(r.Body).Decode(&b)
	}
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(body)
}

func handleError(w http.ResponseWriter, code int, msg string) {
	http.Error(w, msg, code)
}

func readAll(r io.Reader) string {
	b, _ := io.ReadAll(r)
	return string(b)
}

func assertContains(t *testing.T, body, substr, label string) {
	t.Helper()
	if body == "" {
		t.Errorf("%s: empty body", label)
	} else if !contains(body, substr) {
		t.Errorf("%s: body=%q, want substring %q", label, trunc(body, 200), substr)
	}
}

func assertStatus(t *testing.T, got, want int, label string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: status=%d, want %d", label, got, want)
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
