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
	log.Printf(fmt.Sprintf("Access Token: %s", Token))
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", Token))
	if err != nil {
		log.Println(err)
		return
	}

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Println("can't connect database", err)
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
		defer db.Close()

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
				return
			}

			var ID int

			if err := db.QueryRow(
				"SELECT rowid FROM tasks WHERE worker = ? AND task_name = ? AND until = ?",
				msg.Mentions[i].String(),
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
	case "!help":

		add := "!add: 作業を追加します。\n 使い方:!add [作業をする人(メンション付きで)] [作業内容] [期限 例:2019/04/01]"
		finish := "!finish: 自分の作業が完了した旨を報告する際に使用します. \n 使い方: !finish [タスクID(!addした際に表示されます)]"

		s.ChannelMessageSend(
			msg.ChannelID,
			fmt.Sprintf(
				"\n%s\n%s",
				add,
				finish,
			),
		)
	}

}

func (h *Handler) Alerm(s *discordgo.Session, msg *discordgo.MessageCreate) {

	db := h.DB

	var (
		ID       int
		Worker   string
		TaskName string
		Until    time.Time
	)

	schedule := clockwork.NewScheduler()
	schedule.Schedule().Every().Day().At("0:00").Do(func() {
		rows, err := db.Query(
			"SELECT rowid, worker, task_name, until FROM tasks WHERE finished_flag = '0'",
		)
		if err != nil {
			log.Println("Database SELECT Error.", err)
			return
		}
		defer db.Close()
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&ID, &Worker, &TaskName, &Until); err != nil {
				log.Println("Scanning Error")
			}
			if deadline, _ := time.ParseDuration("0s"); time.Until(Until) < deadline {
				log.Println(
					"%sさんが担当の作業%s(taskid:%d)は〆切です.",
					Worker,
					TaskName,
					ID,
				)
				s.ChannelMessageSend(
					msg.ChannelID,
					fmt.Sprintf(
						"%s %sさんが担当の作業%s(taskid:%d)は〆切です.",
						"@everyone",
						Worker,
						TaskName,
						ID,
					),
				)
			}
			if day, _ := time.ParseDuration("24h"); time.Until(Until) < day {
				// TODO:Workerを#で分割し、前をUsername、後をDiscriminatorとしてWorkerTypeに挿入
				// TODO:UsernameとDiscriminatorからUserIDを求める方法を考える(例:直接APIにアクセスする)か、
				// UserIDをSQLに格納する OR
				// everyoneでメンションする
				log.Println(
					"%sさん明日24:00に〆切となる〆切の作業%s(taskid:%d)があります.",
					Worker,
					TaskName,
					ID,
				)
				s.ChannelMessageSend(
					msg.ChannelID,
					fmt.Sprintf(
						"%s %sさんは明日24:00に〆切となる作業%s(taskid:%d)があります.",
						"@everyone",
						Worker,
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
	/*
		schedule.Schedule().Every().Day().At("7:00").Do(func() {
			rows, err := db.Query(
				"SELECT rowid, worker, task_name, until FROM tasks WHERE finished_flag = '0'",
			)
			if err != nil {
				log.Println("Database SELECT Error.")
				return
			}
			defer rows.Close()

			for rows.Next() {
				if err := rows.Scan(&ID, &Worker, &TaskName, &Until); err != nil {
					log.Println("Scanning Error")
				}
				if deadline, _ := time.ParseDuration("24h"); time.Until(Until) < deadline {
					log.Println(
						"%sさんが担当の作業%s(taskid:%d)は今夜24:00に〆切です.",
						Worker,
						TaskName,
						ID,
					)
					s.ChannelMessageSend(
						msg.ChannelID,
						fmt.Sprintf(
							"%sさんが担当の作業%s(taskid:%d)は今夜24:00に〆切です.",
							Worker,
							TaskName,
							ID,
						),
					)
				}
			}
		})
	*/

	schedule.Run()
}
