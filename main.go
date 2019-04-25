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
	"github.com/whiteShtef/clockwork"
)

// Token used for Command line parameters.
var Token string

type Handler struct {
	DB *sql.DB
}

func init() {
	flag.StringVar(&Token, "token", "{Some Token}", "Discord Bot Token.")
	flag.Parse()
}

func main() {

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Println("can't connect database", err)
		return
	}
	db.SetConnMaxLifetime(1)
	defer db.Close()

	log.Printf(fmt.Sprintf("Access Token: %s", Token))
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", Token))
	if err != nil {
		log.Println(err)
		return
	}

	h := &Handler{DB: db}

	discord.AddHandler(h.messageCreate)
	discord.AddHandler(h.Alerm)

	if err := discord.Open(); err != nil {
		log.Println("can't connect Discord Server.", err)
		return
	}
	defer discord.Close()

	log.Println("Bot is now running.")
	<-make(chan struct{})
}

func (h *Handler) messageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {

	if msg.Author.ID == s.State.User.ID {
		return
	}

	db := h.DB
	fields := strings.Fields(msg.Content)
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Println("failed load locaition.", err)
		return
	}

	switch fields[0] {
	case "!add":
		task := fields[len(fields)-2]

		until, err := time.ParseInLocation("2006/01/02", fields[len(fields)-1], jst)
		if err != nil {
			log.Println("can't parse var:until.", err)
			return
		}

		// TODO:タスクの記述方法を考える
		// TODO:SQLiteに接続してタスクの情報を記述する
		for i := 0; i < len(msg.Mentions); i++ {
			if _, err := db.Exec(
				"INSERT INTO tasks (worker, task_name, until) VALUES (?, ?, ?)",
				msg.Mentions[i].ID,
				task,
				until,
			); err != nil {
				log.Println("failed INSERT data.", err)
				return
			}

			var ID int

			if err := db.QueryRow(
				"SELECT rowid FROM tasks WHERE worker = ? AND task_name = ? AND until = ?",
				msg.Mentions[i].ID,
				task,
				until,
			).Scan(&ID); err != nil {
				log.Println("rowid search error.", err)
			}

			log.Println(
				fmt.Sprintf(
					"Called !add: %sは%s(TaskID:%d)を%sまでに終わらせます",
					msg.Mentions[i].String(),
					task,
					ID,
					until,
				),
			)

			s.ChannelMessageSend(
				msg.ChannelID,
				fmt.Sprintf(
					"%sは%s(TaskID:%d)を%sまでに終わらせます",
					msg.Mentions[i].Mention(),
					task,
					ID,
					until,
				),
			)
		}

	case "!finish":
		// TODO:SQLiteに接続してタスクの状態を変化させる

		// ユーザーの取得
		TaskID := fields[len(fields)-1]
		var TaskName, WorkerID string

		if _, err := db.Exec(
			"UPDATE tasks SET finished_flag = 1 WHERE rowid = ?",
			TaskID,
		); err != nil {
			log.Println(err)
		}

		if err := db.QueryRow(
			"SELECT task_name, worker FROM tasks WHERE rowid = ?",
			TaskID,
		).Scan(&TaskName, &WorkerID); err != nil {
			log.Println(err)
		}

		Worker, err := s.User(WorkerID)
		if err != nil {

		}

		log.Println(
			fmt.Sprintf(
				"Called !finished: %sはTaskID:%sの'%s'を完了させました.現在時刻:%s",
				Worker,
				TaskID,
				TaskName,
				time.Now().Format("2006/01/02 Mon 15:04:05 MST"),
			),
		)
		s.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf(
				"%s はTaskID:%sの'%s'を完了させました.現在時刻:%s",
				Worker.Mention(),
				TaskID,
				TaskName,
				time.Now().Format("2006/01/02 Mon 15:04:05 MST"),
			),
		)
	case "!help":

		add := "作業を追加します.\n使い方:`!add [作業をする人(メンション付きで)] [作業内容] [期限 例:2019/04/01]`"
		finish := "自分の作業が完了した旨を報告する際に使用します. \n使い方: `!finish [タスクID(!addした際に表示されます)]`"
		move := "作業の〆切を変更します.\n使い方:`!move [タスクID(!addした際に表示されます)]`"

		EmbedMessage := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				URL:  "https://github.com/eniehack/discord-taskmanager",
				Name: "Discord Task Manager",
			},
			Color:       0x00ff00, // Green
			Description: "This is a command help.",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "!add command",
					Value:  add,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "!finish command",
					Value:  finish,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "!move command",
					Value:  move,
					Inline: true,
				},
			},
			Image:     &discordgo.MessageEmbedImage{},
			Thumbnail: &discordgo.MessageEmbedThumbnail{},
			Timestamp: time.Now().Format(time.RFC3339),
			Title:     "Help",
		}

		s.ChannelMessageSendEmbed(
			msg.ChannelID,
			EmbedMessage,
		)

	case "!move":
		var WorkerID, TaskName string
		TaskID := fields[len(fields)-2]
		Until := fields[len(fields)-1]
		// SQL
		if _, err := db.Exec(
			"UPDATE tasks SET until = ? WHERE rowid = ?",
			Until,
			TaskID,
		); err != nil {
			log.Println(err)
		}

		if err := db.QueryRow(
			"SELECT worker, task_name FROM tasks WHERE TaskID = ?",
			TaskID,
		).Scan(&WorkerID, &TaskName); err != nil {
			log.Println(err)
		}

		Worker, err := s.User(WorkerID)
		if err != nil {
			log.Println(err)
		}

		log.Println(
			"%sさんの作業 %s(taskid:%d)の〆切を%s変更します.",
			Worker.String(),
			TaskName,
			TaskID,
			Until,
		)
		s.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf(
				"%s さんの作業 %s(taskid:%d)の〆切を変更します.",
				Worker.Mention(),
				TaskName,
				TaskID,
				Until,
			),
		)
	}
}

func (h *Handler) Alerm(s *discordgo.Session, msg *discordgo.MessageCreate) {

	db := h.DB

	var (
		ID       int
		WorkerID string
		TaskName string
		Until    time.Time
	)

	schedule := clockwork.NewScheduler()
	schedule.Schedule().Every().Day().At("0:00").Do(func() {

		time.Sleep(2 * time.Minute)

		rows, err := db.Query(
			"SELECT rowid, worker, task_name, until FROM tasks WHERE finished_flag = '0'",
		)
		if err != nil {
			log.Println("Database SELECT Error.", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&ID, &WorkerID, &TaskName, &Until); err != nil {
				log.Println("Scanning Error")
			}

			Worker, err := s.User(WorkerID)
			if err != nil {
				log.Println(err)
			}

			if deadline, _ := time.ParseDuration("0s"); time.Until(Until) < deadline {
				log.Println(
					"%sさんが担当の作業%s(taskid:%d)は〆切です.",
					Worker.String(),
					TaskName,
					ID,
				)
				s.ChannelMessageSend(
					msg.ChannelID,
					fmt.Sprintf(
						"%s さんが担当の作業%s(taskid:%d)は〆切です.",
						Worker.Mention(),
						TaskName,
						ID,
					),
				)
			}
			if day, _ := time.ParseDuration("24h"); time.Until(Until) < day {
				log.Println(
					"%sさん明日24:00に〆切となる〆切の作業%s(taskid:%d)があります.",
					Worker.String(),
					TaskName,
					ID,
				)
				s.ChannelMessageSend(
					msg.ChannelID,
					fmt.Sprintf(
						"%s さんは明日24:00に〆切となる作業%s(taskid:%d)があります.",
						Worker.Mention(),
						TaskName,
						ID,
					),
				)
			}
		}
		if err = rows.Err(); err != nil {
			log.Println(err)
		}
	})

	schedule.Run()
}
