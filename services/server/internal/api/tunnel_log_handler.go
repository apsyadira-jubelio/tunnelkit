package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tunnelkit/services/server/internal/domain"
)

type TunnelLogHandler struct {
	logRepo domain.TunnelLogRepository
}

func NewTunnelLogHandler(logRepo domain.TunnelLogRepository) *TunnelLogHandler {
	return &TunnelLogHandler{logRepo: logRepo}
}

func (h *TunnelLogHandler) List(c echo.Context) error {
	tunnelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tunnel id")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 50
	}

	logs, err := h.logRepo.List(c.Request().Context(), tunnelID, offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch logs")
	}

	return c.JSON(http.StatusOK, logs)
}

// Stream logs via SSE
func (h *TunnelLogHandler) Stream(c echo.Context) error {
	tunnelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tunnel id")
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	ctx := c.Request().Context()
	logCh := h.logRepo.Stream(ctx, tunnelID, time.Now())

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case log, ok := <-logCh:
			if !ok {
				return nil
			}
			// Format as SSE
		
data, _ := json.Marshal(log)
			c.Response().Write([]byte("data: "))
			c.Response().Write(data)
			c.Response().Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}
