package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/vidhanio/dmux"
	"github.com/vidhanio/dmux/middleware"
)

func main() {
	godotenv.Load()

	mux := dmux.NewMux(os.Getenv("DISCORD_TOKEN"))

	mux.HandleFunc("/ping", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "pong",
			},
		})
	})

	mux.HandleFunc("/echo two layers text:boolean", func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: dmux.CommandOption(i, "text").StringValue(),
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

	err := mux.Serve()
	if err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
