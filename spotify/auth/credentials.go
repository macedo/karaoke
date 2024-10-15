package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
)

func LoadCredentials() (*oauth2.Token, error) {
	var token oauth2.Token

	f, err := os.Open(CredentialsFilename())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode credentials file: %w", err)
	}

	return &token, nil
}

func WriteCredentials(t *oauth2.Token) error {
	f, err := os.OpenFile(CredentialsFilename(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to json marshal token: %w", err)
	}

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func CredentialsFilename() string {
	return filepath.Join(UserHomeDir(), ".config", "karaoke", "credentials.json")
}

func UserHomeDir() string {
	home, _ := os.UserHomeDir()

	if len(home) > 0 {
		return home
	}

	u, _ := user.Current()
	if u != nil {
		home = u.HomeDir
	}

	return home
}
