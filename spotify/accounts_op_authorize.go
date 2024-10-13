package spotify

import (
	"log"
	"net/url"
	"strings"

	"github.com/pkg/browser"
)

func (c *Client) Authorize(params *AuthorizeInput) {
	authorizeURL, _ := url.Parse(AccountsEndpoint)
	authorizeURL.Path = "/authorize"
	query := authorizeURL.Query()
	query.Add("client_id", params.ClientID)
	query.Add("response_type", params.ResponseType)
	query.Add("redirect_uri", params.RedirectURI)
	query.Add("scope", strings.Join(params.Scope, ","))
	query.Add("state", params.State)
	authorizeURL.RawQuery = query.Encode()

	if err := browser.OpenURL(authorizeURL.String()); err != nil {
		log.Fatal(err)
	}
}

type AuthorizeInput struct {
	ClientID     string
	ResponseType string
	RedirectURI  string
	State        string
	Scope        []string
}
