package dmux

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type route struct {
	cmd     *discordgo.ApplicationCommand
	handler Handler
	routes  map[string]route
}

type Mux struct {
	session     *discordgo.Session
	middlewares []func(Handler) Handler

	commands map[string]*discordgo.ApplicationCommand
	handlers map[string]Handler
}

func NewMux(token string) *Mux {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(fmt.Errorf("dmux: %w", err))
	}

	return &Mux{
		session:     session,
		middlewares: []func(Handler) Handler{},
		handlers:    make(map[string]Handler),
		commands:    make(map[string]*discordgo.ApplicationCommand),
	}
}

func (m *Mux) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.chain(m.handlers[interactionToPattern(i.ApplicationCommandData())]).HandleInteraction(s, i)
}

func (m *Mux) Serve() error {
	m.session.AddHandler(m.HandleInteraction)

	err := m.session.Open()
	if err != nil {
		return err
	}

	cmds := make([]*discordgo.ApplicationCommand, 0, len(m.commands))
	for _, cmd := range m.commands {
		cmds = append(cmds, cmd)
	}
	_, err = m.session.ApplicationCommandBulkOverwrite(m.session.State.User.ID, "915247482722730014", cmds)

	return err
}

func (m *Mux) Handle(pattern string, handler Handler) {
	if m.handlers == nil {
		panic("dmux: mux not initialized")
	}

	m.commandFromPattern(pattern)

	m.handlers[normalize(pattern)] = handler
}

func (m *Mux) HandleFunc(pattern string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	m.Handle(pattern, HandlerFunc(handler))
}

func (m *Mux) Use(middlewares ...func(Handler) Handler) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *Mux) chain(handler Handler) Handler {
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}

	return handler
}

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
