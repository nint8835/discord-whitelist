package server

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/nint8835/discord-whitelist/pkg/config"
)

type Server struct {
	config *config.Config
	echo   *echo.Echo
}

func (server *Server) Start() error {
	return server.echo.Start(server.config.BindAddr)
}

func New(config *config.Config) *Server {
	echoInstance := echo.New()
	echoInstance.Use(session.Middleware(sessions.NewCookieStore([]byte(config.SecretKey))))

	return &Server{
		config: config,
		echo:   echoInstance,
	}
}
