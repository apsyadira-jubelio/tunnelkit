package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	apiKey    string
	configFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tunnelkit",
		Short: "TunnelKit CLI - expose local services to the internet",
		Long: `TunnelKit CLI creates encrypted tunnels from your local machine 
to the TunnelKit server, allowing external access to local services.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "https://tunnel.localhost", "TunnelKit server URL")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "token", "t", "", "API authentication token")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path (default: ~/.tunnelkit.yml)")

	// Add commands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(tcpCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
