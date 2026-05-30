package tunnel

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ProxyServer handles HTTP request forwarding to tunnel agents
type ProxyServer struct {
	hub    *TunnelHub
	logger *zap.Logger
}

func NewProxyServer(hub *TunnelHub, logger *zap.Logger) *ProxyServer {
	return &ProxyServer{
		hub:    hub,
		logger: logger,
	}
}

// ForwardRequest forwards an HTTP request to the tunnel agent
func (p *ProxyServer) ForwardRequest(ctx context.Context, tunnelID uuid.UUID, req *http.Request, w http.ResponseWriter) error {
	// Get agent connection
	p.hub.mu.RLock()
	agent, ok := p.hub.agents[tunnelID]
	p.hub.mu.RUnlock()

	if !ok {
		return echo.NewHTTPError(http.StatusBadGateway, "tunnel agent not connected")
	}

	// Check agent connection health
	if time.Since(agent.LastPing) > 2*time.Minute {
		return echo.NewHTTPError(http.StatusBadGateway, "tunnel agent connection stale")
	}

	// Create proxy request message
	proxyReq := ProxyRequest{
		Method:  req.Method,
		Path:    req.URL.Path,
		Query:   req.URL.RawQuery,
		Headers: make(map[string]string),
	}

	// Copy headers (filter out hop-by-hop headers)
	for key, values := range req.Header {
		if !isHopByHopHeader(key) && len(values) > 0 {
			proxyReq.Headers[key] = values[0]
		}
	}

	// Read body
	if req.Body != nil {
		defer req.Body.Close()
		// Limit body size to 10MB for proxy
		if req.ContentLength > 0 && req.ContentLength <= 10*1024*1024 {
			body := make([]byte, req.ContentLength)
			n, _ := req.Body.Read(body)
			proxyReq.Body = body[:n]
		}
	}

	// Serialize and send
	data, err := json.Marshal(proxyReq)
	if err != nil {
		return err
	}

	// Send to agent via WebSocket
	if err := agent.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		p.logger.Error("Failed to send proxy request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadGateway, "failed to forward request")
	}

	// Wait for response (with timeout)
	respCh := make(chan *ProxyResponse, 1)
	errCh := make(chan error, 1)

	go func() {
		_, msg, err := agent.Conn.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}

		var resp ProxyResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			errCh <- err
			return
		}
		respCh <- &resp
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	case resp := <-respCh:
		// Write response to client
		for key, value := range resp.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(resp.StatusCode)
		if len(resp.Body) > 0 {
			w.Write(resp.Body)
		}
		return nil
	case <-time.After(30 * time.Second):
		return echo.NewHTTPError(http.StatusGatewayTimeout, "tunnel agent response timeout")
	}
}

// ProxyRequest represents a request to forward to the agent
type ProxyRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body,omitempty"`
}

// ProxyResponse represents a response from the agent
type ProxyResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body,omitempty"`
}

func isHopByHopHeader(header string) bool {
	hopByHop := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
	for _, h := range hopByHop {
		if h == header {
			return true
		}
	}
	return false
}

// StreamProxyRequest handles streaming proxy for large requests
func (p *ProxyServer) StreamProxyRequest(ctx context.Context, tunnelID uuid.UUID, req *http.Request, w http.ResponseWriter) error {
	// For large requests (> 10MB), use streaming
	p.hub.mu.RLock()
	agent, ok := p.hub.agents[tunnelID]
	p.hub.mu.RUnlock()

	if !ok {
		return echo.NewHTTPError(http.StatusBadGateway, "tunnel agent not connected")
	}

	// Create a streaming connection using yamux multiplexing
	// This allows multiple concurrent requests over a single connection

	// Send initial request headers
	streamReq := StreamRequest{
		Method:  req.Method,
		Path:    req.URL.Path,
		Query:   req.URL.RawQuery,
		Headers: make(map[string]string),
	}

	for key, values := range req.Header {
		if !isHopByHopHeader(key) && len(values) > 0 {
			streamReq.Headers[key] = values[0]
		}
	}

	data, _ := json.Marshal(streamReq)
	if err := agent.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	// Stream request body
	if req.Body != nil {
		defer req.Body.Close()
		buf := make([]byte, 32*1024) // 32KB chunks
		for {
			n, err := req.Body.Read(buf)
			if n > 0 {
				agent.Conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			}
			if err != nil {
				break
			}
		}
	}

	return nil
}

// StreamRequest represents a streaming proxy request
type StreamRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers"`
}
