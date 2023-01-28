package aliyundrive

import (
	"net/http"
)

type Drive struct {
	driveId      string
	tokenManager TokenManager
	httpClient   *http.Client
}
type optionFunc func(c *Drive)

func New(options ...optionFunc) *Drive {
	c := new(Drive)
	c.SetOption(options...)
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c
}

func WithDriveId(driveId string) optionFunc {
	return func(c *Drive) {
		c.driveId = driveId
	}
}

func WithHttpClient(httpClient *http.Client) optionFunc {
	return func(c *Drive) {
		c.httpClient = httpClient
	}
}

func WithTokenManager(tokenManager TokenManager) optionFunc {
	return func(c *Drive) {
		c.tokenManager = tokenManager
	}
}

func (c *Drive) SetOption(options ...optionFunc) *Drive {
	for _, setOption := range options {
		setOption(c)
	}
	return c
}
