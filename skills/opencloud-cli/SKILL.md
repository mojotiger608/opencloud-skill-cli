---
name: opencloud-cli
description: 'Manage files, folders, spaces and file uploads on an OpenCloud server. Use TOON format.'
---

# OpenCloud CLI

`oc-cli` manages OpenCloud via LibreGraph API. Responses use TOON. Uploads use WebDAV PUT or TUS.

## Setup

```
oc-cli login --server-url URL [--insecure] [--host H] [--ip X.X.X.X]
```

Config persists at `~/.config/opencloud-cli/config.json`. `--host` and `--ip` save to config — no need to repeat flags.

## Commands

### Upload
```
oc-cli upload <file>                  # PUT upload
oc-cli upload <file> --chunked        # TUS chunked (resumable)
oc-cli upload <file> --chunk-size N   # Custom chunk (default 5MB)
oc-cli upload <file> --name remote    # Rename remote
oc-cli upload --drive-info            # Show drive + host/ip
```

### API
```
oc-cli api -p /PATH -m METHOD -b BODY -q K=V
oc-cli api -p /v1.0/me/drive
oc-cli api -p /v1.0/me/drive/root/children
oc-cli api -p /v1.0/me/drive/root/children -m POST -b '{"name":"d"}'
oc-cli api -p /v1.0/me/drive/items/ID -m DELETE --status-only
oc-cli api -p /v1.0/me/drive/items/ID/createLink -m POST -b '{"type":"view"}'
oc-cli api -p /v1.0/me/drive/items/ID/permissions -m GET
oc-cli api -p /v1.0/users -q '$search="name"'
oc-cli api -p /v1.0/drives -m POST -b '{"name":"s","driveType":"project"}'
oc-cli api -p /v1.0/groups -m POST -b '{"displayName":"g"}'
```

### Override per-command
```
oc-cli upload file --host H --ip X.X.X.X   # overrides config for one call
```

## Rules

- Prefer `v1beta1` over `v1.0` if both exist
- Wrap paths/body in single quotes
- `--status-only` when body not needed
- Link passwords: 8+ chars, upper+lower+digit+special
