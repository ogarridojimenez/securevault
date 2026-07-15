package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	setCmd := &cobra.Command{
		Use:   "set <path> <value>",
		Short: "Set a secret value (e.g. vault set production/db/pwd mypassword)",
		Args:  cobra.ExactArgs(2),
		RunE:  runSet,
	}
	rootCmd.AddCommand(setCmd)

	getCmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Get a secret value",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}
	rootCmd.AddCommand(getCmd)

	listCmd := &cobra.Command{
		Use:   "list <vault-path>",
		Short: "List secrets in a vault",
		Args:  cobra.ExactArgs(1),
		RunE:  runList,
	}
	rootCmd.AddCommand(listCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE:  runDelete,
	}
	rootCmd.AddCommand(deleteCmd)
}

func splitPath(p string) (vaultName, secretName string) {
	parts := strings.SplitN(p, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func getToken(cmd *cobra.Command) string {
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		return token
	}

	// Try config file
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.vault_token")
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}

	return ""
}

func getServer(cmd *cobra.Command) string {
	s, _ := cmd.Flags().GetString("server")
	return s
}

func runSet(cmd *cobra.Command, args []string) error {
	path := args[0]
	value := args[1]
	vaultName, secretName := splitPath(path)
	if secretName == "" {
		return fmt.Errorf("path must be vault-name/secret-name, got: %s", path)
	}

	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in. Run 'vault login' first")
	}
	server := getServer(cmd)

	// Find or create vault
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server+"/api/v1/vaults", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("list vaults: %w", err)
	}

	var vaultsResp struct {
		Vaults []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"vaults"`
		Total int `json:"total"`
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(body, &vaultsResp)

	var vaultID string
	for _, v := range vaultsResp.Vaults {
		if v.Name == vaultName {
			vaultID = v.ID
			break
		}
	}

	if vaultID == "" {
		// Create vault
		createBody := map[string]string{"name": vaultName, "description": "Created by CLI"}
		createData, _ := json.Marshal(createBody)
		req2, _ := http.NewRequest("POST", server+"/api/v1/vaults", bytes.NewReader(createData))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		resp2, err := client.Do(req2)
		if err != nil {
			return fmt.Errorf("create vault: %w", err)
		}
		defer resp2.Body.Close()

		var createdVault struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		json.NewDecoder(resp2.Body).Decode(&createdVault)
		vaultID = createdVault.ID
	}

	// Create secret
	secretBody := map[string]string{"name": secretName, "value": value}
	secretData, _ := json.Marshal(secretBody)
	req3, _ := http.NewRequest("POST", server+"/api/v1/vaults/"+vaultID+"/secrets", bytes.NewReader(secretData))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := client.Do(req3)
	if err != nil {
		return fmt.Errorf("set secret: %w", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode == http.StatusCreated {
		fmt.Printf("Secret %s set successfully\n", path)
	} else {
		body, _ := io.ReadAll(resp3.Body)
		return fmt.Errorf("set secret failed (%d): %s", resp3.StatusCode, string(body))
	}

	return nil
}

func runGet(cmd *cobra.Command, args []string) error {
	path := args[0]
	vaultName, secretName := splitPath(path)
	if secretName == "" {
		return fmt.Errorf("path must be vault-name/secret-name, got: %s", path)
	}

	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)

	// Find vault ID by name
	vaultID, err := findVaultID(server, token, vaultName)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("GET", server+"/api/v1/vaults/"+vaultID+"/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}

	var secretsResp struct {
		Secrets []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"secrets"`
	}
	json.NewDecoder(resp.Body).Decode(&secretsResp)
	resp.Body.Close()

	var secretID string
	for _, s := range secretsResp.Secrets {
		if s.Name == secretName {
			secretID = s.ID
			break
		}
	}
	if secretID == "" {
		return fmt.Errorf("secret %s not found", path)
	}

	req2, _ := http.NewRequest("GET", server+"/api/v1/vaults/"+vaultID+"/secrets/"+secretID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := client.Do(req2)
	if err != nil {
		return fmt.Errorf("get secret: %w", err)
	}
	defer resp2.Body.Close()

	var secretResp struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	json.NewDecoder(resp2.Body).Decode(&secretResp)
	fmt.Println(secretResp.Value)

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	vaultName := args[0]
	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)

	vaultID, err := findVaultID(server, token, vaultName)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("GET", server+"/api/v1/vaults/"+vaultID+"/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}
	defer resp.Body.Close()

	var secretsResp struct {
		Secrets []struct {
			Name    string `json:"name"`
			Version int    `json:"version"`
		} `json:"secrets"`
		Total int `json:"total"`
	}
	json.NewDecoder(resp.Body).Decode(&secretsResp)

	if len(secretsResp.Secrets) == 0 {
		fmt.Println("No secrets found")
		return nil
	}
	fmt.Printf("Secrets in %s:\n", vaultName)
	for _, s := range secretsResp.Secrets {
		fmt.Printf("  %s (v%d)\n", s.Name, s.Version)
	}
	return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
	path := args[0]
	vaultName, secretName := splitPath(path)
	if secretName == "" {
		return fmt.Errorf("path must be vault-name/secret-name, got: %s", path)
	}

	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)

	vaultID, err := findVaultID(server, token, vaultName)
	if err != nil {
		return err
	}

	// Find secret ID
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server+"/api/v1/vaults/"+vaultID+"/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("list secrets: %w", err)
	}

	var secretsResp struct {
		Secrets []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"secrets"`
	}
	json.NewDecoder(resp.Body).Decode(&secretsResp)
	resp.Body.Close()

	var secretID string
	for _, s := range secretsResp.Secrets {
		if s.Name == secretName {
			secretID = s.ID
			break
		}
	}
	if secretID == "" {
		return fmt.Errorf("secret %s not found", path)
	}

	req2, _ := http.NewRequest("DELETE", server+"/api/v1/vaults/"+vaultID+"/secrets/"+secretID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := client.Do(req2)
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	resp2.Body.Close()

	if resp2.StatusCode == http.StatusNoContent {
		fmt.Printf("Secret %s deleted\n", path)
	} else {
		return fmt.Errorf("delete failed with status %d", resp2.StatusCode)
	}

	return nil
}

func findVaultID(server, token, vaultName string) (string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server+"/api/v1/vaults", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("list vaults: %w", err)
	}
	defer resp.Body.Close()

	var vaultsResp struct {
		Vaults []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"vaults"`
	}
	json.NewDecoder(resp.Body).Decode(&vaultsResp)

	for _, v := range vaultsResp.Vaults {
		if v.Name == vaultName {
			return v.ID, nil
		}
	}
	return "", fmt.Errorf("vault %s not found", vaultName)
}
