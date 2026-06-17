package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type DriveInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	DriveType string `json:"driveType"`
	Root      struct {
		WebDavURL string `json:"webDavUrl"`
	} `json:"root"`
}

type UploadResult struct {
	StatusCode int
	FileID     string
	Size       int64
	Method     string // "PUT" or "TUS"
}

// Upload uploads localPath to the personal drive. Small files use a simple
// WebDAV PUT. If the PUT fails (413, 500, connection error), TUS chunked
// upload is automatically retried. Pass chunkSize=0 for default 5MiB.
func (c *Client) Upload(localPath, remoteFilename, mimeType string, chunkSize int64) (*UploadResult, error) {
	if chunkSize <= 0 {
		chunkSize = 5 * 1024 * 1024
	}
	if remoteFilename == "" {
		remoteFilename = filepath.Base(localPath)
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Try simple PUT first
	result, err := c.UploadFile(localPath, remoteFilename, mimeType)
	if err == nil && result.StatusCode < 400 {
		result.Method = "PUT"
		return result, nil
	}

	// PUT failed — log and fall back to TUS
	slog.Debug("PUT failed, falling back to TUS", "error", err)
	result, err = c.UploadFileTUS(localPath, remoteFilename, chunkSize)
	if err != nil {
		return nil, err
	}
	result.Method = "TUS"
	return result, nil
}

func (c *Client) GetPersonalDrive() (*DriveInfo, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("GetPersonalDrive: %w", err)
	}
	u, err := url.JoinPath(c.baseURL, "graph", "v1.0", "me", "drive")
	if err != nil {
		return nil, fmt.Errorf("GetPersonalDrive build URL: %w", err)
	}
	resp, err := c.DoRequest("GET", u, nil, 0, "", map[string]string{
		"Authorization": "Bearer " + token,
	})
	if err != nil {
		return nil, fmt.Errorf("GetPersonalDrive: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GetPersonalDrive HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var drive DriveInfo
	if err := json.NewDecoder(resp.Body).Decode(&drive); err != nil {
		return nil, fmt.Errorf("GetPersonalDrive decode: %w", err)
	}
	return &drive, nil
}

func (c *Client) UploadFile(localPath, remoteFilename, mimeType string) (*UploadResult, error) {
	drive, err := c.GetPersonalDrive()
	if err != nil {
		return nil, err
	}
	if remoteFilename == "" {
		remoteFilename = filepath.Base(localPath)
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	uploadURL := c.buildUploadURL(drive, remoteFilename)
	return c.putFile(localPath, uploadURL, mimeType)
}

func (c *Client) buildUploadURL(drive *DriveInfo, filename string) string {
	return strings.TrimRight(drive.Root.WebDavURL, "/") + "/" + url.PathEscape(filename)
}

func (c *Client) putFile(localPath, uploadURL, mimeType string) (*UploadResult, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("putFile open %s: %w", localPath, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("putFile stat %s: %w", localPath, err)
	}
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("putFile: %w", err)
	}
	slog.Debug("PUT upload", "url", uploadURL, "size", fi.Size())
	resp, err := c.DoRequest("PUT", uploadURL, f, fi.Size(), mimeType, map[string]string{
		"Authorization": "Bearer " + token,
	})
	if err != nil {
		return nil, fmt.Errorf("putFile %s: %w", uploadURL, err)
	}
	defer resp.Body.Close()
	result := &UploadResult{StatusCode: resp.StatusCode, Size: fi.Size()}
	if fid := resp.Header.Get("OC-FileID"); fid != "" {
		result.FileID = fid
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("putFile HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	slog.Debug("PUT complete", "fileID", result.FileID, "status", resp.StatusCode)
	return result, nil
}

// --- TUS ---

const tusVersion = "1.0.0"

type TUSSession struct {
	uploadURL string
	client    *Client
	Offset    int64
}

func (c *Client) UploadFileTUS(localPath, remoteFilename string, chunkSize int64) (*UploadResult, error) {
	drive, err := c.GetPersonalDrive()
	if err != nil {
		return nil, err
	}
	if remoteFilename == "" {
		remoteFilename = filepath.Base(localPath)
	}
	if chunkSize <= 0 {
		chunkSize = 5 * 1024 * 1024
	}
	fi, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("UploadFileTUS stat %s: %w", localPath, err)
	}
	fileSize := fi.Size()
	davBase := strings.TrimRight(drive.Root.WebDavURL, "/")
	slog.Debug("TUS upload", "file", localPath, "filename", remoteFilename, "size", fileSize, "chunks", chunkSize)
	session, err := c.tusCreate(davBase+"/", remoteFilename, fileSize)
	if err != nil {
		return nil, fmt.Errorf("UploadFileTUS create: %w", err)
	}
	slog.Debug("TUS session", "url", session.uploadURL)
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("UploadFileTUS open %s: %w", localPath, err)
	}
	defer f.Close()
	buf := make([]byte, chunkSize)
	remaining := fileSize
	for remaining > 0 {
		n, err := f.ReadAt(buf, session.Offset)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("UploadFileTUS read offset %d: %w", session.Offset, err)
		}
		if n == 0 {
			break
		}
		chunk := buf[:n]
		slog.Debug("TUS chunk", "offset", session.Offset, "len", n)
		if err := session.patch(chunk); err != nil {
			if strings.Contains(err.Error(), "409") {
				co, he := session.head()
				if he != nil {
					return nil, fmt.Errorf("UploadFileTUS offset %d: patch+head failed: patch=%w head=%w", session.Offset, err, he)
				}
				slog.Debug("TUS resume", "from", session.Offset, "to", co)
				session.Offset = co
				if err := session.patch(chunk); err != nil {
					return nil, fmt.Errorf("UploadFileTUS resume offset %d: %w", session.Offset, err)
				}
				session.Offset += int64(n)
				remaining -= int64(n)
				continue
			}
			return nil, fmt.Errorf("UploadFileTUS chunk offset %d: %w", session.Offset, err)
		}
		session.Offset += int64(n)
		remaining -= int64(n)
	}
	slog.Debug("TUS complete", "total", fileSize)
	fid, _ := session.fileID()
	return &UploadResult{StatusCode: 201, FileID: fid, Size: fileSize}, nil
}

func (c *Client) tusCreate(targetURL, filename string, fileSize int64) (*TUSSession, error) {
	meta := fmt.Sprintf("filename %s", base64.StdEncoding.EncodeToString([]byte(filename)))
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("tusCreate: %w", err)
	}
	resp, err := c.DoRequest("POST", targetURL, nil, 0, "application/offset+octet-stream", map[string]string{
		"Authorization":   "Bearer " + token,
		"Tus-Resumable":   tusVersion,
		"Upload-Length":   fmt.Sprintf("%d", fileSize),
		"Upload-Metadata": meta,
	})
	if err != nil {
		return nil, fmt.Errorf("tusCreate POST %s: %w", targetURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tusCreate HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	uploadURL := resp.Header.Get("Location")
	if uploadURL == "" {
		uploadURL = targetURL
	}
	return &TUSSession{uploadURL: uploadURL, client: c}, nil
}

func (s *TUSSession) patch(chunk []byte) error {
	resp, err := s.client.DoRequest("PATCH", s.client.resolveURL(s.uploadURL),
		bytes.NewReader(chunk), int64(len(chunk)),
		"application/offset+octet-stream", map[string]string{
			"Tus-Resumable": tusVersion,
			"Upload-Offset": fmt.Sprintf("%d", s.Offset),
		})
	if err != nil {
		return fmt.Errorf("tusPatch offset %d: %w", s.Offset, err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 409:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tusPatch offset %d: 409 conflict: %s", s.Offset, strings.TrimSpace(string(body)))
	case 204, 200, 201:
		return nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tusPatch offset %d: HTTP %d: %s", s.Offset, resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func (s *TUSSession) head() (int64, error) {
	resp, err := s.client.DoRequest("HEAD", s.client.resolveURL(s.uploadURL),
		nil, 0, "", map[string]string{"Tus-Resumable": tusVersion})
	if err != nil {
		return 0, fmt.Errorf("tusHead %s: %w", s.uploadURL, err)
	}
	defer resp.Body.Close()
	os := resp.Header.Get("Upload-Offset")
	if os == "" {
		return 0, fmt.Errorf("tusHead %s: missing Upload-Offset header", s.uploadURL)
	}
	var o int64
	if _, err := fmt.Sscanf(os, "%d", &o); err != nil {
		return 0, fmt.Errorf("tusHead %s: invalid Upload-Offset %q: %w", s.uploadURL, os, err)
	}
	return o, nil
}

func (s *TUSSession) fileID() (string, error) {
	resp, err := s.client.DoRequest("HEAD", s.client.resolveURL(s.uploadURL),
		nil, 0, "", map[string]string{"Tus-Resumable": tusVersion})
	if err != nil {
		return "", fmt.Errorf("tusFileID %s: %w", s.uploadURL, err)
	}
	defer resp.Body.Close()
	return resp.Header.Get("OC-FileID"), nil
}
