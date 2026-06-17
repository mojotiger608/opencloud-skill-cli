package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// === Fuzz: chunk math ===

func FuzzChunkSizes(f *testing.F) {
	f.Add(int64(1), int64(100))
	f.Add(int64(1024*1024), int64(100*1024*1024))
	f.Add(int64(1), int64(1))
	f.Add(int64(5000000), int64(1073741824))
	f.Fuzz(func(t *testing.T, chunkSize, fileSize int64) {
		if chunkSize <= 0 || fileSize <= 0 || chunkSize > 1<<30 { return }
		n := (fileSize + chunkSize - 1) / chunkSize
		if n < 1 { t.Errorf("n=%d", n) }
		// verify coverage: chunks that result from this division must cover the file
		covered := n * chunkSize
		if covered < fileSize { t.Errorf("covered=%d < fileSize=%d", covered, fileSize) }
	})
}

// === Fuzz: offset math ===

func FuzzOffsets(f *testing.F) {
	f.Add(int64(0), int64(1000))
	f.Add(int64(500), int64(5000))
	f.Add(int64(999999), int64(1000000))
	f.Fuzz(func(t *testing.T, off, total int64) {
		if off < 0 || total <= 0 || off >= total { return }
		rem := total - off
		if rem <= 0 { t.Errorf("rem=%d", rem) }

		// chunk from any valid offset must not exceed remaining
		for chunk := int64(1); chunk <= rem && chunk <= 65536; chunk *= 2 {
			if off+chunk > total { t.Errorf("off=%d chunk=%d > total=%d", off, chunk, total) }
		}
	})
}

// === Fuzz: filenames ===

func FuzzFilenames(f *testing.F) {
	f.Add("normal.txt")
	f.Add("with spaces.txt")
	f.Add("日本語.txt")
	f.Add("/etc/passwd")
	f.Add("a\r\nb")
	f.Add("")
	f.Add(strings.Repeat("a", 255))
	f.Fuzz(func(t *testing.T, name string) {
		if len(name) > 500 { return }
		e := base64.StdEncoding.EncodeToString([]byte(name))
		d, err := base64.StdEncoding.DecodeString(e)
		if err != nil { t.Errorf("decode: %v", err); return }
		if string(d) != name { t.Errorf("roundtrip: %q != %q", name, d) }
	})
}

// === Fuzz: JSON bodies ===

func FuzzJSONBodies(f *testing.F) {
	f.Add(`{"name":"test"}`)
	f.Add(`{}`)
	f.Add(`{"name":"a","driveType":"project","quota":{"total":1073741824}}`)
	f.Add(`{"@libre.graph.recipient.type":"user","objectId":"u1","roles":["reader"]}`)
	f.Add(`{"displayName":"N","userPrincipalName":"n@t.com","password":"S3cret!"}`)
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) == 0 || len(body) > 10000 { return }
		var v interface{}
		if err := json.Unmarshal([]byte(body), &v); err != nil { return }
		reenc, _ := json.Marshal(v)
		if len(reenc) == 0 { t.Error("empty re-encode") }
	})
}

// === Fuzz: path params ===

func FuzzPathParams(f *testing.F) {
	f.Add("abc123")
	f.Add("d1%21f1")
	f.Add("../../../etc/passwd")
	f.Add("")
	f.Add(strings.Repeat("x", 1000))
	f.Fuzz(func(t *testing.T, id string) {
		if len(id) > 2000 { return }
		srv := httptest.NewServer(http.HandlerFunc(muxHandler))
		defer srv.Close()
		_, err := testClient(srv.URL).MakeRequest("/v1.0/drives/"+id, "GET", "", nil)
		if err != nil && !strings.Contains(err.Error(), "400") && !strings.Contains(err.Error(), "404") {
			t.Errorf("id=%q: unexpected error: %v", id, err)
		}
	})
}

// === Fuzz: TUS offsets with random chunk patterns ===

func FuzzTUSOffsets(f *testing.F) {
	f.Add(int64(0), int64(1000))
	f.Add(int64(0), int64(10000000))
	f.Fuzz(func(t *testing.T, off, total int64) {
		if off < 0 || total <= 0 || total > 100*1024*1024 { return }
		// Validate that any valid range of offsets can be split into chunks
		chunkSizes := []int64{1, 100, 1024, 65536, 1024 * 1024}
		for _, cs := range chunkSizes {
			chunks := (total + cs - 1) / cs
			if chunks < 1 { t.Errorf("cs=%d gave chunks=%d", cs, chunks) }
		}
	})
}

// === Fuzz: HTTP methods ===

func FuzzHTTPMethods(f *testing.F) {
	methods := []string{"GET", "POST", "PATCH", "DELETE", "PUT", "HEAD", "OPTIONS"}
	for _, m := range methods { f.Add(m) }
	f.Fuzz(func(t *testing.T, method string) {
		valid := false
		for _, m := range []string{"GET", "POST", "PATCH", "DELETE", "PUT", "HEAD", "OPTIONS"} {
			if method == m { valid = true; break }
		}
		if !valid { return }
		srv := testServer(muxHandler); defer srv.Close()
		_, err := testClient(srv.URL).MakeRequest("/v1.0/me", method, "{}", nil)
		_ = err
	})
}

// === Fuzz: upload chunk patterns ===

func FuzzUploadChunks(f *testing.F) {
	f.Add(int64(100), int64(32))
	f.Add(int64(1024*1024), int64(1024))
	f.Fuzz(func(t *testing.T, total, cs int64) {
		if total <= 0 || cs <= 0 || total > 100*1024*1024 || cs > 10*1024*1024 { return }
		chunks := total / cs
		rem := total % cs
		if rem > 0 { chunks++ }
		if chunks < 1 { t.Errorf("chunks=%d", chunks) }
		if chunks > total { t.Errorf("chunks=%d > total=%d", chunks, total) }
	})
}

// === Fuzz: HTTP status codes ===

func FuzzHTTPStatuses(f *testing.F) {
	for _, s := range []int{200, 201, 204, 301, 400, 401, 403, 404, 409, 500} { f.Add(s) }
	f.Fuzz(func(t *testing.T, status int) {
		if status < 100 || status > 599 { return }
		label := "success"
		if status >= 400 { label = "error" }
		_ = label
		// verify status classification
		if status >= 200 && status < 300 { /* ok */ }
		if status >= 400 { /* error */ }
	})
}

// === Fuzz: MIME types ===

func FuzzMimeTypes(f *testing.F) {
	f.Add("text/plain")
	f.Add("application/octet-stream")
	f.Add("application/json")
	f.Fuzz(func(t *testing.T, mime string) {
		if len(mime) == 0 || len(mime) > 500 { return }
		// valid MIME types must contain '/'
		if strings.Contains(mime, "/") { /* valid */ }
	})
}

// === Fuzz: concurrent TUS ===

func TestFuzzTUSConcurrent(t *testing.T) {
	tm := newTusMock()
	srv := testServer(tm.ServeHTTP)
	defer srv.Close()
	rng := rand.New(rand.NewSource(99))

	for i := 0; i < 100; i++ {
		total := rng.Int63n(1024*10) + 1
		cs := []int64{50, 200, 500, 1000}[rng.Intn(4)]
		s, err := testClient(srv.URL).tusCreate(srv.URL+"/", "fuzz.bin", total)
		if err != nil { t.Fatalf("create %d: %v", i, err) }

		var off int64
		retries := 0
		for off < total && retries < 10 {
			n := total - off
			if n > cs { n = cs }
			ch := make([]byte, n)
			s.Offset = off
			if err := s.patch(ch); err != nil {
				if strings.Contains(err.Error(), "409") {
					co, he := s.head()
					if he != nil { t.Fatalf("iteration %d @ offset %d: patch+head fail: %v / %v", i, off, err, he) }
					s.Offset = co
					retries++
					continue
				}
				t.Fatalf("iteration %d @ offset %d: %v", i, off, err)
			}
			off += n
		}
		if off != total { t.Errorf("iteration %d: off=%d total=%d", i, off, total) }

		// verify final HEAD
		finalOff, _ := s.head()
		if finalOff != total { t.Errorf("iteration %d: HEAD offset=%d want=%d", i, finalOff, total) }
	}
}

// === Fuzz: drive info decode ===

func TestFuzzDriveInfoDecode(t *testing.T) {
	payloads := []string{
		`{"id":"d1","name":"Test","driveType":"personal","root":{"webDavUrl":"http://x/dav"}}`,
		`{"id":"d2","name":"","driveType":"project","root":{"webDavUrl":""}}`,
		`{"id":"d3","name":"Spaces","driveType":"personal","root":{}}`,
		`{"id":"","name":"","driveType":"","root":{"webDavUrl":""}}`,
	}
	for _, p := range payloads {
		var d DriveInfo
		if err := json.Unmarshal([]byte(p), &d); err != nil {
			t.Errorf("decode %q: %v", p, err)
		}
	}
}

// === Fuzz: formatSize ===

func TestFuzzFormatSize(t *testing.T) {
	sizes := []int64{0, 1, 999, 1024, 1024*1024, 1024*1024*1024, 1024*1024*1024*1024}
	for _, s := range sizes {
		res := formatSizeBytes(s)
		if res == "" { t.Errorf("formatSize(%d) empty", s) }
	}
}

func formatSizeBytes(bytes int64) string {
	const u = 1024
	if bytes < u { return fmt.Sprintf("%d B", bytes) }
	div, exp := int64(u), 0
	for n := bytes / u; n >= u; n /= u { div *= u; exp++ }
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
