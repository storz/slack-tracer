package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/nlopes/slack"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const defaultCfgFilePath = "config.tml"
const defaultHistoryLength = 1000

type config struct {
	Token  string `toml:"token"`
	Length int    `toml:"length"`
}

type msg struct {
	text      string
	user      string
	channel   string
	timestamp string
}

var (
	slackClient   *slack.Client
	token         string
	history       []msg
	historyLength int
)

func getConfig(path string) (config, error) {
	var cfg config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func getUserName(uid string) string {
	userInfo, _ := slackClient.GetUserInfo(uid)
	return userInfo.Name
}

func hasDiff(diffs []diffmatchpatch.Diff) bool {
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffDelete || diff.Type == diffmatchpatch.DiffInsert {
			return true
		}
	}
	return false
}

func msgChangedHandler(text string, channel string, ts string) bool {
	for i, msg := range history {
		if msg.channel == channel && msg.timestamp == ts {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(msg.text, text, false)
			if hasDiff(diffs) {
				log.Printf("[Edited] <%s> %s", getUserName(msg.user), dmp.DiffPrettyText(diffs))
				history[i].text = text
				return true
			}
		}
	}

	return false
}

func msgDeletedHandler(channel string, ts string) bool {
	for i, msg := range history {
		if msg.channel == channel && msg.timestamp == ts {
			log.Printf("[Deleted] <%s> %s\n", getUserName(msg.user), color.RedString(msg.text))
			history = append(history[:i], history[i+1:]...)
			return true
		}
	}

	return false
}

func addHistory(row msg) {
	if len(history) < historyLength {
		history = append(history, row)
	} else {
		history = append(history[1:historyLength], row)
	}
}

func listen() {
	rtm := slackClient.NewRTM()
	go rtm.ManageConnection()
	log.Println(color.CyanString("[Ready]"))

Loop:
	for {
		select {
		case ie := <-rtm.IncomingEvents:
			switch ev := ie.Data.(type) {
			case *slack.MessageEvent:
				sm := ev.Msg
				ssm := ev.SubMessage
				switch ev.SubType {
				case "message_changed":
					found := msgChangedHandler(ssm.Text, sm.Channel, ssm.Timestamp)
					if !found {
						addHistory(msg{ssm.Text, ssm.User, sm.Channel, ssm.Timestamp})
					}
					break
				case "message_deleted":
					msgDeletedHandler(sm.Channel, sm.DeletedTimestamp)
					break
				default:
					addHistory(msg{sm.Text, sm.User, sm.Channel, sm.Timestamp})
				}
			case *slack.RTMError:
				log.Println(color.RedString("Error: %s\n", ev.Error()))
			case *slack.InvalidAuthEvent:
				log.Fatalln(color.RedString("Invalid credentials"))
				break Loop
			}
		}
	}
}

func main() {
	cfgFile := flag.String("f", defaultCfgFilePath, "Path of toml file for configure")
	t := flag.String("t", "", "Bearer token")
	l := flag.Int("l", 0, "Length of history")
	flag.Parse()

	if *t == "" {
		cfg, err := getConfig(*cfgFile)
		if err != nil {
			log.Fatalf("Load config error: %v", err)
		}

		token = cfg.Token
	} else {
		token = *t
	}

	if *l == 0 {
		cfg, _ := getConfig(*cfgFile)

		if cfg.Length == 0 {
			historyLength = defaultHistoryLength
		} else {
			historyLength = cfg.Length
		}
	} else {
		historyLength = *l
	}

	if token == "" {
		flag.Usage()
		log.Fatalf("No credential")
	}

	slackClient = slack.New(token)

	if slackClient == nil {
		log.Fatalf("Create client error")
	}

	history = []msg{}
	listen()
}
