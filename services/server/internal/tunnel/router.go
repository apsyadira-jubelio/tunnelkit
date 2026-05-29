package tunnel

import (
	"net/http"
	"strings"
	"github.com/labstack/echo/v4"
	"github.com/tunnelkit/services/server/internal/domain"
)

// SubdomainRouter handles routing public subdomain requests to tunnel agents
type SubdomainRouter struct {
	hub        *TunnelHub
	tunnelRepo domain.TunnelRepository
	baseDomain string
}

func NewSubdomainRouter(hub *TunnelHub, tunnelRepo domain.TunnelRepository, baseDomain string) *SubdomainRouter {
	return &SubdomainRouter{
		hub:        hub,
		tunnelRepo: tunnelRepo,
		baseDomain: baseDomain,
	}
}

func (r *SubdomainRouter) ServeHTTP(c echo.Context) error {
	host := c.Request().Host
	// Extract subdomain: myapp.tunnel.example.com -> myapp
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid host")
	}

	subdomain := parts[0]

	// Lookup tunnel by subdomain
	tunnel, err := r.tunnelRepo.GetBySubdomain(c.Request().Context(), subdomain)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}

	if tunnel.Status != "active" {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "tunnel is inactive")
	}

	// Forward request to agent
	resp, err := r.hub.ForwardToAgent(tunnel.ID, c.Request())
	if err != nil {
		return err
	}
	if resp != nil {
		// Copy response to client
		for k, v := range resp.Header {
			c.Response().Header().Set(k, strings.Join(v, ","))
		}
		c.Response().WriteHeader(resp.StatusCode)
		// resp.Body -> c.Response()
	}

	return nil
}
