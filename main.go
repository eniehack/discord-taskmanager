package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
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

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Println("can't connect database", err)
	}
	defer db.Close()

	fields := strings.Fields(msg.Content)
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Println("failed load locaition.", err)
	}

	switch fields[0] {
	case "!add":
		task := fields[len(fields)-2]

		until, err := time.ParseInLocation("2006/01/02", fields[len(fields)-1], jst)
		if err != nil {
			log.Println("can't parse var:until.", err)
		}

		// TODO:タスクの記述方法を考える
		// TODO:SQLiteに接続してタスクの情報を記述する
		for i := 0; i < len(msg.Mentions); i++ {
			if _, err := db.Exec(
				"INSERT INTO tasks (worker, task_name, until) VALUES (?, ?, ?)",
				msg.Mentions[i].String(),
				task,
				until,
			); err != nil {
				log.Println("failed INSERT data.", err)
			}

		log.Println(
			fmt.Sprintf(
				"Called !add: %sは%sを%sまでに終わらせます",
					msg.Mentions[i].String(),
					task,
				until,
			),
		)

		s.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf(
					"%sはTaskID:%sを%sまでに終わらせます",
					msg.Mentions[i].Mention(),
					task,
				until,
			),
		)
		}

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
