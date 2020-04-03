# Boltz Channel Bot ![GitHub Action](https://github.com/BoltzExchange/channel-bot/workflows/CI/badge.svg)

This is a  bot that sends notifications for imbalanced and (force) closed lightning channels to a Discord channel. It also incorporates features for channel management like force closing channels that have been inactive for longer that a configured amount of time.

Only LND version `0.9.0-beta` or higher is supported. Also Discord is currently the only service to which the bot can send its notifications.

## Features

The features set of the bot are subject to change and are updated as needed for managing our LND nodes.

### Notifications

The main feature of this bot is its notification service. If either less than *30%* or more than *70%* of the capacity of a channel is on the side of the LND that it is connected to, the bot will send a notification. Once the channel is balanced again according to the said requirements, the bot will also send a notification. The bot doesn't send these balance notifications for private channels unless the channel is configured as [significant channel](#significant-channels).

If a channel is closed the bot will also send a notification. Force closed channels have a special notification to indicate that something went wrong and your node needs your attention.

All of these notifications contain the channel ID and, depending on the type of notification, other relevant information. The interval at which the channels should be checked [can be configured](#sample-config).

#### Significant channels

Channels of special significance can be set in the [config file](#sample-config). Those significant channels have an alias to be able to easily identify them in the notifications, custom ratios that make them considered imbalanced and their notifications stick out from all the normal channels.

## Channel management

More to come...

### Channel cleaner

The channel cleaner takes care of force closing [zombie channels](https://medium.com/@gcomxx/get-rid-of-those-zombie-channels-1267d5a2a708). The interval at which the channels should be checked for zombies and the number of days that are needed for a channel to become a zombie are [configurable](#configuration).

## Installation

```shell
go get -d github.com/BoltzExchange/channel-bot

cd $GOPATH/src/github.com/BoltzExchange/channel-bot

make build && make install
```

You can confirm successful installation by running the channel-bot command with the --help argument.

```shell
channel-bot --help
```

## Configuration

The bot can be configured either by CLI arguments, or with a TOML config file.

### CLI Arguments

```shell
channel-bot --help
```

### Config

[`example/channel-bot.toml`](./example/channel-bot.toml)

## Disclaimer

Use at your own discretion. 