package mattermost

import (
	"context"
	"errors"
	"fmt"
	"github.com/BoltzExchange/channel-bot/notifications/providers"
	"github.com/google/logger"
	"github.com/mattermost/mattermost/server/public/model"
)

type Mattermost struct {
	Url     string `long:"mattermost.url" description:"Mattermost server URL"`
	Token   string `long:"mattermost.token" description:"Mattermost authentication token"`
	Channel string `long:"mattermost.channel" description:"Name of the channel to which messages should be sent"`
	Prefix  string `long:"mattermost.prefix" description:"Prefix for every message"`

	channelId string
	client    *model.Client4
}

func (m *Mattermost) Name() string {
	return "Mattermost"
}

func (m *Mattermost) Init() error {
	client := model.NewAPIv4Client(m.Url)
	client.SetToken(m.Token)

	m.client = client

	me, _, err := client.GetMe(context.Background(), "")
	if err != nil {
		return err
	}

	teams, _, err := client.GetTeamsForUser(context.Background(), me.Id, "")
	if err != nil {
		return err
	}

	for _, team := range teams {
		channels, _, err := client.GetChannelsForTeamForUser(context.Background(), team.Id, me.Id, false, "")
		if err != nil {
			return err
		}

		for _, channel := range channels {
			if channel.Name == m.Channel || channel.DisplayName == m.Channel {
				m.channelId = channel.Id
				return nil
			}
		}
	}

	return errors.New("could not find channel")
}

func (m *Mattermost) SendMessage(message string) error {
	if m.channelId == "" || m.client == nil {
		return nil
	}

	_, _, err := m.client.CreatePost(context.Background(), &model.Post{
		ChannelId: m.channelId,
		Message:   providers.AddPrefix(m.Prefix, message),
	})

	if err != nil {
		logger.Warning("Could not send " + message + " to Mattermost: " + fmt.Sprint(err))
	}

	return err
}
