package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"
	"os"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active tunnels",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("GET", cfg.Server+"/api/v1/tunnels", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var tunnels []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Protocol  string `json:"protocol"`
		Subdomain string `json:"subdomain"`
		Status    string `json:"status"`
	}

	if err := json.Unmarshal(body, &tunnels); err != nil {
		return err
	}

	if len(tunnels) == 0 {
		fmt.Println("No tunnels found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPROTOCOL\tSUBDOMAIN\tSTATUS")
	fmt.Fprintln(w, "----\t----\t--------\t---------\t------")

	for _, t := range tunnels {
		id := t.ID
		if len(id) > 8 {
			id = id[:8]
		}
		subdomain := t.Subdomain
		if subdomain == "" {
			subdomain = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			id, t.Name, t.Protocol, subdomain, t.Status)
	}

	w.Flush()
	return nil
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tunnel status",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TunnelKit v0.1.0")
		fmt.Println("Status: Ready")
		return nil
	},
}
