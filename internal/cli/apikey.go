package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	apiKeyCmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage API keys",
	}
	rootCmd.AddCommand(apiKeyCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new API key",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreateAPIKey,
	}
	apiKeyCmd.AddCommand(createCmd)

	listKeysCmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		Args:  cobra.NoArgs,
		RunE:  runListAPIKeys,
	}
	apiKeyCmd.AddCommand(listKeysCmd)

	revokeCmd := &cobra.Command{
		Use:   "revoke <key-id>",
		Short: "Revoke an API key",
		Args:  cobra.ExactArgs(1),
		RunE:  runRevokeAPIKey,
	}
	apiKeyCmd.AddCommand(revokeCmd)
}

func findOrCreateVault(server, token, vaultName string) (string, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", server+"/api/v1/vaults", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var vaultsResp struct {
		Vaults []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"vaults"`
	}
	json.NewDecoder(resp.Body).Decode(&vaultsResp)
	resp.Body.Close()

	for _, v := range vaultsResp.Vaults {
		if v.Name == vaultName {
			return v.ID, nil
		}
	}

	createBody := map[string]string{"name": vaultName, "description": "Created by CLI"}
	createData, _ := json.Marshal(createBody)
	req2, _ := http.NewRequest("POST", server+"/api/v1/vaults", bytes.NewReader(createData))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	var created struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp2.Body).Decode(&created)
	return created.ID, nil
}

func runCreateAPIKey(cmd *cobra.Command, args []string) error {
	name := args[0]
	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)

	body := map[string]string{"name": name}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", server+"/api/v1/api-keys", strings.NewReader(string(data)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("create api key: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create api key failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	fmt.Printf("API Key created:\n")
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  Key: %v\n", result["full_key"])
	fmt.Printf("  ID: %v\n", result["api_key"].(map[string]interface{})["id"])
	fmt.Println("\nMake sure to copy the key now - you won't be able to see it again!")
	return nil
}

func runListAPIKeys(cmd *cobra.Command, args []string) error {
	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)

	req, _ := http.NewRequest("GET", server+"/api/v1/api-keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("list api keys: %w", err)
	}
	defer resp.Body.Close()

	var keys []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&keys)

	if len(keys) == 0 {
		fmt.Println("No API keys found")
		return nil
	}
	fmt.Println("API Keys:")
	for _, k := range keys {
		revoked := ""
		if r, ok := k["revoked"].(bool); ok && r {
			revoked = " [REVOKED]"
		}
		fmt.Printf("  %s (%s)%s\n", k["name"], k["id"], revoked)
	}
	return nil
}

func runRevokeAPIKey(cmd *cobra.Command, args []string) error {
	keyID := args[0]
	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)

	req, _ := http.NewRequest("PUT", server+"/api/v1/api-keys/"+keyID+"/revoke", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("API Key %s revoked\n", keyID)
	} else {
		return fmt.Errorf("revoke failed with status %d", resp.StatusCode)
	}
	return nil
}
