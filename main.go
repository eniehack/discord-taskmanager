package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Token used for Command line parameters.
var Token string

func init() {
	flag.StringVar(&Token, "token", "{Some Token}", "Discord Bot Token.")
	flag.Parse()
}

func main() {
	log.Printf(Token)
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

	log.Println("Bot is now running.")
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
			log.Println(msg.Mentions[i])
		}
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Called !add. %sは%sを%sまでに終わらせます", msg.Mentions[0].Mention(), fields[2], fields[3]))
	case "!finished":
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Called !finished. %sは%sを完了させました.現在時刻:%s", msg.Author.Mention(), fields[1], time.Now()))
	}
}
