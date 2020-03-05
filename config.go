package main

import (
	"fmt"
	"os"

	"github.com/BoltzExchange/channel-bot/build"
	"github.com/BoltzExchange/channel-bot/discord"
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

	Interval int `short:"i" long:"interval" description:"Interval in seconds at which the channel balances should be checked"`

	Lnd     *lnd.LND         `group:"LND Options"`
	Discord *discord.Discord `group:"Discord Options"`

	SignificantChannels []*notifications.SignificantChannel `group:"Significant Channels Options"`

	Help *helpOptions `group:"Help Options"`
}

func loadConfig() *config {
	cfg := config{
		LogFile:    "./channel-bot.log",
		ConfigFile: "./channel-bot.toml",
		Interval:   60,
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
			fmt.Println(fmt.Sprintf("Could not read config file: %s", err))
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
	logger.Info("Loaded config: " + stringify(cfg))
}
