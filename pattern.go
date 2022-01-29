package dmux

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var typeMap = map[discordgo.ApplicationCommandOptionType]string{
	discordgo.ApplicationCommandOptionString:      "string",
	discordgo.ApplicationCommandOptionInteger:     "integer",
	discordgo.ApplicationCommandOptionBoolean:     "boolean",
	discordgo.ApplicationCommandOptionUser:        "user",
	discordgo.ApplicationCommandOptionChannel:     "channel",
	discordgo.ApplicationCommandOptionRole:        "role",
	discordgo.ApplicationCommandOptionMentionable: "mentionable",
}

var reverseTypeMap = map[string]discordgo.ApplicationCommandOptionType{
	"string":      discordgo.ApplicationCommandOptionString,
	"integer":     discordgo.ApplicationCommandOptionInteger,
	"boolean":     discordgo.ApplicationCommandOptionBoolean,
	"user":        discordgo.ApplicationCommandOptionUser,
	"channel":     discordgo.ApplicationCommandOptionChannel,
	"role":        discordgo.ApplicationCommandOptionRole,
	"mentionable": discordgo.ApplicationCommandOptionMentionable,
}

var nameRe = regexp.MustCompile(`^[\w-]{1,32}$`)

func contains[T comparable](s []T, e T) bool {
	for _, c := range s {
		if c == e {
			return true
		}
	}
	return false
}

func parseOption(option string) (string, discordgo.ApplicationCommandOptionType, bool) {
	required := true
	if option[0] == '[' && option[len(option)-1] == ']' {
		required = false
		option = option[1 : len(option)-1]
	}

	parts := strings.SplitN(option, ":", 2)
	if len(parts) == 1 {
		panic("dmux: invalid option: " + option)
	}

	optionType, ok := reverseTypeMap[parts[1]]
	if !ok {
		panic("dmux: invalid option type: " + parts[1])
	}

	if !nameRe.MatchString(parts[0]) {
		panic("dmux: invalid option name: " + parts[0])
	}

	return parts[0], optionType, required
}

func normalize(pattern string) string {
	return strings.Join(strings.Fields(pattern), " ")
}

func (m *Mux) commandFromPattern(pattern string) {
	pattern = normalize(pattern)

	parts := strings.Fields(pattern)

	if len(parts) == 0 || parts[0][0] != '/' {
		panic("dmux: pattern must start with '/'")
	}

	parentName := parts[0][1:]

	parent, ok := m.commands[parentName]
	if ok && len(parts) == 1 {
		panic("dmux: pattern already exists")
	}

	if !ok {
		parent = &discordgo.ApplicationCommand{
			Name:        parentName,
			Description: parentName,
			Type:        discordgo.ChatApplicationCommand,
		}

		m.commands[parentName] = parent
	}

	subcmds := []string{}

	optionNames := []string{}
	optionTypes := []discordgo.ApplicationCommandOptionType{}
	optionRequires := []bool{}

	for _, part := range parts[1:] {
		if contains([]rune(part), ':') {
			name, optionType, required := parseOption(part)

			optionNames = append(optionNames, name)
			optionTypes = append(optionTypes, optionType)
			optionRequires = append(optionRequires, required)
		} else {
			if len(optionNames) != 0 {
				panic("dmux: options must be the last part of the pattern")
			}

			subcmds = append(subcmds, part)

			if len(subcmds) > 2 {
				panic("dmux: subcommands can only be nested 2 layers deep")
			}
		}
	}

	switch len(subcmds) {
	case 0:
		parent.Options = []*discordgo.ApplicationCommandOption{}

		for i, name := range optionNames {
			parent.Options = append(parent.Options,
				&discordgo.ApplicationCommandOption{
					Name:        name,
					Description: name,
					Type:        optionTypes[i],
					Required:    optionRequires[i],
				},
			)
		}
	case 1:
		var subcmd *discordgo.ApplicationCommandOption

		for _, option := range parent.Options {
			if option.Name == subcmds[0] && option.Type == discordgo.ApplicationCommandOptionSubCommand {
				subcmd = option

				break
			}
		}

		if subcmd == nil {
			subcmd = &discordgo.ApplicationCommandOption{
				Name:        subcmds[0],
				Description: subcmds[0],
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			}

			parent.Options = append(parent.Options, subcmd)
		}

		subcmd.Options = []*discordgo.ApplicationCommandOption{}

		for i, name := range optionNames {
			subcmd.Options = append(subcmd.Options,
				&discordgo.ApplicationCommandOption{
					Name:        name,
					Description: name,
					Type:        optionTypes[i],
					Required:    optionRequires[i],
				},
			)
		}
	case 2:
		var group *discordgo.ApplicationCommandOption

		for _, option := range parent.Options {
			if option.Name == subcmds[0] && option.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
				group = option

				break
			}
		}

		if group == nil {
			group = &discordgo.ApplicationCommandOption{
				Name:        subcmds[0],
				Description: subcmds[0],
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			}

			parent.Options = append(parent.Options, group)
		}

		var subcmd *discordgo.ApplicationCommandOption

		for _, option := range group.Options {
			if option.Name == subcmds[1] && option.Type == discordgo.ApplicationCommandOptionSubCommand {
				subcmd = option

				break
			}
		}

		if subcmd == nil {
			subcmd = &discordgo.ApplicationCommandOption{
				Name:        subcmds[1],
				Description: subcmds[1],
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			}

			group.Options = append(group.Options, subcmd)
		}

		subcmd.Options = []*discordgo.ApplicationCommandOption{}

		for i, name := range optionNames {
			subcmd.Options = append(subcmd.Options,
				&discordgo.ApplicationCommandOption{
					Name:        name,
					Description: name,
					Type:        optionTypes[i],
					Required:    optionRequires[i],
				},
			)
		}
	default:
		panic("dmux: too many subcommands")
	}
}

func interactionToPattern(data discordgo.ApplicationCommandInteractionData) string {
	builder := &strings.Builder{}

	builder.WriteRune('/')
	builder.WriteString(data.Name)
	builder.WriteRune(' ')

	for _, option := range data.Options {
		builder.WriteString(option.Name)
		if option.Type != discordgo.ApplicationCommandOptionSubCommand && option.Type != discordgo.ApplicationCommandOptionSubCommandGroup {
			builder.WriteRune(':')
			builder.WriteString(typeMap[option.Type])
		}
		builder.WriteRune(' ')

		if option.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			builder.WriteString(option.Options[0].Name)
			builder.WriteRune(' ')

			for _, suboption := range option.Options[0].Options {
				builder.WriteString(suboption.Name)
				builder.WriteRune(':')
				builder.WriteString(typeMap[suboption.Type])
				builder.WriteRune(' ')
			}
		}

		if option.Type == discordgo.ApplicationCommandOptionSubCommand {
			for _, suboption := range option.Options {
				builder.WriteString(suboption.Name)
				builder.WriteRune(':')
				builder.WriteString(typeMap[suboption.Type])
				builder.WriteRune(' ')
			}
		}
	}

	return normalize(builder.String())
}
