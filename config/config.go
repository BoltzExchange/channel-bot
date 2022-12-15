package config

import (
	"fmt"
	"github.com/BoltzExchange/channel-bot/build"
	"github.com/BoltzExchange/channel-bot/utils"
	"github.com/BurntSushi/toml"
	"github.com/google/logger"
	"github.com/jessevdk/go-flags"
	"os"
)

type HelpOptions struct {
	ShowHelp    bool `short:"h" long:"help" description:"Display this help message"`
	ShowVersion bool `short:"v" long:"version" description:"Display version and exit"`
}

type BaseConfig struct {
	ConfigFile string `short:"c" long:"configfile" description:"Path to configuration file"`
	LogFile    string `short:"l" long:"logfile" description:"Path to the log file"`

	Help *HelpOptions `group:"Help Options"`
}

type baseConfigInterface interface {
	GetStruct() any
	GetHelp() *HelpOptions
	GetConfigFile() string
}

func LoadConfig[T baseConfigInterface](cfg T) T {
	parser := flags.NewParser(cfg.GetStruct(), flags.IgnoreUnknown)
	_, err := parser.Parse()

	if cfg.GetHelp().ShowVersion {
		fmt.Println(build.GetVersion())
		os.Exit(0)
	}

	if cfg.GetHelp().ShowHelp {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if err != nil {
		printCouldNotParseCli(err)
	}

	if cfg.GetConfigFile() != "" {
		_, err := toml.DecodeFile(cfg.GetConfigFile(), &cfg)

		if err != nil {
			fmt.Println("Could not read config file: " + err.Error())
		}
	}

	_, err = flags.Parse(cfg.GetStruct())

	if err != nil {
		printCouldNotParseCli(err)
	}

	return cfg
}

func printCouldNotParseCli(err error) {
	utils.PrintFatal("Could not parse CLI arguments: %s", err)
}

func LogConfig(cfg any) {
	logger.Info("Loaded config: " + utils.Stringify(cfg))
}
