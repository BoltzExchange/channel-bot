package main

import (
	"fmt"
	"github.com/BoltzExchange/channel-bot/cleaner"
	"github.com/BoltzExchange/channel-bot/notifications/providers/discord"
	"github.com/BoltzExchange/channel-bot/notifications/providers/mattermost"
	"github.com/BoltzExchange/channel-bot/utils"
	"os"

	"github.com/BoltzExchange/channel-bot/build"
	"github.com/BoltzExchange/channel-bot/lnd"
	"github.com/BoltzExchange/channel-bot/notifications"
	"github.com/BurntSushi/toml"
	"github.com/google/logger"
	"github.com/jessevdk/go-flags"
)

type helpOptions struct {
	ShowHelp    bool `short:"h" long:"help" description:"Display this help message"`
	ShowVersion bool `short:"v" long:"version" description:"Display version and exit"`
}

type config struct {
	ConfigFile string `short:"c" long:"configfile" description:"Path to configuration file"`
	LogFile    string `short:"l" long:"logfile" description:"Path to the log file"`

	Notifications  *notifications.ChannelManager `group:"Notification Options"`
	ChannelCleaner *cleaner.ChannelCleaner       `group:"Channel Cleaner Options"`

	Lnd        *lnd.LND               `group:"LND Options"`
	Discord    *discord.Discord       `group:"Discord Options"`
	Mattermost *mattermost.Mattermost `group:"Mattermost Options"`

	Help *helpOptions `group:"Help Options"`

	// This option is only parsed in the TOML config file
	SignificantChannels []*notifications.SignificantChannel
}

func loadConfig() *config {
	cfg := config{
		LogFile:    "./channel-bot.log",
		ConfigFile: "./channel-bot.toml",

		Notifications: &notifications.ChannelManager{},

		ChannelCleaner: &cleaner.ChannelCleaner{
			Interval:           24,
			MaxInactive:        30,
			MaxInactivePrivate: 60,
		},
	}

	parser := flags.NewParser(&cfg, flags.IgnoreUnknown)
	_, err := parser.Parse()

	if cfg.Help.ShowVersion {
		fmt.Println(build.GetVersion())
		os.Exit(0)
	}

	if cfg.Help.ShowHelp {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if err != nil {
		printCouldNotParseCli(err)
	}

	if cfg.ConfigFile != "" {
		_, err := toml.DecodeFile(cfg.ConfigFile, &cfg)

		if err != nil {
			fmt.Printf("Could not read config file: " + err.Error())
		}
	}

	_, err = flags.Parse(&cfg)

	if err != nil {
		printCouldNotParseCli(err)
	}

	return &cfg
}

func printCouldNotParseCli(err error) {
	printFatal("Could not parse CLI arguments: %s", err)
}

func logConfig(cfg *config) {
	logger.Info("Loaded config: " + utils.Stringify(cfg))
}
