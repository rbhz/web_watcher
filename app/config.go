package main

import (
	"github.com/rbhz/web_watcher/notifiers"
	"github.com/rbhz/web_watcher/watcher"
)

// Config definition for yaml configuration
type Config struct {
	App      watcher.Config
	Web      notifiers.WebConfig
	PostMark notifiers.PostMarkConfig
	Telegram notifiers.TelegramConfig
	Slack    notifiers.SlackConfig
}
