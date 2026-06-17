package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/JammingBen/opencloud-skill-cli/internal/client"
	"github.com/JammingBen/opencloud-skill-cli/internal/logger"
	"github.com/JammingBen/opencloud-skill-cli/internal/oidc"
	"github.com/spf13/cobra"
)

var (
	path        string
	method      string
	body        string
	queryParams []string
	verbose     bool
	statusOnly  bool
	jsonFormat  bool
	apiHost     string
	apiIP       string
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Interact with the OpenCloud LibreGraph API",
	Long:  "This command allows you to interact with the OpenCloud LibreGraph API.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logLvl := slog.LevelInfo
		if verbose {
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

		params := url.Values{}
		for _, qp := range queryParams {
			parts := strings.SplitN(qp, "=", 2)
			if len(parts) == 2 {
				params.Add(parts[0], parts[1])
			} else {
				params.Add(parts[0], "")
			}
		}

		c := client.NewClient(cfg.ServerURL, cfg.Insecure, ts)
		// CLI flags override config
		c.HostOverride = firstNonEmpty(apiHost, cfg.HostOverride)
		c.ResolveIP = firstNonEmpty(apiIP, cfg.ResolveIP)

		resp, err := c.MakeRequest(path, method, body, params)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}

		encodingFormat := client.TOON
		if jsonFormat {
			encodingFormat = client.JSON
		}
		e := client.NewEncoder(encodingFormat)

		if statusOnly {
			output, err := e.EncodeStatusCode(resp.StatusCode)
			if err != nil {
				return fmt.Errorf("failed to encode status code: %w", err)
			}
			fmt.Println(output)
			return nil
		}

		output, err := e.EncodeBody(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to encode response body: %w", err)
		}
		fmt.Println(output)
		return nil
	},
}

func firstNonEmpty(a, b string) string {
	if a != "" { return a }
	return b
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringVarP(&path, "path", "p", "", "Path of the API endpoint")
	apiCmd.Flags().StringVarP(&method, "method", "m", "GET", "HTTP method to use")
	apiCmd.Flags().StringVarP(&body, "body", "b", "", "JSON body to send with the request")
	apiCmd.Flags().StringArrayVarP(&queryParams, "query", "q", []string{}, "Query parameters to add to the request (e.g. -q key=value)")
	apiCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	apiCmd.Flags().BoolVar(&statusOnly, "status-only", false, "Only return the status code")
	apiCmd.Flags().BoolVar(&jsonFormat, "json-format", false, "Encode the output in JSON format")
	apiCmd.Flags().StringVar(&apiHost, "host", "", "HTTP Host header (overrides config)")
	apiCmd.Flags().StringVar(&apiIP, "ip", "", "DNS resolution IP (overrides config)")
	apiCmd.MarkFlagRequired("path")
}
