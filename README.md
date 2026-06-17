# opencloud-skill-cli

CLI for OpenCloud's LibreGraph API. Upload files (auto PUT → TUS fallback), manage drives, files, folders, users, groups, permissions, shares.

## Quick Start

```sh
go build -o bin/oc-cli ./cmd/oc-cli/
oc-cli login --server-url https://<your-server>[:port]
oc-cli upload file.pdf
oc-cli api -p /v1.0/me/drive
```

## Auth

`oc-cli login` uses OIDC PKCE. Config saved at `~/.config/opencloud-cli/config.json` (0600):

```json
{
  "server_url": "https://your.server:9200",
  "insecure": true,
  "token": {...},
  "client_id": "web",
  "token_endpoint": "https://your.server:9200/konnect/v1/token"
}
```

Optional fields for connecting to an IP while presenting a hostname:

```json
{
  "host": "your.domain.com",
  "ip": "192.168.1.10"
}
```

Set these once via `oc-cli login --host H --ip X --insecure`. All subsequent commands use them automatically.

## Upload Behavior

**Auto-fallback**: WebDAV PUT is tried first. If it fails (HTTP 413, 500, or connection error), the upload automatically falls back to TUS chunked upload.

```sh
oc-cli upload file.pdf              # PUT first, TUS fallback on failure
oc-cli upload large.iso             # auto
oc-cli upload huge.bin --chunk-size 16777216  # custom chunk
```

## Test Environment

For integration tests against a live OpenCloud server:

```sh
export OC_TEST_SERVER_URL="https://your.server:9200"
export OC_TEST_USER="admin"
export OC_TEST_PASS="password"
export OC_TEST_INSECURE="true"
# Optional: for IP-with-hostname setups
export OC_TEST_HOST="cloud.your.domain"
export OC_TEST_IP="192.168.1.10"
```

Tests auto-create and delete data — no server clutter.

## Build & Test

```sh
# Requires Go 1.23+
go build -o bin/oc-cli ./cmd/oc-cli/

# Unit tests (192 tests, mock server — no auth/server needed)
go test -v ./internal/client/

# Fuzz tests (10 suites)
for f in ChunkSizes Offsets Filenames JSONBodies PathParams \
         TUSOffsets HTTPMethods UploadChunks HTTPStatuses MimeTypes; do
  go test -run='^$' -fuzz="^Fuzz$f$" -fuzztime=30s ./internal/client/
done
```

## Commands

| Command | Purpose |
|---------|---------|
| `login` | OIDC auth, saves config (server, host, ip, insecure) |
| `logout` | Clear all config |
| `upload` | Auto PUT → TUS fallback |
| `api` | Raw LibreGraph API calls |
| `version` | Build version |
| `install-skill` | Install agent skill |

## Upload Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--chunk-size` | 5MB | TUS chunk size (used on fallback) |
| `--name` `-n` | filename | Remote filename |
| `--mime` `-m` | auto | MIME type |
| `--drive-info` | false | Show drive (includes host/ip from config) |
| `--host` | config | Host header override |
| `--ip` | config | DNS resolution override |
| `-v` | false | Verbose |

## API Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-p` | (required) | API path |
| `-m` | GET | HTTP method |
| `-b` | "" | JSON body |
| `-q` | [] | Query params (key=value) |
| `--status-only` | false | Print status only |
| `--json-format` | false | JSON output (default: TOON) |
| `--host` | config | Host header |
| `--ip` | config | DNS resolution |
| `-v` | false | Verbose |

## Common Setups

### Local Docker (9200)

```sh
oc-cli login --server-url https://127.0.0.1:9200 --insecure
```

### Reverse proxy / Cloudflare Tunnel (443)

```sh
oc-cli login --server-url https://your.domain.com
```

### IP + hostname (LAN)

```sh
oc-cli login --server-url https://192.168.1.10:9200 \
  --host your.domain.com --ip 192.168.1.10 --insecure
```

## Architecture

```
internal/client/
├── client.go       # HTTP client: Host/IP override, OAuth2, shared transport
├── encoder.go      # TOON/JSON response encoding
├── upload.go       # Upload (PUT → TUS fallback), GetPersonalDrive
├── helpers_test.go # Shared test infrastructure
├── mock_test.go    # Mock handlers + unified mux router (all 68 endpoints)
├── upload_test.go  (27 tests)   upload_test.go
├── drives_test.go  (12 tests)   drives_test.go
├── files_test.go   (14 tests)   files_test.go
├── permissions_test.go (18)     permissions_test.go
├── users_test.go   (21 tests)   users_test.go
├── tags_test.go    (17 tests)   tags_test.go
├── errors_test.go  (28 tests)   errors_test.go
└── fuzz_test.go    (13 + 10 fuzz)

cmd/oc-cli/
├── main.go         # CLI root with categorized help
├── login.go        # OIDC login, saves host/ip to config
├── logout.go       # Clear config
├── api.go          # API command, loads host/ip from config
├── upload.go       # Upload command, auto PUT → TUS

skills/opencloud-cli/
├── SKILL.md        # Agent skill (TOON format, token-efficient)
└── references/     # 68 operations from npx openapi-to-skills
```
