package server

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	glog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"

	"github.com/nint8835/discord-whitelist/pkg/config"
	"github.com/nint8835/discord-whitelist/pkg/whitelist"
)

//go:embed static
var staticFS embed.FS

type errorPageContext struct {
	Code    int
	Message string
}

type whitelistPageContext struct {
	Message string
}

type Server struct {
	config *config.Config
	echo   *echo.Echo

	whitelistProvider whitelist.Provider
}

func (server *Server) Start() error {
	return server.echo.Start(server.config.BindAddr)
}

func (server *Server) HandleIndex(c echo.Context) error {
	sess := getSession(c)
	if _, discordAuthenticated := sess.Values["discordAuthenticated"]; !discordAuthenticated {
		return c.Redirect(http.StatusTemporaryRedirect, server.config.OAuth2Config.AuthCodeURL("state"))
	}

	var tmplContext whitelistPageContext

	if c.Request().Method == http.MethodPost {
		username := c.FormValue("username")
		err := server.whitelistProvider.WhitelistUser(username)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error whitelisting user: %s", err))
		}
		tmplContext.Message = fmt.Sprintf("%s whitelisted successfully!", username)
	}

	return c.Render(http.StatusOK, "whitelist.gohtml", tmplContext)
}

func (server *Server) HandleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if state != "state" {
		return echo.NewHTTPError(http.StatusForbidden, "Invalid state")
	}

	token, err := server.config.OAuth2Config.Exchange(c.Request().Context(), code)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Invalid code")
	}

	discordClient, _ := discordgo.New(fmt.Sprintf("Bearer %s", token.AccessToken))

	guilds, err := discordClient.UserGuilds(100, "", "")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error listing guilds: %s", err))
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

	return echo.NewHTTPError(http.StatusForbidden, "You are not in the required server")
}

func (server *Server) HandleHTTPError(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := err.Error()
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	c.Render(code, "error.gohtml", errorPageContext{
		Code:    code,
		Message: message,
	})
}

func New(config *config.Config) (*Server, error) {
	whitelistProvider, err := whitelist.NewFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating whitelist provider: %w", err)
	}

	echoInstance := echo.New()
	echoInstance.Renderer = NewEmbeddedTemplater()

	logger := lecho.From(log.Logger, lecho.WithLevel(glog.INFO))
	echoInstance.Logger = logger
	echoInstance.Use(lecho.Middleware(lecho.Config{Logger: logger}))
	echoInstance.Use(session.Middleware(sessions.NewCookieStore([]byte(config.SecretKey))))

	server := &Server{
		config:            config,
		echo:              echoInstance,
		whitelistProvider: whitelistProvider,
	}

	echoInstance.HTTPErrorHandler = server.HandleHTTPError

	echoInstance.GET("/", server.HandleIndex)
	echoInstance.POST("/", server.HandleIndex)
	echoInstance.GET("/callback", server.HandleCallback)
	echoInstance.GET("/static/*", echo.WrapHandler(http.FileServer(http.FS(staticFS))))

	return server, nil
}
