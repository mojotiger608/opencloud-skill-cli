package client

import (
	"strings"
	"testing"
)

func TestUsers_Create(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users", "POST", `{"displayName":"N","userPrincipalName":"n@t.com","password":"S3cret!"}`, nil)
	assertContains(t, r.Body, "N", "Create")
}

func TestUsers_Create_MissingUPN(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/users", "POST", `{"displayName":"N"}`, nil)
	if err == nil || !strings.Contains(err.Error(), "400") { t.Fatalf("expected 400: %v", err) }
}

func TestUsers_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1", "GET", "", nil)
	assertContains(t, r.Body, "TestUser", "Get")
}

func TestUsers_Get_404(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/users/missing", "GET", "", nil)
	if err == nil || !strings.Contains(err.Error(), "404") { t.Fatalf("expected 404: %v", err) }
}

func TestUsers_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users", "GET", "", nil)
	assertContains(t, r.Body, "Alice", "List")
}

func TestUsers_Update(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1", "PATCH", `{"displayName":"Updated"}`, nil)
	assertContains(t, r.Body, "Updated", "Update")
}

func TestUsers_Delete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "Delete")
}

func TestUsers_ChangePassword(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/changepassword", "POST", `{"cur":"o","new":"N3w!"}`, nil)
	assertStatus(t, r.StatusCode, 204, "Password")
}

func TestUsers_Photo(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1/photo", "GET", "", nil)
	assertContains(t, r.Body, "FAKEJPEG", "Photo")
}

func TestUsers_PhotoDelete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/photo", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "PhotoDel")
}

func TestUsers_PhotoPut(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/photo", "PUT", "", nil)
	assertStatus(t, r.StatusCode, 200, "PhotoPut")
}

func TestUsers_PhotoPatch(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/photo", "PATCH", "", nil)
	assertStatus(t, r.StatusCode, 200, "PhotoPatch")
}

func TestUsers_Search(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	v := make(map[string][]string)
	v["$search"] = []string{"alice"}
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users", "GET", "", v)
	assertContains(t, r.Body, "Alice", "Search")
}

func TestUsers_EmptyBody(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users", "GET", "", nil)
	if r.Body == "" { t.Error("empty body") }
}

// Groups
func TestGroups_Create(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/groups", "POST", `{"displayName":"G"}`, nil)
	assertContains(t, r.Body, "G", "Create")
}

func TestGroups_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/groups/g1", "GET", "", nil)
	assertContains(t, r.Body, "TestGroup", "Get")
}

func TestGroups_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/groups", "GET", "", nil)
	assertContains(t, r.Body, "G1", "List")
}

func TestGroups_Update(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/groups/g1", "PATCH", `{"displayName":"Updated"}`, nil)
	assertContains(t, r.Body, "Updated", "Update")
}

func TestGroups_Delete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/groups/g1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "Delete")
}

// AppRole
func TestAppRole_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1/appRoleAssignments", "GET", "", nil)
	assertContains(t, r.Body, "ar1", "List")
}

func TestAppRole_Create(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1/appRoleAssignments", "POST", `{"pid":"u1","rid":"a1","aid":"r1"}`, nil)
	assertContains(t, r.Body, "ar2", "Create")
}

func TestAppRole_Delete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/users/u1/appRoleAssignments/ar1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "Delete")
}
