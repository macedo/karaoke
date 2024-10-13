package spotify

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/schema"
)

func (c *Client) Token(params *TokenInputParams) (*TokenOutputParams, error) {
	var out TokenOutputParams

	if params == nil {
		params = &TokenInputParams{}
	}

	body := url.Values{}
	if err := schema.NewEncoder().Encode(params, body); err != nil {
		return nil, err
	}

	tokenURL, _ := url.Parse(AccountsEndpoint)
	tokenURL.Path = "/api/token"

	request, err := http.NewRequest(http.MethodPost, tokenURL.String(), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", c.basicAuth())
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err = c.sendRequest(request, &out)

	return &out, err
}

type TokenInputParams struct {
	Code        string `schema:"code"`
	GrantType   string `schema:"grant_type"`
	RedirectURI string `schema:"redirect_uri"`
}

type TokenOutputParams struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
