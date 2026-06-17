package client

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestError_GetPersonalDrive_401(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
	defer srv.Close()
	_, err := testClient(srv.URL).GetPersonalDrive()
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "GetPersonalDrive") { t.Errorf("missing op prefix: %s", e) }
	if !strings.Contains(e, "401") { t.Errorf("missing status code: %s", e) }
}

func TestError_GetPersonalDrive_500(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	defer srv.Close()
	_, err := testClient(srv.URL).GetPersonalDrive()
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "GetPersonalDrive") { t.Errorf("missing op prefix: %s", e) }
	if !strings.Contains(e, "500") { t.Errorf("missing status: %s", e) }
}

func TestError_GetPersonalDrive_BadJSON(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("not json")) })
	defer srv.Close()
	_, err := testClient(srv.URL).GetPersonalDrive()
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "decode") { t.Errorf("expected decode error: %s", err) }
}

func TestError_UploadFile_NotFound(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		w.WriteHeader(404)
	})
	defer srv.Close()
	c := testClient(srv.URL)
	_, err := c.UploadFile(tmpFile(t, "x.bin", []byte("x")), "", "")
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "putFile") { t.Errorf("missing putFile: %s", e) }
	if !strings.Contains(e, "404") { t.Errorf("missing 404: %s", e) }
}

func TestError_UploadFile_Forbidden(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		w.WriteHeader(403)
	})
	defer srv.Close()
	_, err := testClient(srv.URL).UploadFile(tmpFile(t, "f.bin", []byte("x")), "", "")
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "403") { t.Errorf("missing 403: %s", err) }
}

func TestError_UploadFile_ReadFail(t *testing.T) {
	c := NewClient("http://[::1]:1", false, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "tok"}))
	_, err := c.UploadFile("/nonexistent/file.bin", "", "")
	if err == nil { t.Fatal("expected error") }
	// GetPersonalDrive fails first because the server is unreachable
	if !strings.Contains(err.Error(), "GetPersonalDrive") { t.Errorf("expected chain: %s", err) }
}

func TestError_UploadFileTUS_StatFail(t *testing.T) {
	_, err := testClient("http://[::1]:1").UploadFileTUS("/nope.bin", "", 1024)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "GetPersonalDrive") { t.Errorf("expected chain: %s", err) }
}

func TestError_TusCreate_400(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400) })
	defer srv.Close()
	_, err := testClient(srv.URL).tusCreate(srv.URL+"/", "t.bin", 100)
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "tusCreate") { t.Errorf("missing tusCreate: %s", e) }
	if !strings.Contains(e, "400") { t.Errorf("missing 400: %s", e) }
}

func TestError_TusCreate_500(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	defer srv.Close()
	_, err := testClient(srv.URL).tusCreate(srv.URL+"/", "t.bin", 100)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "500") { t.Errorf("missing 500: %s", err) }
}

func TestError_TusCreate_NoDir(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	_, err := testClient(srv.URL).tusCreate(srv.URL+"/file.bin", "t.bin", 100)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "412") { t.Errorf("expected 412 for non-dir: %s", err) }
}

func TestError_TusPatch_409(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "w.bin", 100)
	s.patch([]byte("AAAA")); s.Offset += 4
	s.Offset = 0 // wrong offset
	err := s.patch([]byte("bad"))
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "tusPatch") { t.Errorf("missing tusPatch: %s", e) }
	if !strings.Contains(e, "409") { t.Errorf("missing 409: %s", e) }
	if !strings.Contains(e, "conflict") { t.Errorf("missing conflict: %s", e) }
}

func TestError_TusPatch_500(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "e.bin", 100)
	s.patch([]byte("A")); s.Offset += 1
	srv.Close() // force network error
	err := s.patch([]byte("B"))
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "tusPatch") { t.Errorf("missing tusPatch: %s", err) }
}

func TestError_TusHead_NotFound(t *testing.T) {
	s := &TUSSession{uploadURL: "http://x/tus/nope", client: testClient("http://x")}
	_, err := s.head()
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "tusHead") { t.Errorf("missing tusHead: %s", err) }
}

func TestError_TusHead_MissingHeader(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Tus-Resumable", "1.0.0")
		w.WriteHeader(200)
	})
	defer srv.Close()
	s := &TUSSession{uploadURL: srv.URL, client: testClient(srv.URL)}
	_, err := s.head()
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "tusHead") { t.Errorf("missing tusHead: %s", e) }
	if !strings.Contains(e, "Upload-Offset") { t.Errorf("missing header name: %s", e) }
}

func TestError_TusHead_BadOffset(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Tus-Resumable", "1.0.0")
		w.Header().Set("Upload-Offset", "not-a-number")
		w.WriteHeader(200)
	})
	defer srv.Close()
	s := &TUSSession{uploadURL: srv.URL, client: testClient(srv.URL)}
	_, err := s.head()
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "invalid") { t.Errorf("expected invalid: %s", err) }
}

func TestError_MakeRequest_401(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
	defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "MakeRequest") { t.Errorf("missing MakeRequest: %s", e) }
	if !strings.Contains(e, "401") { t.Errorf("missing 401: %s", e) }
}

func TestError_MakeRequest_403(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) })
	defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "403") { t.Errorf("missing 403: %s", err) }
}

func TestError_MakeRequest_404(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) })
	defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/drives/missing", "GET", "", nil)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "404") { t.Errorf("missing 404: %s", err) }
}

func TestError_MakeRequest_500(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "500") { t.Errorf("missing 500: %s", err) }
}

func TestError_MakeRequest_Drive404(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/drives/missing", "GET", "", nil)
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "MakeRequest") { t.Errorf("missing MakeRequest: %s", e) }
	if !strings.Contains(e, "404") { t.Errorf("missing 404: %s", e) }
}

func TestError_MakeRequest_NoToken(t *testing.T) {
	c := NewClient("http://x", false, nil)
	_, err := c.MakeRequest("/v1.0/me", "GET", "", nil)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "token") { t.Errorf("expected token error: %s", err) }
}

func TestError_Invite_MissingType(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1beta1/drives/d1/items/f1/invite", "POST", `{"objectId":"u1"}`, nil)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "400") { t.Errorf("expected 400: %s", err) }
}

func TestError_CreateDrive_MissingName(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/drives", "POST", `{"driveType":"project"}`, nil)
	if err == nil { t.Fatal("expected error") }
	if !strings.Contains(err.Error(), "400") { t.Errorf("expected 400: %s", err) }
}

func TestError_CreateUser_MissingUPN(t *testing.T) {
	srv := testServer(muxHandler); defer srv.Close()
	_, err := testClient(srv.URL).MakeRequest("/v1.0/users", "POST", `{"displayName":"N"}`, nil)
	if err == nil { t.Fatal("expected error") }
	e := err.Error()
	if !strings.Contains(e, "400") { t.Errorf("expected 400: %s", e) }
}

func TestError_ErrorFormatting_HasOperationPrefix(t *testing.T) {
	tests := []struct {
		fn    func() error
		label string
		want  string
	}{
		{func() error {
			srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
			defer srv.Close()
			_, err := testClient(srv.URL).GetPersonalDrive()
			return err
		}, "GetPersonalDrive", "GetPersonalDrive"},
		{func() error {
			srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400) })
			defer srv.Close()
			_, err := testClient(srv.URL).tusCreate(srv.URL+"/", "t.bin", 100)
			return err
		}, "tusCreate", "tusCreate"},
		{func() error {
			srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
			defer srv.Close()
			_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
			return err
		}, "MakeRequest", "MakeRequest"},
	}

	for _, tt := range tests {
		err := tt.fn()
		if err == nil { t.Errorf("%s: expected error", tt.label); continue }
		if !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: error %q missing prefix %q", tt.label, err.Error(), tt.want)
		}
	}
}

func TestError_AllHTTPStatuses(t *testing.T) {
	codes := []int{400, 401, 403, 404, 405, 409, 413, 429, 500, 502, 503}
	prefixes := []string{"GetPersonalDrive", "MakeRequest", "tusCreate", "putFile", "tusPatch"}

	for _, code := range codes {
		srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code) })
		_, err := testClient(srv.URL).MakeRequest("/v1.0/me", "GET", "", nil)
		srv.Close()
		if err == nil { t.Errorf("code %d: expected error", code); continue }
		e := err.Error()
		hasCode := strings.Contains(e, fmt.Sprintf("%d", code))
		hasPrefix := false
		for _, p := range prefixes {
			if strings.Contains(e, p) { hasPrefix = true; break }
		}
		if !hasCode { t.Errorf("code %d: missing status in %s", code, e) }
		if !hasPrefix { t.Errorf("code %d: missing op prefix in %s", code, e) }
	}

	_ = prefixes // satisfy unused
}

func TestError_AllOutputFormats(t *testing.T) {
	type testCase struct {
		name string
		gen  func() error
		want string
	}

	cases := []testCase{
		{"no token", func() error {
			c := NewClient("http://x", false, nil)
			_, e := c.MakeRequest("/v1.0/me", "GET", "", nil)
			return e
		}, "token"},

		{"upload dir", func() error {
			srv := testServer(func(w http.ResponseWriter, r *http.Request) {
				handleDriveOK(w, r)
			})
			defer srv.Close()
			_, e := testClient(srv.URL).UploadFile(t.TempDir(), "", "")
			return e
		}, ""},  // on Windows the error is different — just verify it fails

		{"resume offset", func() error {
			tm := newTusMock()
			srv := testServer(tm.ServeHTTP)
			defer srv.Close()
			s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "r.bin", 100)
			s.patch([]byte("X")); s.Offset += 1
			s.Offset = 0
			e := s.patch([]byte("Y"))
			return e
		}, "409"},
	}

	for _, tc := range cases {
		err := tc.gen()
		if err == nil { t.Errorf("%s: expected error", tc.name); continue }
		if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q missing %q", tc.name, err.Error(), tc.want)
		}
	}
}
