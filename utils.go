package dmux

import "github.com/bwmarrin/discordgo"

func optionsSlice(i *discordgo.InteractionCreate) []*discordgo.ApplicationCommandInteractionDataOption {
	data := i.ApplicationCommandData()

	if len(data.Options) == 0 {
		return nil
	}

	switch data.Options[0].Type {
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		return data.Options[0].Options[0].Options
	case discordgo.ApplicationCommandOptionSubCommand:
		return data.Options[0].Options
	default:
		return data.Options
	}
}

func CommandOption(i *discordgo.InteractionCreate, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range optionsSlice(i) {
		if option.Name == name {
			return option
		}
	}

	return nil
}
