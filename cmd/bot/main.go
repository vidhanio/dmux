package main

import (
	"os"
	"os/signal"
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

	mux := dmux.NewMux(os.Getenv("DISCORD_TOKEN"), os.Getenv("DISCORD_GUILD_ID"))

	mux.HandleFunc("/hi happy:string<happy,sad>", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		happy := dmux.Option(i, "happy").StringValue()

		happyString := "How can I help you?"
		if happy == "happy" {
			happyString = "Good to hear!"
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: happyString,
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
