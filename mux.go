package dmux

import (
	"fmt"
	"os"

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
	_, err = m.session.ApplicationCommandBulkOverwrite(m.session.State.User.ID, os.Getenv("DISCORD_GUILD_ID"), cmds)

	return err
}

func (m *Mux) Close() error {
	return m.session.Close()
}

func (m *Mux) Handle(pattern string, handler Handler) {
	if m.handlers == nil {
		panic("dmux: mux not initialized")
	}

	m.commandFromPattern(pattern)

	m.handlers[patternWithoutOptions(pattern)] = handler
}

func (m *Mux) HandleFunc(pattern string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	m.Handle(pattern, HandlerFunc(handler))
}

func (m *Mux) Use(middlewares ...func(Handler) Handler) {
	m.middlewares = append(m.middlewares, middlewares...)
}
