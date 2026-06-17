package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// === Mock handlers ===

// Drives
func hCreateDrive(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	if _, ok := b["name"]; !ok { badRequest(w, "missing name"); return }
	createdJSON(w, map[string]interface{}{"id": "d1!new", "name": b["name"], "driveType": b["driveType"]})
}
func hGetDrive(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	if id == "missing" { notFound(w); return }
	okJSON(w, map[string]string{"id": id, "name": "TestDrive", "driveType": "personal"})
}
func hListDrives(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "d1", "name": "D1"}, {"id": "d2", "name": "D2"}}})
}
func hListDrivesBeta(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "d1", "name": "BetaDrive"}}})
}
func hUpdateDrive(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]string{"id": "d1", "name": "Updated"})
}
func hDeleteDrive(w http.ResponseWriter, r *http.Request) { noContent(w) }

// DriveItems
func hCreateDriveItem(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	createdJSON(w, map[string]interface{}{"id": "d1!fi1", "name": b["name"], "folder": map[string]int{"childCount": 0}})
}
func hGetDriveItem(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]string{"id": "d1!fi1", "name": "file.txt", "size": "100"})
}
func hDriveContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "http://example.com/dl")
	w.WriteHeader(302)
}
func hUpdateDriveItem(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]string{"id": "d1!fi1", "name": "updated.txt"})
}
func hDeleteDriveItem(w http.ResponseWriter, r *http.Request) { noContent(w) }
func hFollowDriveItem(w http.ResponseWriter, r *http.Request)  { noContent(w) }
func hUnfollowDriveItem(w http.ResponseWriter, r *http.Request) { noContent(w) }

// Permissions
func hCreateLink(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	link := map[string]interface{}{"id": "perm1", "link": map[string]interface{}{"webUrl": "http://example.com/s/abc", "type": b["type"]}}
	if pw, ok := b["password"]; ok && pw != "" && pw != nil {
		link["hasPassword"] = true
	}
	okJSON(w, link)
}
func hPermissionsList(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]interface{}{
		{"id": "p1", "link": map[string]interface{}{"webUrl": "http://x.com/s/1", "type": "view"}},
		{"id": "p2", "roles": []string{"reader"}, "grantedToV2": map[string]map[string]string{"user": {"id": "u1"}}},
	}})
}
func hPermissionsGet(w http.ResponseWriter, r *http.Request) { hPermissionsList(w, r) }
func hUpdatePermission(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]string{"id": "p1", "roles": "[\"writer\"]"})
}
func hDeletePermission(w http.ResponseWriter, r *http.Request) { noContent(w) }
func hSetPassword(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]string{"id": "p1"})
}
func hInvite(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	if _, ok := b["@libre.graph.recipient.type"]; !ok {
		badRequest(w, "missing recipient type")
		return
	}
	okJSON(w, map[string]interface{}{"id": "p3", "grantedToV2": b})
}
func hAddMember(w http.ResponseWriter, r *http.Request)     { hInvite(w, r) }
func hDeleteMember(w http.ResponseWriter, r *http.Request)   { noContent(w) }
func hListMembers(w http.ResponseWriter, r *http.Request)    { hPermissionsList(w, r) }

// Users
func hCreateUser(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	if _, ok := b["userPrincipalName"]; !ok { badRequest(w, "missing userPrincipalName"); return }
	createdJSON(w, map[string]interface{}{"id": "u-new", "displayName": b["displayName"], "userPrincipalName": b["userPrincipalName"]})
}
func hGetUser(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	if id == "missing" { notFound(w); return }
	okJSON(w, map[string]interface{}{"id": id, "displayName": "TestUser", "userPrincipalName": "test@example.com"})
}
func hListUsers(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]interface{}{{"id": "u1", "displayName": "Alice"}, {"id": "u2", "displayName": "Bob"}}})
}
func hUpdateUser(w http.ResponseWriter, r *http.Request)    { okJSON(w, map[string]string{"id": "u1", "displayName": "Updated"}) }
func hDeleteUser(w http.ResponseWriter, r *http.Request)    { noContent(w) }
func hChangePassword(w http.ResponseWriter, r *http.Request) { noContent(w) }
func hPhotoDelete(w http.ResponseWriter, r *http.Request)    { noContent(w) }
func hPhotoPut(w http.ResponseWriter, r *http.Request)       { w.WriteHeader(200) }
func hPhotoPatch(w http.ResponseWriter, r *http.Request)     { w.WriteHeader(200) }
func hUserPhoto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write([]byte("FAKEJPEG"))
}

// Groups
func hCreateGroup(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	createdJSON(w, map[string]interface{}{"id": "g-new", "displayName": b["displayName"]})
}
func hGetGroup(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"id": "g1", "displayName": "TestGroup"})
}
func hListGroups(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "g1", "displayName": "G1"}}})
}
func hUpdateGroup(w http.ResponseWriter, r *http.Request)   { okJSON(w, map[string]string{"id": "g1", "displayName": "Updated"}) }
func hDeleteGroup(w http.ResponseWriter, r *http.Request)   { noContent(w) }

// Tags
func hTags(w http.ResponseWriter, r *http.Request)          { okJSON(w, map[string]interface{}{"value": []string{"tag1", "tag2"}}) }
func hAssignTags(w http.ResponseWriter, r *http.Request)    { noContent(w) }
func hUnassignTags(w http.ResponseWriter, r *http.Request)  { noContent(w) }

// Activities, Apps
func hActivities(w http.ResponseWriter, r *http.Request)     { okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "a1"}}}) }
func hApplications(w http.ResponseWriter, r *http.Request)   { okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "app1", "displayName": "App1"}}}) }
func hGetApplication(w http.ResponseWriter, r *http.Request) { okJSON(w, map[string]string{"id": "app1", "displayName": "App1"}) }

// Invitations
func hCreateInvitation(w http.ResponseWriter, r *http.Request) {
	var b map[string]interface{}
	json.NewDecoder(r.Body).Decode(&b)
	okJSON(w, map[string]interface{}{"id": "inv1", "inviteRedirectUrl": "http://example.com/accept"})
}
func hGetInvitation(w http.ResponseWriter, r *http.Request)  { okJSON(w, map[string]string{"id": "inv1"}) }
func hListInvitations(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "inv1"}}})
}
func hRoleDefs(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "r1", "displayName": "Reader"}}})
}
func hGetRoleDef(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]string{"id": "r1", "displayName": "Reader"})
}

// AppRole
func hAppRoleList(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]string{{"id": "ar1"}}})
}
func hAppRoleCreate(w http.ResponseWriter, r *http.Request) { okJSON(w, map[string]string{"id": "ar2"}) }
func hAppRoleDelete(w http.ResponseWriter, r *http.Request) { noContent(w) }

// Home, Export
func hGetHome(w http.ResponseWriter, r *http.Request)  { okJSON(w, map[string]string{"id": "home"}) }
func hExport(w http.ResponseWriter, r *http.Request)   { w.WriteHeader(202) }
func hHomeRoot(w http.ResponseWriter, r *http.Request) { okJSON(w, map[string]string{"id": "root", "name": "root"}) }
func hHomeChildren(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]interface{}{{"id": "f1", "name": "file.txt"}, {"id": "f2", "name": "folder"}}})
}

// Shared
func hSharedByMe(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]interface{}{{"id": "f1", "name": "shared.txt"}}})
}
func hSharedWithMe(w http.ResponseWriter, r *http.Request) {
	okJSON(w, map[string]interface{}{"value": []map[string]interface{}{{"id": "f2", "name": "from_other.txt", "remoteItem": map[string]string{"id": "r1"}}}})
}

// === TUS mock ===

type tusMock struct{ sessions map[string]*tusState }
type tusState struct{ offset, total int64; name string }
var tusCnt atomic.Int64

func newTusMock() *tusMock { return &tusMock{sessions: make(map[string]*tusState)} }

func (tm *tusMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Tus-Resumable", "1.0.0")
	switch r.Method {
	case "POST":
		if !strings.HasSuffix(r.URL.Path, "/") { w.WriteHeader(412); return }
		var total int64
		fmt.Sscanf(r.Header.Get("Upload-Length"), "%d", &total)
		fn := "unknown"
		mp := strings.SplitN(r.Header.Get("Upload-Metadata"), " ", 2)
		if len(mp) == 2 && mp[0] == "filename" {
			if d, e := base64.StdEncoding.DecodeString(mp[1]); e == nil { fn = string(d) }
		}
		id := fmt.Sprintf("s%d", tusCnt.Add(1))
		tm.sessions[id] = &tusState{total: total, name: fn}
		w.Header().Set("Location", fmt.Sprintf("http://%s/data/tus/%s", r.Host, id))
		w.WriteHeader(201)
	case "HEAD":
		id := filepath.Base(r.URL.Path)
		if s, ok := tm.sessions[id]; ok {
			w.Header().Set("Upload-Offset", fmt.Sprintf("%d", s.offset))
			w.Header().Set("Upload-Length", fmt.Sprintf("%d", s.total))
			w.Header().Set("OC-FileID", fmt.Sprintf("d1!%s", id))
			w.WriteHeader(200)
		} else { w.WriteHeader(404) }
	case "PATCH":
		id := filepath.Base(r.URL.Path)
		s, ok := tm.sessions[id]
		if !ok { w.WriteHeader(404); return }
		var co int64
		fmt.Sscanf(r.Header.Get("Upload-Offset"), "%d", &co)
		if co != s.offset { w.WriteHeader(409); return }
		b, _ := io.ReadAll(r.Body)
		s.offset += int64(len(b))
		w.Header().Set("Upload-Offset", fmt.Sprintf("%d", s.offset))
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}

// === Unified mux router ===

func muxHandler(w http.ResponseWriter, r *http.Request) {
	p, m := r.URL.Path, r.Method

	switch {
	// Drives
	case p == "/graph/v1.0/drives" && m == "POST": hCreateDrive(w, r)
	case p == "/graph/v1.0/drives" && m == "GET": hListDrives(w, r)
	case contains(p, "/graph/v1.0/drives/") && m == "GET": hGetDrive(w, r)
	case contains(p, "/graph/v1.0/drives/") && m == "PATCH": hUpdateDrive(w, r)
	case contains(p, "/graph/v1.0/drives/") && m == "DELETE": hDeleteDrive(w, r)

	// Beta drives (exact paths, not substrings)
	case p == "/graph/v1beta1/me/drives" && m == "GET": hListDrivesBeta(w, r)
	case p == "/graph/v1beta1/drives" && m == "GET": hListDrivesBeta(w, r)

	// Permissions / Links (before generic /items/)
	case contains(p, "/createLink") && m == "POST": hCreateLink(w, r)
	case contains(p, "/invite") && m == "POST" && contains(p, "/root/"): hAddMember(w, r)
	case contains(p, "/invite") && m == "POST": hInvite(w, r)
	case contains(p, "/permissions/") && hasSuffix(p, "/setPassword") && m == "POST": hSetPassword(w, r)
	case contains(p, "/permissions/") && m == "DELETE": hDeletePermission(w, r)
	case contains(p, "/permissions/") && m == "PATCH": hUpdatePermission(w, r)
	case contains(p, "/permissions/") && m == "GET": hPermissionsGet(w, r)
	case contains(p, "/permissions") && m == "GET": hPermissionsList(w, r)
	case contains(p, "/members/") && m == "DELETE": hDeleteMember(w, r)
	case contains(p, "/members") && m == "GET": hListMembers(w, r)

	// Space root operations (before beta drives)
	case contains(p, "/root/createLink") && m == "POST": hCreateLink(w, r)
	case contains(p, "/root/permissions/") && hasSuffix(p, "/setPassword") && m == "POST": hSetPassword(w, r)
	case contains(p, "/root/permissions/") && m == "DELETE": hDeletePermission(w, r)
	case contains(p, "/root/permissions/") && m == "PATCH": hUpdatePermission(w, r)
	case contains(p, "/root/permissions") && m == "GET": hPermissionsList(w, r)

	// Tags (before /items/)
	case contains(p, "/tags") && contains(p, "/assign"): hAssignTags(w, r)
	case contains(p, "/tags") && contains(p, "/unassign"): hUnassignTags(w, r)
	case contains(p, "/tags"): hTags(w, r)

	// Photo
	case contains(p, "/photo") && m == "PUT": hPhotoPut(w, r)
	case contains(p, "/photo") && m == "PATCH": hPhotoPatch(w, r)
	case contains(p, "/photo") && m == "DELETE": hPhotoDelete(w, r)
	case contains(p, "/photo") && m == "GET": hUserPhoto(w, r)

	// Drive Items
	case contains(p, "/items/") && hasSuffix(p, "/content") && m == "GET": hDriveContent(w, r)
	case contains(p, "/follow"): hFollowDriveItem(w, r)
	case contains(p, "/unfollow"): hUnfollowDriveItem(w, r)
	case contains(p, "/items/") && m == "PATCH": hUpdateDriveItem(w, r)
	case contains(p, "/items/") && m == "DELETE": hDeleteDriveItem(w, r)
	case contains(p, "/items/") && m == "GET": hGetDriveItem(w, r)
	case contains(p, "/root/children") && m == "POST": hCreateDriveItem(w, r)

	// Shared
	case contains(p, "/sharedByMe"): hSharedByMe(w, r)
	case contains(p, "/sharedWithMe"): hSharedWithMe(w, r)

	// AppRole (before /users/)
	case contains(p, "/appRoleAssignments") && m == "DELETE": hAppRoleDelete(w, r)
	case contains(p, "/appRoleAssignments") && m == "POST": hAppRoleCreate(w, r)
	case contains(p, "/appRoleAssignments"): hAppRoleList(w, r)

	// Change password
	case contains(p, "/changepassword"): hChangePassword(w, r)

	// Users
	case p == "/graph/v1.0/users" && m == "POST": hCreateUser(w, r)
	case p == "/graph/v1.0/users" && m == "GET": hListUsers(w, r)
	case contains(p, "/users/") && contains(p, "/photo"): hUserPhoto(w, r)
	case contains(p, "/users/") && m == "PATCH": hUpdateUser(w, r)
	case contains(p, "/users/") && m == "GET": hGetUser(w, r)
	case contains(p, "/users/") && m == "DELETE": hDeleteUser(w, r)

	// Groups
	case p == "/graph/v1.0/groups" && m == "POST": hCreateGroup(w, r)
	case p == "/graph/v1.0/groups" && m == "GET": hListGroups(w, r)
	case contains(p, "/groups/") && m == "PATCH": hUpdateGroup(w, r)
	case contains(p, "/groups/") && m == "GET": hGetGroup(w, r)
	case contains(p, "/groups/") && m == "DELETE": hDeleteGroup(w, r)

	// Activities / Applications
	case contains(p, "/activities"): hActivities(w, r)
	case hasSuffix(p, "/applications") && m == "GET": hApplications(w, r)
	case contains(p, "/applications/") && m == "GET": hGetApplication(w, r)

	// Invitations
	case contains(p, "/invitations") && m == "POST": hCreateInvitation(w, r)
	case contains(p, "/invitations/") && m == "GET": hGetInvitation(w, r)
	case contains(p, "/invitations") && m == "GET": hListInvitations(w, r)

	// Role management
	case contains(p, "/roleManagement") && contains(p, "/roleDefinitions"):
		if strings.Count(p, "/") > strings.Count("/v1.0/roleManagement/roleDefinitions", "/") {
			hGetRoleDef(w, r)
		} else {
			hRoleDefs(w, r)
		}

	// Home / Export
	case contains(p, "/home"): hGetHome(w, r)
	case contains(p, "/exportPersonalData"): hExport(w, r)
	case hasSuffix(p, "/root") && m == "GET": hHomeRoot(w, r)
	case contains(p, "/root/children") && m == "GET": hHomeChildren(w, r)

	default:
		notFound(w)
	}
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
