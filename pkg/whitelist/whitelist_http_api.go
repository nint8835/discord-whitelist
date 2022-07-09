package whitelist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nint8835/discord-whitelist/pkg/config"
)

type WhitelistHTTPApiProvider struct {
	url   string
	token string
}

func (provider *WhitelistHTTPApiProvider) WhitelistUser(username string) error {
	reqBody, err := json.Marshal(map[string]string{
		"name": username,
	})
	if err != nil {
		return fmt.Errorf("error encoding body: %w", err)
	}

	req, _ := http.NewRequest(http.MethodPost, provider.url, bytes.NewReader(reqBody))
	req.Header.Set("Authorization", fmt.Sprintf("WHA %s", provider.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("got unexpected status %d", resp.StatusCode)
	}
	return nil
}

var _ Provider = (*WhitelistHTTPApiProvider)(nil)

func NewWhitelistHTTPApiProvider(config *config.Config) *WhitelistHTTPApiProvider {
	return &WhitelistHTTPApiProvider{
		url:   config.WhitelistApiUrl,
		token: config.WhitelistApiToken,
	}
}
