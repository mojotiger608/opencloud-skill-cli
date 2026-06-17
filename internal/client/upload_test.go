package client

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"
)

func TestUploadFile_Success(t *testing.T) {
	var body []byte
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("OC-FileID", "d1!new")
		w.WriteHeader(201)
	})
	defer srv.Close()
	r, err := testClient(srv.URL).UploadFile(tmpFile(t, "a.txt", []byte("hello")), "", "text/plain")
	if err != nil { t.Fatal(err) }
	if r.FileID != "d1!new" { t.Errorf("FileID=%q", r.FileID) }
	if string(body) != "hello" { t.Errorf("body=%q", body) }
}

func TestUploadFile_CustomName(t *testing.T) {
	var path string
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		path = r.URL.Path; w.WriteHeader(201)
	})
	defer srv.Close()
	testClient(srv.URL).UploadFile(tmpFile(t, "orig.txt", []byte("x")), "renamed.txt", "")
	if !strings.Contains(path, "renamed.txt") { t.Errorf("path=%q", path) }
}

func TestUploadFile_Empty(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		w.WriteHeader(201)
	})
	defer srv.Close()
	r, err := testClient(srv.URL).UploadFile(tmpFile(t, "e.bin", nil), "", "")
	if err != nil { t.Fatal(err) }
	if r.Size != 0 { t.Errorf("Size=%d", r.Size) }
}

func TestUploadFile_NotFound(t *testing.T) {
	_, err := testClient("http://x").UploadFile("/nope.bin", "", "")
	if err == nil { t.Fatal("expected error") }
}

func TestUploadFile_Dir(t *testing.T) {
	_, err := testClient("http://x").UploadFile(t.TempDir(), "", "")
	if err == nil { t.Fatal("expected error") }
}

func TestUploadFile_500(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		w.WriteHeader(500)
	})
	defer srv.Close()
	_, err := testClient(srv.URL).UploadFile(tmpFile(t, "f.bin", []byte("x")), "", "")
	if err == nil || !strings.Contains(err.Error(), "500") { t.Fatalf("expected 500: %v", err) }
}

func TestUploadFile_401(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
	defer srv.Close()
	_, err := testClient(srv.URL).UploadFile(tmpFile(t, "f.bin", []byte("x")), "", "")
	if err == nil { t.Fatal("expected 401 error") }
}

func TestUploadFile_Multiple(t *testing.T) {
	var count int
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		count++; w.WriteHeader(201)
	})
	defer srv.Close()
	c := testClient(srv.URL)
	for i := 0; i < 5; i++ {
		if _, err := c.UploadFile(tmpFile(t, "multi.bin", []byte{byte(i)}), "", ""); err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
	}
	if count != 5 { t.Errorf("count=%d", count) }
}

func TestUploadFile_LargeData(t *testing.T) {
	var sz int64
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		b, _ := io.ReadAll(r.Body)
		sz = int64(len(b))
		w.WriteHeader(201)
	})
	defer srv.Close()
	data := bytes.Repeat([]byte("X"), 1024*100)
	r, err := testClient(srv.URL).UploadFile(tmpFile(t, "large.bin", data), "", "")
	if err != nil { t.Fatal(err) }
	if sz != r.Size { t.Errorf("sent=%d received=%d", r.Size, sz) }
}

// TUS
func TestTusCreate_Success(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, err := testClient(srv.URL).tusCreate(srv.URL+"/", "t.bin", 100)
	if err != nil { t.Fatal(err) }
	if !strings.Contains(s.uploadURL, "/data/tus/") { t.Errorf("url=%q", s.uploadURL) }
}

func TestTusCreate_NoDir(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	_, err := testClient(srv.URL).tusCreate(srv.URL+"/file.bin", "t.bin", 100)
	if err == nil { t.Fatal("expected error") }
}

func TestTusCreate_NoLocation(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Tus-Resumable", "1.0.0")
		w.WriteHeader(201)
	})
	defer srv.Close()
	s, err := testClient(srv.URL).tusCreate(srv.URL+"/", "t.bin", 100)
	if err != nil { t.Fatal(err) }
	if s.uploadURL != srv.URL+"/" { t.Errorf("fallback=%q", s.uploadURL) }
}

func TestTusCreate_400(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400) })
	defer srv.Close()
	_, err := testClient(srv.URL).tusCreate(srv.URL+"/", "t.bin", 100)
	if err == nil || !strings.Contains(err.Error(), "400") { t.Fatalf("expected 400: %v", err) }
}

func TestTusCreate_ZeroSize(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, err := testClient(srv.URL).tusCreate(srv.URL+"/", "zero.bin", 0)
	if err != nil { t.Fatal(err) }
	if s == nil { t.Fatal("nil session") }
}

func TestTusPatch_Success(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "p.bin", 100)
	if err := s.patch([]byte("AAAA")); err != nil { t.Fatal(err) }; s.Offset += 4
	if err := s.patch([]byte("BBBB")); err != nil { t.Fatal(err) }; s.Offset += 4
	off, _ := s.head()
	if off != 8 { t.Errorf("offset=%d", off) }
}

func TestTusPatch_WrongOffset(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "w.bin", 100)
	s.patch([]byte("data")); s.Offset += 4
	s.Offset = 0
	if err := s.patch([]byte("bad")); err == nil || !strings.Contains(err.Error(), "409") {
		t.Fatalf("expected 409: %v", err)
	}
}

func TestTusHead_Success(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "h.bin", 200)
	off, _ := s.head()
	if off != 0 { t.Errorf("initial=%d", off) }
	s.patch(bytes.Repeat([]byte("X"), 100)); s.Offset += 100
	off, _ = s.head()
	if off != 100 { t.Errorf("after=%d", off) }
}

func TestTusHead_NotFound(t *testing.T) {
	s := &TUSSession{uploadURL: "http://x/data/tus/nope", client: testClient("http://x")}
	_, err := s.head()
	if err == nil { t.Fatal("expected error") }
}

func TestTusFileID(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "f.bin", 100)
	id, _ := s.fileID()
	if id == "" { t.Error("empty fileID") }
}

func TestTusFullFlow(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "full.bin", 1024)
	for i := 0; i < 10; i++ {
		ch := bytes.Repeat([]byte{byte(i)}, 100)
		if err := s.patch(ch); err != nil { t.Fatal(err) }; s.Offset += int64(len(ch))
	}
	off, _ := s.head()
	if off != 1000 { t.Errorf("final=%d", off) }
	id, _ := s.fileID()
	if id == "" { t.Error("no fileID") }
}

func TestUpload_AutoFallbackPUTtoTUS(t *testing.T) {
	// Server where PUT fails with 500 but TUS works
	tm := newTusMock()
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		if r.Method == "PUT" { w.WriteHeader(500); return }
		tm.ServeHTTP(w, r)
	})
	defer srv.Close()
	result, err := testClient(srv.URL).Upload(tmpFile(t, "fallback.bin", []byte("hello")), "", "", 1024)
	if err != nil { t.Fatal(err) }
	if result.Method != "TUS" { t.Errorf("expected TUS fallback, got %s", result.Method) }
	if result.Size != 5 { t.Errorf("size=%d", result.Size) }
}

func TestUpload_PUT_Success_NoFallback(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		w.Header().Set("OC-FileID", "d1!f1")
		w.WriteHeader(201)
	})
	defer srv.Close()
	result, err := testClient(srv.URL).Upload(tmpFile(t, "ok.bin", []byte("data")), "", "", 0)
	if err != nil { t.Fatal(err) }
	if result.Method != "PUT" { t.Errorf("expected PUT, got %s", result.Method) }
}

func TestUpload_PUT_Fails_TUS_Fallback(t *testing.T) {
	srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/graph/") { handleDriveOK(w, r); return }
		w.WriteHeader(413) // Payload Too Large
	})
	defer srv.Close()
	_, err := testClient(srv.URL).Upload(tmpFile(t, "big.bin", []byte("x")), "", "", 0)
	if err == nil { t.Fatal("expected error after both PUT and TUS fail") }
}

func TestTusFuzzChunks(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	seed := int64(42)
	for i := 0; i < 50; i++ {
		rng := rand.New(rand.NewSource(seed + int64(i)))
		total := rng.Int63n(1024*50) + 1
		cs := []int64{100, 500, 1024, 4999}[rng.Intn(4)]
		s, _ := testClient(srv.URL).tusCreate(srv.URL+"/", "fuzz.bin", total)
		var off int64
		for off < total {
			n := total - off
			if n > cs { n = cs }
			ch := make([]byte, n)
			s.Offset = off
			if err := s.patch(ch); err != nil {
				co, he := s.head()
				if he != nil { t.Fatalf("patch+head fail: %v / %v", err, he) }
				s.Offset = co
				continue
			}
			off += n
		}
	}
}
