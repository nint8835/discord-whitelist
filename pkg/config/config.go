package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type Config struct {
	SecretKey string `default:"correct-horse-battery-staple" split_words:"true"`
	BindAddr  string `default:":12500" split_words:"true"`

	WhitelistProvider string `default:"Whitelist-HTTP-API" split_words:"true"`
	WhitelistApiUrl   string `default:"http://localhost:8500" split_words:"true"`
	WhitelistApiToken string `split_words:"true"`

	LogLevel       string `split_words:"true" default:"debug"`
	UseJsonLogging bool   `split_words:"true" default:"false"`

	DiscordClientId     string `split_words:"true"`
	DiscordClientSecret string `split_words:"true"`
	DiscordCallbackUrl  string `default:"http://localhost:12500/callback" split_words:"true"`
	DiscordGuildId      string `split_words:"true"`

	OAuth2Config *oauth2.Config
}

var loadedConfig *Config

func Load() (*Config, error) {
	if loadedConfig != nil {
		return loadedConfig, nil
	}

	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Msg("Error loading .env file")
	}

	var newConfig Config

	err = envconfig.Process("discord_whitelist", &newConfig)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(newConfig.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error parsing provided log level: %w", err)
	}

	zerolog.SetGlobalLevel(logLevel)

	if !newConfig.UseJsonLogging {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	newConfig.OAuth2Config = &oauth2.Config{
		ClientID:     newConfig.DiscordClientId,
		ClientSecret: newConfig.DiscordClientSecret,
		Scopes:       []string{"identify", "guilds"},
		RedirectURL:  newConfig.DiscordCallbackUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}

	log.Debug().Msg("Config loaded.")

	loadedConfig = &newConfig

	return loadedConfig, nil
}
