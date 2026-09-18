package bot

import (
	"strconv"

	"github.com/bwmarrin/discordgo"

	"discoring/internal/player"
)

func userVoiceChannel(s *discordgo.Session, guildID, userID string) (string, bool) {
	g, err := s.State.Guild(guildID)
	if err != nil {
		return "", false
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == userID && vs.ChannelID != "" {
			return vs.ChannelID, true
		}
	}
	return "", false
}

func runJoin(s *discordgo.Session, guildID, userID, _ string) string {
	if guildID == "" {
		return "Commands are guild only!"
	}
	ch, ok := userVoiceChannel(s, guildID, userID)
	if !ok {
		return "You must be in a voice channel!"
	}
	vc, err := s.ChannelVoiceJoin(guildID, ch, false, false)
	if err != nil {
		return "Error joining: " + err.Error()
	}
	player.Get(guildID, vc)
	return "Joined your voice channel!"
}

func runPlay(_ *discordgo.Session, guildID, _, query string) string {
	if query == "" {
		return "Usage: !play <url or search term>"
	}
	p, ok := player.Find(guildID)
	if !ok {
		return "I'm not in a voice channel! Use `!join` or `/join` first."
	}
	pos := p.Add(&player.Track{Query: query})
	if p.Playing() {
		return "Added to queue! Position: " + strconv.Itoa(pos)
	}
	return "Now playing!"
}

func runSkip(_ *discordgo.Session, guildID, _, _ string) string {
	p, ok := player.Find(guildID)
	if !ok || !p.Playing() {
		return "Nothing is playing!"
	}
	p.Skip()
	return "Skipped!"
}

func runLeave(_ *discordgo.Session, guildID, _, _ string) string {
	p, ok := player.Find(guildID)
	if !ok {
		return "I'm not in a voice channel!"
	}
	p.StopAll()
	return "Left and cleared the queue!"
}