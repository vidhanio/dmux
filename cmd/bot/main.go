package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vidhanio/dmux"
	"github.com/vidhanio/dmux/middleware"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	mux := dmux.NewMux(os.Getenv("DISCORD_TOKEN"))

	mux.HandleFunc("/ping", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
	})

	mux.HandleFunc("/echo two layers boolean:boolean", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: strconv.FormatBool(dmux.CommandOption(i, "text").BoolValue()),
			},
		})
	})

	mux.HandleFunc("/echo one [text:string]", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: dmux.CommandOption(i, "text").StringValue(),
			},
		})
	})

	mux.Use(middleware.LoggerMiddleware)

	err = mux.Serve()
	if err != nil {
		log.Fatal().Err(err).Msg("Error serving")
	}

	log.Info().Msg("Serving. Press Ctrl+C to stop.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = mux.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("Error closing")
	}
}
