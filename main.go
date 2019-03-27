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
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Called !add. %sは%sを%sまでに終わらせます", msg.Mentions[0].Mention(), fields[2], fields[3]))
	case "!finished":
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Called !finished. %sは%sを完了させました.現在時刻:%s", msg.Author.Mention(), fields[1], time.Now()))
	}
}
