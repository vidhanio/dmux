package dmux

import "github.com/bwmarrin/discordgo"

type Handler interface {
	HandleInteraction(*discordgo.Session, *discordgo.InteractionCreate)
}

type HandlerFunc func(*discordgo.Session, *discordgo.InteractionCreate)

func (hf HandlerFunc) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	hf(s, i)
}
