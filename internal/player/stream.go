package player

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"os/exec"

	"layeh.com/gopus"
)

const (
	frameMs  = 20
	rate     = 48000
	channels = 2
	bitrate  = 256000

	frameSize   = rate * frameMs / 1000
	frameInts   = frameSize * channels
	frameBytes  = frameInts * 2
	opusMaxData = 4096
)

func stream(p *Player, t *Track) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	select {
	case <-p.stop:
	default:
	}

	go func() {
		select {
		case <-p.done:
			cancel()
		case <-ctx.Done():
		}
	}()

	yt := exec.CommandContext(ctx, "yt-dlp",
		"-f", "bestaudio",
		"--no-playlist",
		"--quiet", "--no-warnings",
		"-o", "-", t.Query)
	ytOut, err := yt.StdoutPipe()
	if err != nil {
		return err
	}

	ff := exec.CommandContext(ctx, "ffmpeg",
		"-v", "error",
		"-i", "pipe:0",
		"-vn", "-ac", "2", "-ar", "48000",
		"-f", "s16le", "pipe:1")
	ff.Stdin = ytOut
	pcmOut, err := ff.StdoutPipe()
	if err != nil {
		return err
	}

	if err := ff.Start(); err != nil {
		return err
	}
	if err := yt.Start(); err != nil {
		ff.Process.Kill()
		return err
	}

	enc, err := gopus.NewEncoder(rate, channels, gopus.Audio)
	if err != nil {
		return err
	}
	enc.SetBitrate(bitrate)

	p.vc.Speaking(true)
	defer p.vc.Speaking(false)

	rd := bufio.NewReaderSize(pcmOut, 64*1024)
	buf := make([]byte, frameBytes)
	pcm := make([]int16, frameInts)

	for {
		select {
		case <-p.stop:
			return errSkipped
		case <-p.done:
			return errStopped
		default:
		}

		if _, err := io.ReadFull(rd, buf); err != nil {
			break
		}
		for i := range pcm {
			pcm[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
		}
		data, err := enc.Encode(pcm, frameSize, opusMaxData)
		if err != nil {
			continue
		}

		select {
		case p.vc.OpusSend <- data:
		case <-p.stop:
			return errSkipped
		case <-p.done:
			return errStopped
		}
	}

	yt.Wait()
	ff.Wait()
	return nil
}
