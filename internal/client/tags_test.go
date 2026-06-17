package client

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
)

func TestTags_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/tags", "GET", "", nil)
	assertContains(t, r.Body, "tag1", "List")
}

func TestTags_Assign(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/tags/assign", "POST", `{"resourceId":"tag1"}`, nil)
	assertStatus(t, r.StatusCode, 204, "Assign")
}

func TestTags_Unassign(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/items/f1/tags/unassign", "POST", `{"resourceId":"tag1"}`, nil)
	assertStatus(t, r.StatusCode, 204, "Unassign")
}

func TestActivities_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/drive/activities", "GET", "", nil)
	assertContains(t, r.Body, "a1", "Activities")
}

func TestApplications_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/applications", "GET", "", nil)
	assertContains(t, r.Body, "App1", "List")
}

func TestApplications_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/applications/app1", "GET", "", nil)
	assertContains(t, r.Body, "App1", "Get")
}

func TestInvitations_Create(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/invitations", "POST", `{"invitedUserEmailAddress":"u@t.com","inviteRedirectUrl":"http://x.com"}`, nil)
	assertContains(t, r.Body, "inv1", "Create")
}

func TestInvitations_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/invitations", "GET", "", nil)
	assertContains(t, r.Body, "inv1", "List")
}

func TestInvitations_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/me/invitations/inv1", "GET", "", nil)
	assertContains(t, r.Body, "inv1", "Get")
}

func TestRoleDefs_List(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/roleManagement/roleDefinitions", "GET", "", nil)
	assertContains(t, r.Body, "Reader", "List")
}

func TestRoleDefs_Get(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	r, _ := testClient(srv.URL).MakeRequest("/v1.0/roleManagement/roleDefinitions/r1", "GET", "", nil)
	assertContains(t, r.Body, "Reader", "Get")
}

func TestHostOverride(t *testing.T) {
	var host string
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		host = r.Host
		handleDriveOK(w, r)
	})
	defer srv.Close()
	c := testClient(srv.URL)
	c.HostOverride = "overridden.host"
	c.GetPersonalDrive()
	if host != "overridden.host" { t.Errorf("host=%q", host) }
}

func TestResolveIP(t *testing.T) {
	c := testClient("http://x")
	c.ResolveIP = "10.0.0.1"
	if c.ResolveIP != "10.0.0.1" { t.Error("ResolveIP not set") }
}

func TestMakeRequest_400(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400) }); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
	if err == nil || !strings.Contains(err.Error(), "400") { t.Fatalf("expected 400: %v", err) }
}

func TestMakeRequest_500(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
	if err == nil || !strings.Contains(err.Error(), "500") { t.Fatalf("expected 500: %v", err) }
}

func TestBase64Roundtrip(t *testing.T) {
	for _, n := range []string{"a.txt", "with spaces.txt", "unicode_日本語.txt"} {
		d, err := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString([]byte(n)))
		if err != nil { t.Errorf("decode %q: %v", n, err) }
		if string(d) != n { t.Errorf("%q != %q", n, d) }
	}
}
