package main

import (
	"fmt"
	"github.com/nlopes/slack"
)

type SlackUser struct {
	slack.User
}

func NewSlackUser(api *slack.Client, name string) (*SlackUser, error) {
	users, err := api.GetUsers()
	if err != nil {
		return nil, err
	}

	for i := range users {
		user := users[i]
		if name == user.Name {
			return &SlackUser{User: user}, nil
		}
	}

	return nil, fmt.Errorf("User not found")
}

func (self *SlackUser) DirectMessageHistory(api *slack.Client) (*slack.History, error) {
	ims, err := api.GetIMChannels()
	if err != nil {
		return nil, err
	}

	historyParams := slack.HistoryParameters{
		Count:   1000,
		Unreads: true,
	}

	for i := range ims {
		im := ims[i]
		if im.User == self.ID {
			return api.GetIMHistory(im.ID, historyParams)
		}
	}

	return nil, fmt.Errorf("History not found for user")
}
