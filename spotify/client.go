package spotify

import (
	"context"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/macedo/karaoke/internal/config"
	"github.com/macedo/karaoke/spotify/auth"
	"golang.org/x/oauth2"
)

type Client struct {
	oAuthHandler  *auth.OAuthHandler
	token         *oauth2.Token
	logger        *log.Logger
	http          *http.Client
	authenticated bool
}

func New(config *config.AppConfig, l *log.Logger) (*Client, error) {
	storedToken, err := auth.LoadCredentials()
	if err != nil {
		return nil, err
	}
	if storedToken != nil {
		l.Debug("loaded oauth credentials from file")
	} else {
		l.Debug("oauth credentials file not found")
	}

	return &Client{
		oAuthHandler:  auth.NewOAuthHandler(config, l),
		authenticated: false,
		logger:        l,
		token:         storedToken,
	}, nil
}

func (c *Client) Authenticate(ctx context.Context) error {
	if c.token == nil {
		token, err := c.oAuthHandler.AuthorizationCodeFlow(ctx)
		if err != nil {
			return err
		}

		if err := auth.WriteCredentials(c.token); err != nil {
			return err
		}
		c.logger.Debug("saved credentials to disk")

		c.token = token
		c.logger.Debug("got oauth credentials")
	} else {
		now := time.Now()
		if now.After(c.token.Expiry) {
			c.logger.Debug("token expired", "expiry", c.token.Expiry)

			newToken, err := c.oAuthHandler.RefreshToken(ctx, c.token)
			if err != nil {
				c.logger.Info("saved credentials expired, we need to reauthenticate..", "error", err)
				c.authenticated = false

				return nil
			}

			if err := auth.WriteCredentials(c.token); err != nil {
				return err
			}
			c.logger.Debug("saved credentials to disk")

			c.token = newToken
			c.logger.Debug("refreshed oauth credentials using the refresh token", "expiry", newToken.Expiry)
		}

		c.authenticated = true

		c.logger.Info("authenticated")
	}

	c.authenticated = true
	c.http = c.oAuthHandler.Client(ctx, c.token)

	return nil
}
