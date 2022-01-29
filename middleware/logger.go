package middleware

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/vidhanio/dmux"
)

func LoggerMiddleware(next dmux.Handler) dmux.Handler {
	return dmux.HandlerFunc(
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Debug().Str("user", i.Member.User.ID).Str("channel", i.ChannelID).Msg("interaction")
			next.HandleInteraction(s, i)
		},
	)
}
