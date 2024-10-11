package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
)

const (
	spotifyURL = "https://accounts.spotify.com"
)

var (
	client = &http.Client{
		Timeout: 2 * time.Second,
	}

	loginURL = fmt.Sprintf("%s/authorize", spotifyURL)
	tokenURL = fmt.Sprintf("%s/api/token", spotifyURL)
)

func init() {
	browser.Stderr = nil
	browser.Stdout = nil
}

func main() {
	code := ""
	state := fmt.Sprint(rand.Int())

	doneCh := make(chan bool)

	redirectURI, shutdown := setupServer(
		"/oauth/callback",
		func(w http.ResponseWriter, r *http.Request) {
			if s, ok := r.URL.Query()["state"]; ok && s[0] == state {
				if c, ok := r.URL.Query()["code"]; ok && len(c) == 1 {
					code = c[0]
				}
			}
			doneCh <- true
			io.WriteString(w, "You can close this page and return to your CLI")
		})
	defer shutdown()

	u, _ := url.Parse(loginURL)
	query := u.Query()
	query.Add("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
	query.Add("response_type", "code")
	query.Add("redirect_uri", redirectURI)
	query.Add("scope", "user-read-currently-playing")
	query.Add("state", state)
	u.RawQuery = query.Encode()

	if err := browser.OpenURL(u.String()); err != nil {
		log.Fatal(err)
	}

	<-doneCh

	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("code", code)
	values.Add("redirect_uri", redirectURI)

	data, err := doPostRequest(tokenURL, values)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}

func doPostRequest(path string, values url.Values) ([]byte, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	auth := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(os.Getenv("SPOTIFY_CLIENT_ID")+":"+os.Getenv("SPOTIFY_CLIENT_SECRET"))))

	request.Header.Add("Authorization", auth)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func setupServer(path string, callbackHandler func(http.ResponseWriter, *http.Request)) (string, func()) {
	addr := ":4321"

	mux := http.NewServeMux()
	mux.HandleFunc(path, callbackHandler)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	return fmt.Sprintf("http://localhost%s%s", server.Addr, path), func() {
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
}
