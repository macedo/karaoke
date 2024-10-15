package spotify

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/macedo/karaoke/internal/config"
	"github.com/macedo/karaoke/spotify/auth"
	"golang.org/x/oauth2"
)

type Client struct {
	oAuthHandler  *auth.OAuthHandler
	token         *oauth2.Token
	logger        *log.Logger
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

		c.token = token
		c.logger.Debug("got oauth credentials")
		if err := auth.WriteCredentials(token); err != nil {
			return err
		}
		c.logger.Debug("saved credentials to disk")

		c.authenticated = true

		fmt.Println(color.GreenString("Authenticated"))

		return nil
	}

	now := time.Now()
	if now.After(c.token.Expiry) {
		c.logger.Debug("token expired", "expiry", c.token.Expiry)

		newToken, err := c.oAuthHandler.RefreshToken(ctx, c.token)
		if err != nil {
			c.logger.Info("saved credentials expired, we need to reauthenticate..", "error", err)
			c.authenticated = false

			return nil
		}

		c.token = newToken
		c.logger.Debug("refreshed oauth credentials using the refresh token", "expiry", newToken.Expiry)

		if err := auth.WriteCredentials(newToken); err != nil {
			return err
		}
		c.logger.Debug("saved credentials to disk")
	}

	c.authenticated = true

	fmt.Println(color.GreenString("Authenticated"))

	return nil
}
