package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate and store access token",
		Args:  cobra.NoArgs,
		RunE:  runLogin,
	}
	loginCmd.Flags().StringP("email", "e", "", "email address")
	loginCmd.Flags().StringP("password", "p", "", "password")

	rootCmd.AddCommand(loginCmd)

	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Create a new account",
		Args:  cobra.NoArgs,
		RunE:  runRegister,
	}
	registerCmd.Flags().StringP("email", "e", "", "email address")
	registerCmd.Flags().StringP("password", "p", "", "password")

	rootCmd.AddCommand(registerCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	if email == "" {
		fmt.Print("Email: ")
		fmt.Scanln(&email)
	}
	if password == "" {
		fmt.Print("Password: ")
		fmt.Fscanln(os.Stdin, &password)
	}

	server, _ := cmd.Flags().GetString("server")

	body := map[string]string{"email": email, "password": password}
	data, _ := json.Marshal(body)

	resp, err := http.Post(server+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	token, _ := result["access_token"].(string)
	fmt.Printf("Login successful!\n")
	fmt.Printf("Access token: %s\n", token)

	// Save token to config file
	os.WriteFile(configPath(), []byte(token), 0600)

	return nil
}

func runRegister(cmd *cobra.Command, args []string) error {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")

	if email == "" || password == "" {
		return fmt.Errorf("email and password are required")
	}

	server, _ := cmd.Flags().GetString("server")

	body := map[string]string{"email": email, "password": password}
	data, _ := json.Marshal(body)

	resp, err := http.Post(server+"/api/v1/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("register failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("register failed (%d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Account created for %s\n", email)
	return nil
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.vault_token"
}
