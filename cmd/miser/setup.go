package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Waxmard/miser/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure external integrations",
}

var setupSimpleFinCmd = &cobra.Command{
	Use:   "simplefin <setup-token>",
	Short: "Exchange a SimpleFIN setup token for an access URL",
	Long: `Claim a SimpleFIN setup token and save the access URL to your config.

Get a setup token at https://beta-bridge.simplefin.org/
This token can only be claimed once.`,
	Args: cobra.ExactArgs(1),
	RunE: runSetupSimpleFin,
}

func init() {
	setupCmd.AddCommand(setupSimpleFinCmd)
	rootCmd.AddCommand(setupCmd)
}

func runSetupSimpleFin(_ *cobra.Command, args []string) error {
	setupToken := args[0]

	// Base64-decode the setup token to get the claim URL.
	claimURL, err := base64.StdEncoding.DecodeString(setupToken)
	if err != nil {
		return fmt.Errorf("invalid setup token (not valid base64): %w", err)
	}

	fmt.Println("Claiming SimpleFIN access token...")

	// POST to the claim URL to get the access URL.
	resp, err := http.Post(string(claimURL), "", nil) //nolint:noctx // one-shot CLI command
	if err != nil {
		return fmt.Errorf("claim request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("setup token already claimed or invalid — generate a new one at https://beta-bridge.simplefin.org")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	accessURL := strings.TrimSpace(string(body))
	if accessURL == "" {
		return fmt.Errorf("received empty access URL")
	}

	// Write the access URL to the config file.
	configPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	if err := updateSimpleFinConfig(configPath, accessURL); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	fmt.Printf("Access URL saved to %s\n", configPath)
	fmt.Println("SimpleFIN is now enabled. Run `miser sync simplefin` to pull transactions.")
	return nil
}

func updateSimpleFinConfig(path, accessURL string) error {
	cfg := &config.Config{}
	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, cfg); err != nil {
			return fmt.Errorf("read existing config: %w", err)
		}
	}

	cfg.SimpleFin.Enabled = true
	cfg.SimpleFin.AccessURL = accessURL

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer func() { _ = f.Close() }()

	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}
