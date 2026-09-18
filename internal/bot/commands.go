package bot

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var prefix = func() string {
	if p := os.Getenv("BOT_PREFIX"); p != "" {
		return p
	}
	return "!"
}()

type cmd struct {
	name string
	desc string
	opts []*discordgo.ApplicationCommandOption
	run  func(*discordgo.Session, string, string, string) string
}

var commands = []*cmd{
	{"join", "Join your voice channel", nil, runJoin},
	{"play", "Play a song from YouTube", []*discordgo.ApplicationCommandOption{
		{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "URL or search term", Required: true},
	}, runPlay},
	{"skip", "Skip the current track", nil, runSkip},
	{"leave", "Leave and clear the queue", nil, runLeave},
}

func RegisterCommands(s *discordgo.Session) {
	cmds := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, c := range commands {
		cmds = append(cmds, &discordgo.ApplicationCommand{
			Name: c.name, Description: c.desc, Options: c.opts,
		})
	}
	if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", cmds); err != nil {
		log.Println("failed to register commands:", err)
	}
}

func OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	name := i.ApplicationCommandData().Name
	arg, _ := optionValue(i, "query")
	for _, c := range commands {
		if c.name != name {
			continue
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: c.run(s, interactionGuildID(i), interactionUserID(i), arg),
			},
		})
		return
	}
}

func OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	fields := strings.Fields(m.Content)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], prefix) {
		return
	}
	name := strings.TrimPrefix(fields[0], prefix)
	arg := strings.Join(fields[1:], " ")
	for _, c := range commands {
		if c.name != name {
			continue
		}
		s.ChannelMessageSend(m.ChannelID, c.run(s, m.GuildID, m.Author.ID, arg))
		return
	}
}

func interactionGuildID(i *discordgo.InteractionCreate) string {
	if i.GuildID != "" {
		return i.GuildID
	}
	return i.Interaction.GuildID
}

func interactionUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func optionValue(i *discordgo.InteractionCreate, name string) (string, bool) {
	for _, o := range i.ApplicationCommandData().Options {
		if o.Name == name {
			if v, ok := o.Value.(string); ok {
				return v, true
			}
		}
	}
	return "", false
}
