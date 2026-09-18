package bot

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"discoring/internal/player"
)

var (
	errNotInChannel = errors.New("user not in a voice channel")

	joinAttempts = 3
	joinMu       sync.Map
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

// joinLock serializes join attempts per guild. Sending multiple OP4 voice
// state updates concurrently (e.g. !join fired while /join is still
// handshaking) races discordgo's handshake and ends in "timeout waiting
// for voice".
func joinLock(guildID string) *sync.Mutex {
	mu, _ := joinMu.LoadOrStore(guildID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// voiceConn returns the current VoiceConnection for a guild, if any.
func voiceConn(s *discordgo.Session, guildID string) (*discordgo.VoiceConnection, bool) {
	s.RLock()
	defer s.RUnlock()
	vc, ok := s.VoiceConnections[guildID]
	return vc, ok
}

// disconnectVoice drops any stale/garbage VoiceConnection before retrying.
// A timed-out ChannelVoiceJoin leaves a closed (but not removed) connection
// behind, which would otherwise be reused by the next attempt.
func disconnectVoice(s *discordgo.Session, guildID string) {
	if vc, ok := voiceConn(s, guildID); ok && vc != nil {
		vc.Disconnect()
	}
}

func joinVoice(s *discordgo.Session, guildID, userID string) error {
	ch, ok := userVoiceChannel(s, guildID, userID)
	if !ok {
		return errNotInChannel
	}
	if guildID == "" {
		return errNotInChannel
	}

	mu := joinLock(guildID)
	mu.Lock()
	defer mu.Unlock()

	var firstErr error
	for i := 0; i < joinAttempts; i++ {
		disconnectVoice(s, guildID)
		vc, err := s.ChannelVoiceJoin(guildID, ch, false, false)
		if err == nil {
			if vc != nil {
				vc.LogLevel = s.LogLevel
			}
			player.Get(guildID, vc)
			return nil
		}
		if firstErr == nil {
			firstErr = err
		}
		if i+1 < joinAttempts {
			time.Sleep(2 * time.Second)
		}
	}
	return firstErr
}

func runJoin(s *discordgo.Session, guildID, userID, _ string) string {
	if guildID == "" {
		return "Commands are guild only!"
	}
	if err := joinVoice(s, guildID, userID); err != nil {
		if errors.Is(err, errNotInChannel) {
			return "You must be in a voice channel!"
		}
		return "Error joining: " + err.Error() +
			"\nIf this keeps happening, check that Discord voice (UDP) isn't blocked on this network and try another voice region."
	}
	return "Joined your voice channel!"
}

func runPlay(s *discordgo.Session, guildID, userID, query string) string {
	if query == "" {
		return "Usage: !play <url or search term>"
	}
	p, ok := player.Find(guildID)
	if !ok {
		// Auto-join, so a plain !play just works after joining a voice channel.
		if err := joinVoice(s, guildID, userID); err != nil {
			return "I'm not in a voice channel! Use `!join` or `/join` first."
		}
		p, _ = player.Find(guildID)
	}
	if p == nil {
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

func runLeave(s *discordgo.Session, guildID, _, _ string) string {
	p, ok := player.Find(guildID)
	if !ok {
		disconnectVoice(s, guildID)
		return "I'm not in a voice channel!"
	}
	p.StopAll()
	return "Left and cleared the queue!"
}