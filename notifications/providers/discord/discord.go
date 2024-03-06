package discord

import (
	"errors"
	"fmt"
	"github.com/BoltzExchange/channel-bot/notifications/providers"

	"github.com/bwmarrin/discordgo"
	"github.com/google/logger"
)

type Discord struct {
	Token   string `long:"discord.token" description:"Discord authentication token"`
	Channel string `long:"discord.channel" description:"Name of the channel to which messages should be sent"`
	Prefix  string `long:"discord.prefix" description:"Prefix for every message"`

	api       *discordgo.Session
	channelID string
}

func (d *Discord) getChannelId() error {
	guilds, err := d.api.UserGuilds(1, "", "")

	if err != nil {
		return err
	}

	guild, err := d.api.Guild(guilds[0].ID)

	if err != nil {
		return err
	}

	channels, err := d.api.GuildChannels(guild.ID)

	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name == d.Channel {
			d.channelID = channel.ID
			return nil
		}
	}

	return errors.New("could not find channel with name: " + d.Channel)
}

func (d *Discord) Name() string {
	return "Discord"
}

func (d *Discord) Init() (err error) {
	if d.Token == "" {
		return errors.New("no token configured")
	}

	d.api, err = discordgo.New("Bot " + d.Token)

	if err != nil {
		return err
	}

	err = d.api.Open()

	if err != nil {
		return err
	}

	err = d.getChannelId()
	if err != nil {
		return err
	}

	logger.Info("Initialized Discord client")
	return nil
}

func (d *Discord) SendMessage(message string) error {
	if d.api == nil {
		return nil
	}

	_, err := d.api.ChannelMessageSend(d.channelID, providers.AddPrefix(d.Prefix, message))

	if err != nil {
		logger.Warning("Could not send " + message + " to Discord: " + fmt.Sprint(err))
	}

	return err
}
