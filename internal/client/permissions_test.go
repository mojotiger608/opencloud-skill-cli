package client

import (
	"strings"
	"testing"
)

// Permissions
func TestPerm_CreateLinkView(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/createLink", "POST", `{"type":"view"}`, nil)
	assertContains(t, r.Body, "abc", "View")
}

func TestPerm_CreateLinkEdit(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/createLink", "POST", `{"type":"edit"}`, nil)
	assertContains(t, r.Body, "abc", "Edit")
}

func TestPerm_CreateLinkPassword(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/createLink", "POST", `{"type":"view","password":"Secret123"}`, nil)
	assertContains(t, r.Body, "hasPassword", "Password")
}

func TestPerm_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/permissions", "GET", "", nil)
	assertContains(t, r.Body, "p1", "List")
}

func TestPerm_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/permissions/p1", "GET", "", nil)
	assertContains(t, r.Body, "p1", "Get")
}

func TestPerm_Update(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/permissions/p1", "PATCH", `{"roles":["writer"]}`, nil)
	assertContains(t, r.Body, "p1", "Update")
}

func TestPerm_Delete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/permissions/p1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "Delete")
}

func TestPerm_SetPassword(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/permissions/p1/setPassword", "POST", `{"password":"New"}`, nil)
	assertStatus(t, r.StatusCode, 200, "SetPassword")
}

func TestPerm_Invite(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/items/f1/invite", "POST",
		`{"@libre.graph.recipient.type":"user","objectId":"u1","roles":["reader"]}`, nil)
	assertContains(t, r.Body, "p3", "Invite")
}

func TestPerm_Invite_MissingType(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/items/f1/invite", "POST", `{"objectId":"u1"}`, nil)
	if err == nil || !strings.Contains(err.Error(), "400") { t.Fatalf("expected 400: %v", err) }
}

func TestPerm_AddMember(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/invite", "POST",
		`{"@libre.graph.recipient.type":"user","objectId":"u1","roles":["reader"]}`, nil)
	assertContains(t, r.Body, "p3", "AddMember")
}

func TestPerm_DeleteMember(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/members/p1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "DeleteMember")
}

func TestPerm_SpaceRootList(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/permissions", "GET", "", nil)
	assertContains(t, r.Body, "p1", "SpaceRoot")
}

func TestPerm_SpaceRootCreateLink(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/createLink", "POST", `{"type":"view"}`, nil)
	assertContains(t, r.Body, "abc", "SpaceRootLink")
}

func TestPerm_SpaceRootDelete(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/permissions/p1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "SpaceRootDel")
}

func TestPerm_SpaceRootSetPassword(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/permissions/p1/setPassword", "POST", "", nil)
	assertStatus(t, r.StatusCode, 200, "SpaceRootPwd")
}

func TestPerm_SpaceRootUpdate(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/root/permissions/p1", "PATCH", "", nil)
	assertContains(t, r.Body, "p1", "SpaceRootUpd")
}
