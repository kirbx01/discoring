package player

import (
	"errors"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var players sync.Map

type Track struct {
	Query string
}

var (
	errSkipped = errors.New("track skipped")
	errStopped = errors.New("playback stopped")
)

type Player struct {
	guildID string
	vc      *discordgo.VoiceConnection
	stop    chan struct{}
	wake    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	queue   []*Track
	live    bool
}

func Get(guildID string, vc *discordgo.VoiceConnection) *Player {
	if p, ok := players.Load(guildID); ok {
		return p.(*Player)
	}
	p := &Player{
		guildID: guildID,
		vc:      vc,
		stop:    make(chan struct{}, 1),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	players.Store(guildID, p)
	go p.run()
	return p
}

func Find(guildID string) (*Player, bool) {
	p, ok := players.Load(guildID)
	if !ok {
		return nil, false
	}
	return p.(*Player), true
}

func (p *Player) run() {
	for {
		select {
		case <-p.done:
			return
		case <-p.wake:
		}
		for {
			t := p.next()
			if t == nil {
				break
			}
			p.setLive(true)
			err := stream(p, t)
			p.setLive(false)
			if err == errStopped {
				return
			}
			if err != nil && err != errSkipped {
				log.Printf("[%s] track failed: %v", p.guildID, err)
			}
		}
	}
}

func (p *Player) Add(t *Track) (pos int) {
	p.mu.Lock()
	p.queue = append(p.queue, t)
	pos = len(p.queue)
	p.mu.Unlock()
	select {
	case p.wake <- struct{}{}:
	default:
	}
	return pos
}

func (p *Player) next() *Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.queue) == 0 {
		return nil
	}
	t := p.queue[0]
	p.queue = p.queue[1:]
	return t
}

func (p *Player) Playing() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.live
}

func (p *Player) setLive(b bool) {
	p.mu.Lock()
	p.live = b
	p.mu.Unlock()
}

func (p *Player) Skip() {
	select {
	case p.stop <- struct{}{}:
	default:
	}
}

func (p *Player) StopAll() {
	close(p.done)
	p.mu.Lock()
	p.queue = nil
	p.mu.Unlock()
	p.vc.Speaking(false)
	p.vc.Disconnect()
	p.vc.Close()
	players.Delete(p.guildID)
}
