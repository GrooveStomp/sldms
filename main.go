package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	SlackUsername string
	Config        Cfg
)

type CfgAlias struct {
	Nick string
	User string
}

type Cfg struct {
	FilterText string
	MyUsername string `toml:"myself"`
	Alias      []CfgAlias
}

func main() {
	var slackKey string
	configFile := flag.String("config", "./config.toml", "which config file to load")
	flag.StringVar(&slackKey, "slack", "", "Slack API token")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: PROGRAM [OPTIONS] USER")
		fmt.Println("\tYou must have a valid .toml config file as well.")
		os.Exit(1)
	}

	SlackUsername = flag.Args()[0]

	if _, err := toml.DecodeFile(*configFile, &Config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if slackKey == "" {
		slackKey = os.Getenv("SLAEARCH_SLACK_TOKEN")
	}
	if slackKey == "" {
		fmt.Println("Usage: PROGRAM [OPTIONS] USER")
		fmt.Println("\tYou must have a valid .toml config file as well.")
		os.Exit(1)
	}
	api := slack.New(slackKey)

	myself, err := NewSlackUser(api, Config.MyUsername)
	if err != nil {
		panic(err)
	}
	myHistory, err := myself.DirectMessageHistory(api)
	if err != nil {
		panic(err)
	}
	myFilteredMsgs := FilterMsgs1On1(myHistory.Messages)

	for i := range Config.Alias {
		alias := Config.Alias[i]
		if alias.Nick == SlackUsername {
			SlackUsername = alias.User
		}
	}

	teammate, err := NewSlackUser(api, SlackUsername)
	if err != nil {
		panic(err)
	}
	memberHistory, err := teammate.DirectMessageHistory(api)
	if err != nil {
		panic(err)
	}
	filtered := FilterMsgs1On1(memberHistory.Messages)

	numPrinted := 0
	for k := range filtered {
		msg := filtered[k]

		if Contains(myFilteredMsgs, msg) {
			continue
		}

		timeStr := ""
		time, _ := MsgTime(msg)
		if err == nil {
			timeStr = time.Format("2006-01-02 15:04")
		}

		fmt.Printf("%v\n%v\n\n", timeStr, msg.Text)
		numPrinted++
	}

	if numPrinted == 0 {
		fmt.Println("No unaddressed 1:1 items.")
	}
}

func MsgTime(self slack.Message) (time.Time, error) {
	parts := strings.Split(self.Timestamp, ".")
	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Now(), err
	}
	nano, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Now(), err
	}
	t := time.Unix(seconds, nano)
	return t, nil
}

func FilterMsgs1On1(msgs []slack.Message) []slack.Message {
	var wanted []slack.Message = nil

	for i := range msgs {
		msg := msgs[i]

		if strings.Contains(msg.Text, Config.FilterText) {
			wanted = append(wanted, msg)
		}
	}

	return wanted
}

func Contains(haystack []slack.Message, needle slack.Message) bool {
	time, err := MsgTime(needle)
	if err != nil {
		return false
	}
	timeFormat := time.Format("2006-01-02 15:04")

	for i := range haystack {
		curr := haystack[i]
		if strings.Contains(curr.Text, timeFormat) &&
			strings.Contains(curr.Text, SlackUsername) {
			return true
		}
	}

	return false
}
