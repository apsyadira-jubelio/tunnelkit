package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var tcpCmd = &cobra.Command{
	Use:   "tcp [port]",
	Short: "Create TCP tunnel to local port",
	Args:  cobra.ExactArgs(1),
	RunE:  runTCPTunnel,
}

func runTCPTunnel(cmd *cobra.Command, args []string) error {
	port := args[0]
	remotePort, _ := cmd.Flags().GetInt("remote-port")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Printf("⚡ TunnelKit - TCP tunnel to localhost:%s\n", port)

	payload := map[string]any{
		"name":        fmt.Sprintf("tcp-%s", port),
		"protocol":    "tcp",
		"remote_port": remotePort,
	}

	tunnelID, err := createTunnel(cfg, payload)
	if err != nil {
		return err
	}

	fmt.Printf("✓ TCP tunnel created!\n")
	fmt.Printf("  Tunnel ID: %s\n", tunnelID)
	fmt.Printf("  Remote Port: %d\n\n", remotePort)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", remotePort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", remotePort, err)
	}
	defer listener.Close()

	fmt.Printf("Listening on port %d...\n\n", remotePort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleTCPConnection(conn, port)
		}
	}()

	<-sigCh
	fmt.Println("\n✓ TCP tunnel closed")
	return nil
}

func handleTCPConnection(clientConn net.Conn, targetPort string) {
	defer clientConn.Close()

	target, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", targetPort))
	if err != nil {
		fmt.Printf("Failed to connect to target: %v\n", err)
		return
	}
	defer target.Close()

	go func() {
		io.Copy(target, clientConn)
	}()
	io.Copy(clientConn, target)
}

func init() {
	tcpCmd.Flags().Int("remote-port", 0, "Remote port to expose")
}

func createTunnel(cfg *Config, payload map[string]any) (string, error) {
	// TODO: implement actual API call
	return "placeholder-tunnel-id", nil
}
