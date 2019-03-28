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
	log.Printf(fmt.Sprintf("Access Token: %s", Token))
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", Token))
	if err != nil {
		log.Println(err)
		return
	}

	discord.AddHandler(messageCreate)

	if err := discord.Open(); err != nil {
		log.Println("can't connect Discord Server.", err)
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
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Println("failed load locaition.", err)
	}

	switch fields[0] {
	case "!add":
		work := fields[len(fields)-2]

		until, err := time.ParseInLocation("2006/01/02", fields[len(fields)-1], jst)
		if err != nil {
			log.Println("can't parse var:until.", err)
		}

		// TODO:タスクの記述方法を考える
		// TODO:SQLiteに接続してタスクの情報を記述する
		log.Println(
			fmt.Sprintf(
				"Called !add: %sは%sを%sまでに終わらせます",
				msg.Mentions[0].String(),
				work,
				until,
			),
		)

		s.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf(
				"%sは%sを%sまでに終わらせます",
				msg.Mentions[0].Mention(),
				work,
				until,
			),
		)
	case "!finished":
		// TODO:SQLiteに接続してタスクの状態を変化させる
		log.Println(
			fmt.Sprintf(
				"Called !finished: %sは%sを完了させました.現在時刻:%s",
				msg.Author.String(),
				fields[1],
				time.Now().Format("2006/01/02 Mon 15:04:05 MST"),
			),
		)
		s.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf(
				"Called !finished: %sは%sを完了させました.現在時刻:%s",
				msg.Author.Mention(),
				fields[1],
				time.Now().Format("2006/01/02 Mon 15:04:05 MST"),
			),
		)
	}
}
