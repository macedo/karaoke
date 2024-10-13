package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/macedo/karaoke/spotify"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	client = &http.Client{
		Timeout: 2 * time.Second,
	}

	loginURL = fmt.Sprintf("%s/authorize", spotify.AccountsEndpoint)

	spotifyCmd = &cobra.Command{
		Use:   "spotify",
		Short: "",
	}

	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return loginAction(os.Stdout)
		},
	}
)

func loginAction(w io.Writer) error {
	cli, err := spotify.New(
		spotify.WithCredentialsPath(viper.ConfigFileUsed()),
	)
	if err != nil {
		return fmt.Errorf("failed to create spotify client: %v", err)
	}

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

	cli.Authorize(&spotify.AuthorizeInput{
		ClientID:     viper.GetString("spotify-client-id"),
		ResponseType: "code",
		Scope:        []string{"user-read-playback-state"},
		RedirectURI:  redirectURI,
		State:        state,
	})

	<-doneCh

	out, err := cli.Token(&spotify.TokenInputParams{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectURI: redirectURI,
	})
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	f, err := os.Create(cli.CredentialsPath())
	if err != nil {
		return err
	}

	if err := toml.NewEncoder(f).Encode(&spotify.Credentials{
		AccessToken:  out.AccessToken,
		ClientID:     viper.GetString("spotify-client-id"),
		ClientSecret: viper.GetString("spotify-client-secret"),
		RefreshToken: out.RefreshToken,
	}); err != nil {
		return err
	}

	fmt.Fprintln(w, "authenticated!!!")
	return nil
}

func setupServer(path string, handler func(http.ResponseWriter, *http.Request)) (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc(path, handler)

	server := &http.Server{
		Addr:    ":4321",
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

func init() {
	browser.Stderr = nil
	browser.Stdout = nil

	loginCmd.PersistentFlags().String("spotify-client-id", "", "spotify client id")
	viper.BindPFlag("spotify-client-id", loginCmd.PersistentFlags().Lookup("spotify-client-id"))

	loginCmd.PersistentFlags().String("spotify-client-secret", "", "spotify client secret")
	viper.BindPFlag("spotify-client-secret", loginCmd.PersistentFlags().Lookup("spotify-client-secret"))

	spotifyCmd.AddCommand(loginCmd)

	rootCmd.AddCommand(spotifyCmd)
}
