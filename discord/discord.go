package discord

import (
	"errors"
	"fmt"

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

func (discord *Discord) getChannelId() error {
	guilds, err := discord.api.UserGuilds(1, "", "")

	if err != nil {
		return err
	}

	guild, err := discord.api.Guild(guilds[0].ID)

	if err != nil {
		return err
	}

	channels, err := discord.api.GuildChannels(guild.ID)

	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name == discord.Channel {
			discord.channelID = channel.ID
			return nil
		}
	}

	return errors.New("could not find channel with name: " + discord.Channel)
}

func (discord *Discord) Init() (err error) {
	discord.api, err = discordgo.New("Bot " + discord.Token)

	if err != nil {
		return err
	}

	err = discord.api.Open()

	if err != nil {
		return err
	}

	return discord.getChannelId()
}

func (discord *Discord) SendMessage(message string) error {
	if discord.Prefix != "" {
		message = discord.Prefix + ": " + message
	}

	_, err := discord.api.ChannelMessageSend(discord.channelID, message)

	if err != nil {
		logger.Warning("Could not send " + message + " to Discord: " + fmt.Sprint(err))
	}

	return err
}
