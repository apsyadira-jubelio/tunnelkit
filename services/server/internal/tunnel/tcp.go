package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TCPServer handles TCP tunnel connections
type TCPServer struct {
	hub        *TunnelHub
	tunnelRepo interface{} // domain.TunnelRepository
	logger     *zap.Logger
	listeners  map[int]*net.Listener
	mu         sync.RWMutex
}

func NewTCPServer(hub *TunnelHub, logger *zap.Logger) *TCPServer {
	return &TCPServer{
		hub:       hub,
		logger:    logger,
		listeners: make(map[int]*net.Listener),
	}
}

// StartPortListener starts listening on a specific port for TCP traffic
func (s *TCPServer) StartPortListener(ctx context.Context, port int, tunnelID uuid.UUID) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	s.mu.Lock()
	s.listeners[port] = &listener
	s.mu.Unlock()

	s.logger.Info("TCP listener started", zap.Int("port", port), zap.String("tunnel_id", tunnelID.String()))

	go func() {
		<-ctx.Done()
		listener.Close()
		s.mu.Lock()
		delete(s.listeners, port)
		s.mu.Unlock()
	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					s.logger.Error("TCP accept error", zap.Error(err))
					continue
				}
			}
			go s.handleConnection(conn, tunnelID)
		}
	}()

	return nil
}

func (s *TCPServer) handleConnection(client net.Conn, tunnelID uuid.UUID) {
	defer client.Close()

	// Get tunnel agent connection
	s.mu.RLock()
	agent, ok := s.hub.agents[tunnelID]
	s.mu.RUnlock()

	if !ok {
		s.logger.Warn("No agent connected for tunnel", zap.String("tunnel_id", tunnelID.String()))
		return
	}

	// Forward to agent via yamux (simplified - in production would use actual yamux stream)
	// For now, just log the connection
	s.logger.Info("TCP connection received",
		zap.String("client", client.RemoteAddr().String()),
		zap.String("tunnel_id", tunnelID.String()),
	)

	// Simulate proxying by just copying data
	// In real implementation, would open a yamux stream to agent
	_ = agent

	// Keep connection alive
	buffer := make([]byte, 4096)
	for {
		n, err := client.Read(buffer)
		if err != nil {
			if err != io.EOF {
				s.logger.Error("TCP read error", zap.Error(err))
			}
			return
		}

		// Forward data
		// In production: agent.Conn.Write(buffer[:n])
		_ = buffer[:n]
	}
}

// StopAllListeners stops all TCP listeners
func (s *TCPServer) StopAllListeners() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for port, listener := range s.listeners {
		if listener != nil {
			(*listener).Close()
			s.logger.Info("TCP listener stopped", zap.Int("port", port))
		}
	}
}

// HandleTCPRequest handles incoming TCP connection for the tunnel server
func (s *TCPServer) HandleTCPRequest(c echo.Context) error {
	// WebSocket upgrade for TCP tunnel
	return c.JSON(200, map[string]string{"status": "tcp tunnel ready"})
}

// GetListenerPorts returns all active listener ports
func (s *TCPServer) GetListenerPorts() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ports := make([]int, 0, len(s.listeners))
	for port := range s.listeners {
		ports = append(ports, port)
	}
	return ports
}
