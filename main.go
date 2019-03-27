package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	BotName = "560490545605246997"
	Token   = "NTYwNDkwNTQ1NjA1MjQ2OTk3.D30uhg.1OSv4OCu2ajJcQCA-iRDFu1qmt4"
)

func main() {
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", Token))
	if err != nil {
		log.Println(err)
		return
	}

	discord.AddHandler(messageCreate)

	if err := discord.Open(); err != nil {
		log.Println(err)
		return
	}
	defer discord.Close()

	fmt.Println("Bot is now running.")
	<-make(chan struct{})
}

func messageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {

	if msg.Author.ID == s.State.User.ID {
		return
	}

	fields := strings.Fields(msg.Content)

	switch fields[0] {
	case "!add":
		for i := 0; i < len(msg.Mentions); i++ {
			fmt.Println(msg.Mentions[i])
		}
		s.ChannelMessageSend(msg.ChannelID, "Called !add. "+msg.Mentions[0].Mention()+"は"+fields[2]+"を"+fields[3]+"までに終わらせます")
	case "!finished":
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Called !finished. "+msg.Author.Mention()+"は"+fields[1]+"を完了させました.現在時刻:%s", time.Now()))
	}
}
