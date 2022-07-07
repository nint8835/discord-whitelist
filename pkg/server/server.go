package server

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/nint8835/discord-whitelist/pkg/config"
)

type Server struct {
	config *config.Config
	echo   *echo.Echo
}

func (server *Server) Start() error {
	return server.echo.Start(server.config.BindAddr)
}

func (server *Server) HandleIndex(c echo.Context) error {
	sess := getSession(c)
	if _, discordAuthenticated := sess.Values["discordAuthenticated"]; !discordAuthenticated {
		return c.Redirect(http.StatusTemporaryRedirect, server.config.OAuth2Config.AuthCodeURL("state"))
	}

	return nil
}

func (server *Server) HandleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if state != "state" {
		// TODO: better error page
		return c.JSON(http.StatusForbidden, "Invalid state")
	}

	token, err := server.config.OAuth2Config.Exchange(c.Request().Context(), code)
	if err != nil {
		// TODO: better error page
		return c.String(http.StatusForbidden, "Invalid code")
	}

	discordClient, _ := discordgo.New(fmt.Sprintf("Bearer %s", token.AccessToken))

	guilds, err := discordClient.UserGuilds(100, "", "")
	if err != nil {
		log.Error().Err(err).Msg("Error listing user servers")
		// TODO: better error page
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	for _, guild := range guilds {
		if guild.ID != server.config.DiscordGuildId {
			continue
		}

		sess := getSession(c)
		sess.Values["discordAuthenticated"] = true
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	return c.String(http.StatusForbidden, "You are not in the required server")
}

func New(config *config.Config) *Server {
	echoInstance := echo.New()
	echoInstance.Use(session.Middleware(sessions.NewCookieStore([]byte(config.SecretKey))))

	server := &Server{
		config: config,
		echo:   echoInstance,
	}

	echoInstance.GET("/", server.HandleIndex)
	echoInstance.GET("/callback", server.HandleCallback)

	return server
}
