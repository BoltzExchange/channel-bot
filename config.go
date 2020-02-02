package main

import (
	"encoding/json"
	"fmt"
	"github.com/BoltzExchange/channel-bot/build"
	"github.com/google/logger"
	"github.com/jessevdk/go-flags"
	"os"
)

type lndConfig struct {
	Host        string `long:"host" description:"gRPC host of the LND node"`
	Port        int    `long:"port" description:"gRPC port of the LND node"`
	Macaroon    string `long:"macaron" description:"Path to a macaroon file of the LND node"`
	Certificate string `long:"certificate" description:"Path to a certificate file of the LND node"`
}

type helpOptions struct {
	ShowHelp    bool `short:"h" long:"help" description:"Display this help message"`
	ShowVersion bool `short:"v" long:"version" description:"Display version and exit"`
}

type config struct {
	ConfigFile string `short:"c" long:"configfile" description:"Path to configuration file"`
	LogFile    string `short:"l" long:"logfile" description:"Path to the log file"`

	Lnd *lndConfig `group:"LND Options"`

	Help *helpOptions `group:"Help Options"`
}

func loadConfig() *config {
	cfg := config{
		LogFile: "./channel-bot.log",
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
		printFatal("Could not prase CLI arguments: %s\n", err)
	}

	if cfg.ConfigFile != "" {
		err = flags.IniParse(cfg.ConfigFile, &cfg)

		if err != nil {
			printFatal("Could not read config file: %s\n", err)
		}
	}

	return &cfg
}

func logConfig(cfg *config) {
	configJSON, _ := json.MarshalIndent(cfg, "", "  ")

	logger.Info("Config: " + string(configJSON))
}
