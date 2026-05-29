package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var httpCmd = &cobra.Command{
	Use:   "http [port]",
	Short: "Create HTTP tunnel to local port",
	Args:  cobra.ExactArgs(1),
	RunE:  runHTTPTunnel,
}

type Config struct {
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
}

func loadConfig() (*Config, error) {
	homeDir, _ := os.UserHomeDir()
	data, err := os.ReadFile(homeDir + "/.tunnelkit.yml")
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run: tunnelkit login")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func runHTTPTunnel(cmd *cobra.Command, args []string) error {
	port := args[0]
	subdomain, _ := cmd.Flags().GetString("subdomain")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Printf("⚡ TunnelKit - Opening tunnel to localhost:%s\n", port)
	fmt.Printf("  Server: %s\n\n", cfg.Server)

	// Create tunnel via API
	payload := map[string]any{
		"name":      fmt.Sprintf("local-%s", port),
		"protocol":  "http",
		"subdomain": subdomain,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", cfg.Server+"/api/v1/tunnels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tunnel creation failed: %s", respBody)
	}

	var tunnelResp struct {
		ID        string `json:"id"`
		Subdomain string `json:"subdomain"`
	}
	json.NewDecoder(resp.Body).Decode(&tunnelResp)

	fmt.Printf("✓ Tunnel created!\n")
	fmt.Printf("  URL: http://%s.%s\n\n", tunnelResp.Subdomain, cfg.Server)

	// Connect via WebSocket
	wsURL, _ := url.Parse(cfg.Server)
	wsURL.Scheme = "ws"
	wsURL.Path = "/ws/agent"

	c, _, err := websocket.DefaultDialer.Dial(wsURL.String(), http.Header{
		"Authorization": []string{"Bearer " + cfg.Token},
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Send hello
	hello := map[string]any{
		"type":      "hello",
		"version":   "1.0",
		"tunnel_id": tunnelResp.ID,
	}
	if err := c.WriteJSON(hello); err != nil {
		return err
	}

	// Read hello response
	_, _, err = c.ReadMessage()
	if err != nil {
		return err
	}
	fmt.Printf("Connected to server. Ready to proxy requests.\n\n")

	// Setup yamux session over WebSocket
	session, err := yamux.Client(&wsConn{conn: c}, nil)
	if err != nil {
		return fmt.Errorf("failed to create multiplexer: %w", err)
	}

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	
	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.WriteJSON(map[string]string{"type": "ping"})
			case <-done:
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			stream, err := session.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					fmt.Printf("Accept error: %v\n", err)
					continue
				}
			}
			go handleStream(stream, port)
		}
	}()

	fmt.Println("Tunnel active. Press Ctrl+C to stop.")
	<-sigCh
	
	close(done)
	session.Close()
	wg.Wait()

	fmt.Println("\n✓ Tunnel closed")
	return nil
}

func init() {
	httpCmd.Flags().String("subdomain", "", "Custom subdomain")
}

func handleStream(stream interface{}, port string) {
	// TODO: implement actual HTTP forwarding
	// For now, just close the stream
	if closer, ok := stream.(io.Closer); ok {
		closer.Close()
	}
}

// wsConn wraps websocket.Conn to implement net.Conn for yamux
type wsConn struct {
	conn *websocket.Conn
}

func (c *wsConn) Read(b []byte) (n int, err error) {
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	n = copy(b, msg)
	return n, nil
}

func (c *wsConn) Write(b []byte) (n int, err error) {
	err = c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *wsConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *wsConn) SetDeadline(t time.Time) error {
	c.conn.SetReadDeadline(t)
	c.conn.SetWriteDeadline(t)
	return nil
}

func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
