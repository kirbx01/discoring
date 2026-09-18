# discoring

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white&style=flat)
![Language](https://img.shields.io/badge/language-Go-00ADD8?style=flat)
![Discord bot](https://img.shields.io/badge/Discord-music%20bot-5865F2?logo=discord&logoColor=white&style=flat)
![FFmpeg](https://img.shields.io/badge/ffmpeg-required-007808?logo=ffmpeg&logoColor=white&style=flat)
![yt-dlp](https://img.shields.io/badge/yt--dlp-required-DF0024?style=flat)

[![Build](https://img.shields.io/github/actions/workflow/status/kirbx01/discoring/go.yml?branch=main&label=build&style=flat)](https://github.com/kirbx01/discoring/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/kirbx01/discoring?style=flat)](https://github.com/kirbx01/discoring/releases)
[![Last commit](https://img.shields.io/github/last-commit/kirbx01/discoring?style=flat)](https://github.com/kirbx01/discoring)
[![Commit activity](https://img.shields.io/github/commit-activity/m/kirbx01/discoring?style=flat)](https://github.com/kirbx01/discoring)
[![Stars](https://img.shields.io/github/stars/kirbx01/discoring?style=flat)](https://github.com/kirbx01/discoring)

>[!NOTE]
> Since this is an old project almost 3 months ago it's totally buggy and has a lot to unfold since go rewrite has broken the orginal sound extraction so it really doesnt do as promised sound quality.

A free Discord music bot in Go. Streams from YouTube with **crisp, lossless-to-Opus** audio: 48 kHz stereo at 256 kbps, encoded from raw PCM (`s16le`) so there are **no grainy artifacts**, and feed to Discord is buffer-paced to stay smooth and low-latency.

> **Migrated from Rust to Go.** The original bot was built with `poise` + `serenity` + `songbird` (`Cargo.toml`, `src/`). It was rewritten in Go with `bwmarrin/discordgo`, `joho/godotenv`, and `layeh.com/gopus`.

## What changed from the Rust version

| | Old (Rust) | New (Go) |
| --- | --- | --- |
| Framework | `poise` + `serenity`, `songbird` driver | `discordgo` + own queue/player |
| Dependencies | `Cargo.toml` (tokio, reqwest, dotenv…) | `go.mod` (small, no async runtime) |
| Audio path | `songbird` `YoutubeDl` input | explicit `yt-dlp → ffmpeg(PCM) → gopus(Opus)` |
| Audio quality | grainy / artifacts on some tracks | fixed: correct 20ms Opus frames, no double-encode, no dropped packets |
| Build | `cargo build` / `cargo test` | `go build` / `go vet` (CI: `rust.yml` → `go.yml`) |
| Commands | `join`, `play`, `skip`, `leave` | identical `join`, `play`, `skip`, `leave` |

Same features: slash (`/`) and prefix (`!`) commands, per-guild queue, YouTube URLs and search terms.

## Requirements

- Go 1.22+
- `ffmpeg`
- `yt-dlp`
- A C compiler (cgo). On amd64 the Opus encoder ships bundled sources; on other architectures `libopus-dev` is needed.

## Setup

### 1. Install the audio tools

`ffmpeg` and `yt-dlp` must be on `PATH`:

```sh
sudo apt install ffmpeg
pip install --upgrade yt-dlp
```

Verify with `ffmpeg -version` and `yt-dlp --version`.

### 2. Create a bot on Discord

1. Create an application at the [Discord Developer Portal](https://discord.com/developers/applications).
2. Add a bot user and copy its token.
3. Enable the privileged gateway intents: **MESSAGE CONTENT** and **GUILD VOICE STATES**.
4. Invite it with the OAuth2 URL generator: `bot` scope with **Connect**, **Speak**, **Send Messages**, **Read Message History**, and **Use Slash Commands** permissions.

### 3. Configure

```sh
cp .env.example .env
```

Put your bot token in `.env`. Commands use the `!` prefix by default; override it with `BOT_PREFIX` (for example `BOT_PREFIX=$`) if you'd rather.

### 4. Run

```sh
go mod tidy
go run ./cmd/discoring
```

Or build a binary:

```sh
go build -o discoring ./cmd/discoring
./discoring
```

Prefix (`!`) commands work instantly. Slash commands are registered globally and can take a few minutes to an hour to appear, so use `!` while testing.

## Commands

| Command | Description |
| --- | --- |
| `/join` `!join` | Join your voice channel |
| `/play <query>` `!play <query>` | Play a YouTube URL or search term |
| `/skip` `!skip` | Skip the current track |
| `/leave` `!leave` | Leave and clear the queue |

## Audio quality

- `yt-dlp` grabs `bestaudio`, decoded straight to raw PCM by `ffmpeg` (`-vn -ac 2 -ar 48000 -f s16le`) — no lossy double-encode.
- Encoded once to Opus at 256 kbps, 20 ms frames, full-band `gopus` encoder.
- Frames are paced by Discord's own 20 ms sender, so there are no dropped or overlapping packets — no cracks, no grain, no skipping.

## Layout

```
cmd/discoring/    main.go     bot startup, intents, handler wiring
internal/bot/     commands.go slash/prefix command table + dispatch
                  actions.go  join/play/skip/leave implementations
internal/player/  player.go   per-guild queue and playback state
                  stream.go   yt-dlp → ffmpeg → Opus pipeline
```
