package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/macedo/karaoke/internal/config"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type AuthoriztionResponse struct {
	Error error
	Code  string
}

type OAuthHandler struct {
	config      *oauth2.Config
	logger      *log.Logger
	redirectURI url.URL
}

func NewOAuthHandler(config *config.AppConfig, l *log.Logger) *OAuthHandler {
	redirectURI := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("localhost", config.Spotify.RedirectURIPort),
		Path:   "/oauth/callback",
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.Spotify.ClientID,
		ClientSecret: config.Spotify.ClientSecret,
		Scopes:       config.Spotify.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.Spotify.AuthURL,
			TokenURL: endpoints.Spotify.TokenURL,
		},
		RedirectURL: redirectURI.String(),
	}

	return &OAuthHandler{
		config:      oauthConfig,
		logger:      l,
		redirectURI: redirectURI,
	}
}

func (h *OAuthHandler) AuthorizationCodeFlow(ctx context.Context) (*oauth2.Token, error) {
	code := ""
	state := fmt.Sprint(rand.Int())

	doneCh := make(chan bool)

	http.HandleFunc(h.redirectURI.Path, func(w http.ResponseWriter, r *http.Request) {
		if s, ok := r.URL.Query()["state"]; ok && s[0] == state {
			code = r.URL.Query().Get("code")

			msg := "<p><strong>Success!</strong></p>"
			msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
			fmt.Fprint(w, msg)

			doneCh <- true

			return
		}
	})

	server := &http.Server{Addr: h.redirectURI.Host}

	go func() {
		okToClose := <-doneCh
		if okToClose {
			h.logger.Debug("shutdown server")
			if err := server.Shutdown(ctx); err != nil {
				log.Fatal("server shutdown failed", "error", err)
			}
		}
	}()

	u := h.config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println(color.CyanString("You will now be taken to your browser for authentication"))
	if err := browser.OpenURL(u); err != nil {
		fmt.Println("browser did not open, please authenticate manually.")
		fmt.Println(u)
	}

	h.logger.Debug("starting server", "hostname", h.redirectURI.Hostname())
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("server listen failure", "error", err)
	}

	return h.config.Exchange(ctx, code)
}

func (h *OAuthHandler) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return h.config.Client(ctx, token)
}

func (h *OAuthHandler) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	src := h.config.TokenSource(ctx, token)

	return src.Token()
}
