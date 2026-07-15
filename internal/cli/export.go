package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export all secrets as JSON",
		Args:  cobra.NoArgs,
		RunE:  runExport,
	}
	rootCmd.AddCommand(exportCmd)

	importCmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import secrets from JSON file",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runImport,
	}
	rootCmd.AddCommand(importCmd)
}

type exportData struct {
	Vaults []exportVault `json:"vaults"`
}

type exportVault struct {
	Name    string         `json:"name"`
	Secrets []exportSecret `json:"secrets"`
}

type exportSecret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func runExport(cmd *cobra.Command, args []string) error {
	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)
	client := &http.Client{}

	// Get all vaults
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
	}
	json.NewDecoder(resp.Body).Decode(&vaultsResp)
	resp.Body.Close()

	var export exportData
	for _, v := range vaultsResp.Vaults {
		ev := exportVault{Name: v.Name}

		req2, _ := http.NewRequest("GET", server+"/api/v1/vaults/"+v.ID+"/secrets", nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		resp2, err := client.Do(req2)
		if err != nil {
			return fmt.Errorf("list secrets for vault %s: %w", v.Name, err)
		}

		var secretsResp struct {
			Secrets []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"secrets"`
		}
		json.NewDecoder(resp2.Body).Decode(&secretsResp)
		resp2.Body.Close()

		for _, s := range secretsResp.Secrets {
			req3, _ := http.NewRequest("GET", server+"/api/v1/vaults/"+v.ID+"/secrets/"+s.ID, nil)
			req3.Header.Set("Authorization", "Bearer "+token)
			resp3, err := client.Do(req3)
			if err != nil {
				return fmt.Errorf("get secret %s: %w", s.Name, err)
			}

			var secretResp struct {
				Value string `json:"value"`
			}
			json.NewDecoder(resp3.Body).Decode(&secretResp)
			resp3.Body.Close()

			ev.Secrets = append(ev.Secrets, exportSecret{
				Name:  s.Name,
				Value: secretResp.Value,
			})
		}

		export.Vaults = append(export.Vaults, ev)
	}

	data, _ := json.MarshalIndent(export, "", "  ")
	fmt.Println(string(data))
	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	token := getToken(cmd)
	if token == "" {
		return fmt.Errorf("not logged in")
	}
	server := getServer(cmd)
	client := &http.Client{}

	// Read input
	var data []byte
	var err error
	if len(args) > 0 {
		data, err = os.ReadFile(args[0])
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	var importData exportData
	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	for _, v := range importData.Vaults {
		// Find or create vault
		vaultID, err := findOrCreateVault(server, token, v.Name)
		if err != nil {
			return fmt.Errorf("ensure vault %s: %w", v.Name, err)
		}

		for _, s := range v.Secrets {
			body := map[string]string{"name": s.Name, "value": s.Value}
			bodyData, _ := json.Marshal(body)

			req, _ := http.NewRequest("POST", server+"/api/v1/vaults/"+vaultID+"/secrets", strings.NewReader(string(bodyData)))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("import secret %s/%s: %w", v.Name, s.Name, err)
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				fmt.Printf("Imported %s/%s\n", v.Name, s.Name)
			} else if resp.StatusCode == http.StatusConflict {
				fmt.Printf("Skipped %s/%s (already exists)\n", v.Name, s.Name)
			}
		}
	}

	return nil
}
