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

type optionInfo struct {
	name       string
	optionType discordgo.ApplicationCommandOptionType
	required   bool
}

func parseOption(option string) optionInfo {
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

	return optionInfo{
		name:       parts[0],
		optionType: optionType,
		required:   required,
	}
}

func normalize(pattern string) string {
	return strings.Join(strings.Fields(pattern), " ")
}

func parsePattern(pattern string) (string, []string, []optionInfo) {
	pattern = normalize(pattern)

	parts := strings.Fields(pattern)

	if len(parts) == 0 {
		panic("dmux: pattern must not be empty")
	}

	if parts[0][0] != '/' {
		panic("dmux: pattern must start with '/'")
	}

	command := parts[0][1:]
	subcmds := []string{}
	options := []optionInfo{}

	for _, part := range parts[1:] {
		if contains([]rune(part), ':') {
			options = append(options, parseOption(part))
		} else {
			if len(options) != 0 {
				panic("dmux: options must be the last part of the pattern")
			}

			subcmds = append(subcmds, part)

			if len(subcmds) > 2 {
				panic("dmux: subcommands can only be nested 2 layers deep")
			}
		}
	}

	return command, subcmds, options
}

func patternWithoutOptions(pattern string) string {
	command, subcmds, _ := parsePattern(pattern)
	subcmdString := strings.Join(subcmds, " ")

	return "/" + command + " " + subcmdString
}

func (m *Mux) commandFromPattern(pattern string) {
	command, subcmds, options := parsePattern(pattern)

	parent, ok := m.commands[command]
	if ok && len(subcmds) == 0 {
		panic("dmux: pattern already exists")
	}

	if !ok {
		parent = &discordgo.ApplicationCommand{
			Name:        command,
			Description: command,
			Type:        discordgo.ChatApplicationCommand,
		}

		m.commands[command] = parent
	}

	switch len(subcmds) {
	case 0:
		parent.Options = []*discordgo.ApplicationCommandOption{}

		for _, option := range options {
			parent.Options = append(parent.Options,
				&discordgo.ApplicationCommandOption{
					Name:        option.name,
					Description: option.name,
					Type:        option.optionType,
					Required:    option.required,
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

		for _, option := range options {
			subcmd.Options = append(subcmd.Options,
				&discordgo.ApplicationCommandOption{
					Name:        option.name,
					Description: option.name,
					Type:        option.optionType,
					Required:    option.required,
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

		for _, option := range options {
			subcmd.Options = append(subcmd.Options,
				&discordgo.ApplicationCommandOption{
					Name:        option.name,
					Description: option.name,
					Type:        option.optionType,
					Required:    option.required,
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

	if len(data.Options) == 0 {
		return normalize(builder.String())
	}

	switch data.Options[0].Type {
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		builder.WriteString(data.Options[0].Name)
		builder.WriteRune(' ')

		builder.WriteString(data.Options[0].Options[0].Name)
		builder.WriteRune(' ')
	case discordgo.ApplicationCommandOptionSubCommand:
		builder.WriteString(data.Options[0].Name)
		builder.WriteRune(' ')
	}

	return normalize(builder.String())
}
