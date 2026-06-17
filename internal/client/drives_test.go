package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDrives_Create(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, err := testClient(srv.URL).MakeRequest("/v1.0/drives", "POST", `{"name":"S","driveType":"project"}`, nil)
	if err != nil { t.Fatal(err) }
	assertContains(t, r.Body, "S", "Create")
}

func TestDrives_Create_MissingName(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/drives", "POST", `{"driveType":"project"}`, nil)
	if err == nil || !strings.Contains(err.Error(), "400") { t.Fatalf("expected 400: %v", err) }
}

func TestDrives_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/drives/d1", "GET", "", nil)
	assertContains(t, r.Body, "TestDrive", "Get")
}

func TestDrives_Get_404(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/drives/missing", "GET", "", nil)
	if err == nil || !strings.Contains(err.Error(), "404") { t.Fatalf("expected 404: %v", err) }
}

func TestDrives_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/drives", "GET", "", nil)
	assertContains(t, r.Body, "D1", "List")
	assertContains(t, r.Body, "D2", "List")
}

func TestDrives_ListBeta(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/me/drives", "GET", "", nil)
	assertContains(t, r.Body, "BetaDrive", "Beta")
}

func TestDrives_Update(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/drives/d1", "PATCH", `{"name":"Updated"}`, nil)
	assertContains(t, r.Body, "Updated", "Update")
}

func TestDrives_Delete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/drives/d1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "Delete")
}

func TestDrives_GetPersonal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handleDriveOK)); defer srv.Close()
	d, err := testClient(srv.URL).GetPersonalDrive()
	if err != nil { t.Fatal(err) }
	if d.ID != "d1!r1" { t.Errorf("ID=%q", d.ID) }
	if d.Name != "TestDrive" { t.Errorf("Name=%q", d.Name) }
}

func TestDrives_GetPersonal_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })); defer srv.Close()
	_, err := testClient(srv.URL).GetPersonalDrive()
	if err == nil { t.Fatal("expected error") }
}

func TestDrives_GetPersonal_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "nope") })); defer srv.Close()
	_, err := testClient(srv.URL).GetPersonalDrive()
	if err == nil { t.Fatal("expected error") }
}

func TestDrives_NotExistBeta(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives", "GET", "", nil)
	assertContains(t, r.Body, "BetaDrive", "DrivesBeta")
}
