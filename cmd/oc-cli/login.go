package main

import (
	"github.com/JammingBen/opencloud-skill-cli/internal/oidc"
	"github.com/spf13/cobra"
)

var (
	serverUrl     string
	clientID      string
	insecure      bool
	clipboard     bool
	loginHost     string
	loginIP       string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the OpenCloud server via OIDC",
	Long: `Authenticate via OIDC PKCE. Saves server URL, token, and connection
settings (--host, --ip, --insecure) to ~/.config/opencloud-cli/config.json.
These persist across commands so you don't need to repeat flags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := oidc.NewOIDCClient(serverUrl, insecure, clientID, clipboard)
		if err := c.Login(); err != nil {
			return err
		}
		// After login, save host/ip overrides to config if provided
		if loginHost != "" || loginIP != "" {
			cfg, err := oidc.LoadConfig()
			if err != nil {
				return err
			}
			if loginHost != "" {
				cfg.HostOverride = loginHost
			}
			if loginIP != "" {
				cfg.ResolveIP = loginIP
			}
			if err := cfg.Save(); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&serverUrl, "server-url", "s", "https://localhost:9200", "URL of the OpenCloud server")
	loginCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure TLS connections")
	loginCmd.Flags().StringVarP(&clientID, "client-id", "i", "", "OAuth2 Client ID")
	loginCmd.Flags().BoolVarP(&clipboard, "clipboard", "c", false, "Copy access token to clipboard")
	loginCmd.Flags().StringVar(&loginHost, "host", "", "HTTP Host header override (persisted to config)")
	loginCmd.Flags().StringVar(&loginIP, "ip", "", "DNS resolution override — connect to this IP (persisted to config)")
	loginCmd.MarkFlagRequired("server-url")
}
