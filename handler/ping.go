package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Ping struct {
}

const PingPath = "/ping"
const pongResponse = "pong"

func NewPingHandler() *Ping {
	return &Ping{}
}

func (v *Ping) GetPingHandler(c echo.Context) error {
	return c.String(http.StatusOK, pongResponse)
}

func (v *Ping) AddHandlers(e *echo.Echo) {
	e.GET(PingPath, v.GetPingHandler)
}
