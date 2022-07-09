package whitelist

import (
	"errors"

	"github.com/nint8835/discord-whitelist/pkg/config"
)

type Provider interface {
	WhitelistUser(username string) error
}

type ProviderType string

const (
	ProviderTypeWhitelistHTTPAPI ProviderType = "Whitelist-HTTP-API"
)

var ErrInvalidProviderType = errors.New("specified provider type is invalid")

func NewFromConfig(config *config.Config) (Provider, error) {
	switch config.WhitelistProvider {
	case string(ProviderTypeWhitelistHTTPAPI):
		return NewWhitelistHTTPApiProvider(config), nil
	default:
		return nil, ErrInvalidProviderType
	}
}
