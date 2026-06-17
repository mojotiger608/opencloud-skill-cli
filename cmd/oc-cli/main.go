package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/JammingBen/opencloud-skill-cli/internal/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oc-cli",
	Short: "OpenCloud CLI — LibreGraph API + file uploads",
	Long: `OpenCloud CLI — interact with OpenCloud servers via the LibreGraph API.

COMMAND GROUPS:
  Auth:     login, logout
  Files:    upload
  API:      api
  System:   version, install-skill

AUTH:
  oc-cli login --server-url <url> [--insecure] [--host H] [--ip X.X.X.X]
  oc-cli logout
  Config saved at ~/.config/opencloud-cli/config.json (0600).
  --host and --ip persist — no need to repeat flags.
  Env fallback: OC_ACCESS_TOKEN.

FILES:
  oc-cli upload <file>                  PUT first, auto-fallback to TUS
  oc-cli upload <file> --chunk-size N   Custom TUS chunk size (default 5MB)
  oc-cli upload <file> --name remote    Rename on upload
  oc-cli upload --drive-info            Show drive + host/ip from config

API (LibreGraph):
  oc-cli api -p /PATH -m METHOD -b BODY -q KEY=VALUE
  oc-cli api -p /v1.0/me/drive
  oc-cli api -p /v1.0/me/drive/root/children
  oc-cli api -p /v1.0/me/drive/root/children -m POST -b '{"name":"d"}'
  oc-cli api -p /v1.0/me/drive/items/ID -m DELETE --status-only
  oc-cli api -p /v1.0/me/drive/items/ID/createLink -m POST -b '{"type":"view"}'
  oc-cli api -p /v1.0/me/drive/items/ID/permissions -m GET
  oc-cli api -p /v1.0/users -q '$search="name"'
  oc-cli api -p /v1.0/drives -m POST -b '{"name":"s","driveType":"project"}'
  oc-cli api -p /v1.0/groups -m POST -b '{"displayName":"g"}'

FLAGS:
  --status-only   Print HTTP status only (no body)
  --json-format   Output as JSON (default: TOON)
  --host          HTTP Host header override (persists via login)
  --ip            DNS resolution override (persists via login)
  -v --verbose    Debug output`,
	SilenceUsage:  false,
	SilenceErrors: true,
}

func main() {
	logger.SetupLogging(slog.LevelInfo)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.RedString("error: %v", err))
		os.Exit(1)
	}
}
