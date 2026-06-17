package client

import (
	"testing"
)

func TestFiles_CreateItem(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/root/children", "POST", `{"name":"newdir"}`, nil)
	assertContains(t, r.Body, "newdir", "Create")
}

func TestFiles_CreateItem_EmptyName(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/root/children", "POST", `{"name":""}`, nil)
	if r.StatusCode < 200 || r.StatusCode >= 300 { t.Errorf("status=%d", r.StatusCode) }
}

func TestFiles_GetItem(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1", "GET", "", nil)
	assertContains(t, r.Body, "file.txt", "Get")
}

func TestFiles_GetContent(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/content", "GET", "", nil)
	_ = err
}

func TestFiles_UpdateItem(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1", "PATCH", `{"name":"updated.txt"}`, nil)
	assertContains(t, r.Body, "updated.txt", "Update")
}

func TestFiles_DeleteItem(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1", "DELETE", "", nil)
	assertStatus(t, r.StatusCode, 204, "Delete")
}

func TestFiles_Follow(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/follow", "POST", "", nil)
	assertStatus(t, r.StatusCode, 204, "Follow")
}

func TestFiles_Unfollow(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/unfollow", "POST", "", nil)
	assertStatus(t, r.StatusCode, 204, "Unfollow")
}

func TestFiles_SharedByMe(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/sharedByMe", "GET", "", nil)
	assertContains(t, r.Body, "shared.txt", "SharedByMe")
}

func TestFiles_SharedWithMe(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/sharedWithMe", "GET", "", nil)
	assertContains(t, r.Body, "from_other.txt", "SharedWithMe")
}

func TestFiles_HomeRoot(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/root", "GET", "", nil)
	assertContains(t, r.Body, "root", "Home")
}

func TestFiles_HomeChildren(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/root/children", "GET", "", nil)
	assertContains(t, r.Body, "file.txt", "Children")
	assertContains(t, r.Body, "folder", "Children")
}

func TestFiles_Export(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/exportPersonalData", "GET", "", nil)
	assertStatus(t, r.StatusCode, 202, "Export")
}
