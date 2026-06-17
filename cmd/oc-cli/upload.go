package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/JammingBen/opencloud-skill-cli/internal/client"
	"github.com/JammingBen/opencloud-skill-cli/internal/logger"
	"github.com/JammingBen/opencloud-skill-cli/internal/oidc"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	uploadRemoteName    string
	uploadMimeType      string
	uploadChunkSize     int64
	uploadVerbose       bool
	uploadShowDriveInfo bool
	uploadHostOverride  string
	uploadIPOverride    string
)

var uploadCmd = &cobra.Command{
	Use:   "upload [local-file]",
	Short: "Upload a file to your OpenCloud personal drive",
	Long: `Upload a file to your OpenCloud personal drive.

Attempts a simple WebDAV PUT first. If the PUT fails (413, 500, or
connection error), automatically falls back to TUS chunked upload.

Use --drive-info to display drive details including host/ip from config.
--host and --ip override config values (set in config via oc-cli login).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logLvl := slog.LevelInfo
		if uploadVerbose {
			logLvl = slog.LevelDebug
		}
		logger.SetupLogging(logLvl)

		cfg, err := oidc.LoadConfig()
		if err != nil {
			return err
		}
		if cfg.ServerURL == "" {
			return fmt.Errorf("Server URL is not configured. Run 'oc-cli login' first.")
		}

		ctx := context.Background()
		ts, err := cfg.GetTokenSource(ctx)
		if err != nil {
			return err
		}
		if ts == nil {
			return fmt.Errorf("Access token not found. Run 'oc-cli login' first.")
		}

		c := client.NewClient(cfg.ServerURL, cfg.Insecure, ts)
		c.HostOverride = firstNonEmpty(uploadHostOverride, cfg.HostOverride)
		c.ResolveIP = firstNonEmpty(uploadIPOverride, cfg.ResolveIP)

		if uploadShowDriveInfo {
			return showDriveInfo(c)
		}
		if len(args) == 0 {
			return fmt.Errorf("missing local file path")
		}

		localPath := args[0]
		fi, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", localPath, err)
		}
		if fi.IsDir() {
			return fmt.Errorf("%s is a directory; only files can be uploaded", localPath)
		}

		start := time.Now()
		fmt.Fprintf(cmd.OutOrStdout(), "Uploading %s (%s)...\n", localPath, formatSize(fi.Size()))

		result, err := c.Upload(localPath, uploadRemoteName, uploadMimeType, uploadChunkSize)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		elapsed := time.Since(start)
		var speedStr string
		if elapsed.Seconds() > 0 {
			speedMBs := float64(result.Size) / elapsed.Seconds() / 1024 / 1024
			speedStr = fmt.Sprintf(", %.1f MB/s", speedMBs)
		}

		msg := color.GreenString("✓ Uploaded via %s in %.1fs%s", result.Method, elapsed.Seconds(), speedStr)
		fmt.Fprintln(cmd.OutOrStdout(), msg)
		if result.FileID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  File ID: %s\n", result.FileID)
		}
		return nil
	},
}

func showDriveInfo(c *client.Client) error {
	drive, err := c.GetPersonalDrive()
	if err != nil {
		return fmt.Errorf("failed to get drive info: %w", err)
	}
	fmt.Printf("Drive ID:     %s\n", drive.ID)
	fmt.Printf("Drive Name:   %s\n", drive.Name)
	fmt.Printf("Drive Type:   %s\n", drive.DriveType)
	fmt.Printf("WebDAV URL:   %s\n", drive.Root.WebDavURL)
	fmt.Printf("Host:         %s\n", mapEmpty(c.HostOverride))
	fmt.Printf("IP:           %s\n", mapEmpty(c.ResolveIP))
	return nil
}

func mapEmpty(s string) string {
	if s == "" { return "(none)" }
	return s
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&uploadRemoteName, "name", "n", "", "Remote filename (default: same as local)")
	uploadCmd.Flags().StringVarP(&uploadMimeType, "mime", "m", "", "MIME type (default: application/octet-stream)")
	uploadCmd.Flags().Int64Var(&uploadChunkSize, "chunk-size", 5*1024*1024, "TUS chunk size (used if PUT fails)")
	uploadCmd.Flags().BoolVarP(&uploadVerbose, "verbose", "v", false, "Verbose debug output")
	uploadCmd.Flags().BoolVar(&uploadShowDriveInfo, "drive-info", false, "Show drive info (includes host/ip from config)")
	uploadCmd.Flags().StringVar(&uploadHostOverride, "host", "", "HTTP Host header (overrides config)")
	uploadCmd.Flags().StringVar(&uploadIPOverride, "ip", "", "DNS resolution IP (overrides config)")
}
