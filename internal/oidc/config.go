package oidc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

type Config struct {
	ServerURL     string        `json:"server_url"`
	Token         *oauth2.Token `json:"token"`
	Insecure      bool          `json:"insecure"`
	ClientID      string        `json:"client_id"`
	TokenEndpoint string        `json:"token_endpoint"`
	HostOverride  string        `json:"host,omitempty"`
	ResolveIP     string        `json:"ip,omitempty"`
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "opencloud-cli", "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) Clear() error {
	c.Token = nil
	c.ServerURL = ""
	c.Insecure = false
	c.ClientID = ""
	c.TokenEndpoint = ""
	c.HostOverride = ""
	c.ResolveIP = ""
	return c.Save()
}

func (c *Config) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	if c.Token == nil || c.ClientID == "" || c.TokenEndpoint == "" {
		t, _ := os.LookupEnv("OC_ACCESS_TOKEN")
		if t != "" {
			return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t}), nil
		}
		return nil, nil
	}
	if c.Insecure {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}
	conf := &oauth2.Config{
		ClientID: c.ClientID,
		Endpoint: oauth2.Endpoint{
			TokenURL: c.TokenEndpoint,
		},
	}
	ts := conf.TokenSource(ctx, c.Token)
	return &savingTokenSource{ts: ts, cfg: c}, nil
}

type savingTokenSource struct {
	ts  oauth2.TokenSource
	cfg *Config
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.ts.Token()
	if err != nil {
		return nil, err
	}
	if s.cfg.Token == nil || token.AccessToken != s.cfg.Token.AccessToken {
		s.cfg.Token = token
		if err := s.cfg.Save(); err != nil {
			return nil, err
		}
	}
	return token, nil
}
