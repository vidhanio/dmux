package middleware

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/vidhanio/dmux"
)

func AdminMiddleware(next dmux.Handler) dmux.Handler {
	return dmux.HandlerFunc(
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
			if err != nil {
				log.Error().Err(err).Msg("failed to get user permissions")
				return
			}
			if perms&discordgo.PermissionAdministrator != 0 {
				next.HandleInteraction(s, i)
			} else {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You don't have permission to do that.",
					},
				})
			}
		},
	)
}
