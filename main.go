package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/macedo/karaoke/internal/config"
	"github.com/macedo/karaoke/spotify"
)

var cfgFile string

func init() {
	flag.StringVar(&cfgFile, "config", "", "path to config file")
}

func main() {
	flag.Parse()

	appConfig, err := config.NewFromFile(cfgFile)
	if err != nil {
		log.Fatal("failed to read config file", "error", err)
	}

	level, err := log.ParseLevel(appConfig.LogLevel)
	if err != nil {
		log.Fatal("could not parse log level", "error", err)
	}
	log.SetLevel(level)
	log.SetTimeFormat(time.Kitchen)
	if level == log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.Debug(
		"config file loaded",
		"spotify_client_id", appConfig.Spotify.ClientID,
		"spotify_client_secret", appConfig.Spotify.ClientSecret,
		"spotify_scopes", appConfig.Spotify.Scopes,
	)

	cli, err := spotify.New(appConfig, log.With())
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.Authenticate(context.Background()); err != nil {
		log.Fatal(err)
	}

	out, _ := cli.PlayingTrack()

	name := out.Item.Name
	artist := out.Item.Artists[0].Name

	duration := out.Item.DurationMs * int(time.Millisecond)
	progress := out.ProgressMs * int(time.Millisecond)

	fmt.Println(color.YellowString("%s - %s | %s/%s", name, artist, progress, duration))

	t := time.NewTimer(time.Duration(duration) - time.Duration(progress))

	<-t.C

	time.Sleep(2 * time.Second)
	out, _ = cli.PlayingTrack()

	name = out.Item.Name
	artist = out.Item.Artists[0].Name

	fmt.Println(color.YellowString("%s - %s | %s", name, artist, progress))
}
