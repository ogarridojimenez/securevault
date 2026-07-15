package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vault",
	Short: "SecureVault CLI - Secrets management",
	Long: `A professional secrets management CLI for SecureVault.
Store, retrieve, and manage your secrets with ease.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default $HOME/.vault.yaml)")
	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "SecureVault server URL")
	rootCmd.PersistentFlags().StringP("token", "t", "", "access token (or $VAULT_TOKEN)")
}
