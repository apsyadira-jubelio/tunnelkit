package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with TunnelKit server",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		payload := map[string]string{
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("login failed: %s", body)
		}

		var result struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		// Save token to config file
		configDir, _ := os.UserHomeDir()
		configPath := configDir + "/.tunnelkit.yml"
		
		config := fmt.Sprintf("server: %s\ntoken: %s\n", serverURL, result.Token)
		if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("✓ Logged in successfully")
		fmt.Printf("  Config saved to: %s\n", configPath)
		return nil
	},
}

func init() {
	loginCmd.Flags().String("email", "", "Email address")
	loginCmd.Flags().String("password", "", "Password")
	loginCmd.MarkFlagRequired("email")
	loginCmd.MarkFlagRequired("password")
}
