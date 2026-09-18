package main

import (
	"log"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"discoring/internal/bot"
)

func main() {
	godotenv.Load()
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN not set")
	}

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	s.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	s.LogLevel = discordgo.LogWarning
	if lvl := os.Getenv("DISCORD_LOG"); lvl != "" {
		if n, err := strconv.Atoi(lvl); err == nil {
			s.LogLevel = n
		}
	}

	s.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		bot.RegisterCommands(s)
		log.Println("discoring online, audio is crisp")
	})
	s.AddHandler(bot.OnInteraction)
	s.AddHandler(bot.OnMessage)

	if err := s.Open(); err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	s.UpdateGameStatus(0, "!play | /play")
	select {}
}