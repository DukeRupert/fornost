package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dukerupert/fornost/pkg/hetzner"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var client *hetzner.Client

var rootCmd = &cobra.Command{
	Use:   "fornost",
	Short: "A CLI tool for managing Hetzner Cloud infrastructure",
	Long:  `Fornost inspects and manages Hetzner Cloud servers, SSH keys, and firewalls.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initClient()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initClient() error {
	// Try ~/.dotfiles/.env first, fall back to .env in current directory
	dotfilePath := filepath.Join(os.Getenv("HOME"), ".dotfiles", ".env")
	if _, err := os.Stat(dotfilePath); err == nil {
		_ = godotenv.Load(dotfilePath)
	} else {
		_ = godotenv.Load(".env")
	}

	token := os.Getenv("HETZNER_API_TOKEN")
	if token == "" {
		return fmt.Errorf("HETZNER_API_TOKEN must be set in ~/.dotfiles/.env or .env")
	}

	client = hetzner.NewClient(token)
	return nil
}
