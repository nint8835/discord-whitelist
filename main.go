package main

import (
	stdlog "log"

	"github.com/rs/zerolog/log"

	"github.com/nint8835/discord-whitelist/pkg/config"
	"github.com/nint8835/discord-whitelist/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		stdlog.Fatalf("error loading config: %w", err)
	}

	instance := server.New(cfg)

	err = instance.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running app")
	}
}
