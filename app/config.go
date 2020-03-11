package main

import "github.com/rbhz/web_watcher/watcher"

// Config definition for yaml configuration
type Config struct {
	App watcher.Config
	Web struct {
		Active bool `default:"false"`
		Port   int  `default:"8080"`
	}
	PostMark struct {
		Active      bool `default:"false"`
		APIKey      string
		Emails      []string
		FromEmail   string
		Subject     string `default:"Http checker errors"`
		MessageText string `default:"Request failed for"`
		OnlyErrors  bool   `default:"false"`
	}
	Telegram struct {
		Active      bool `default:"false"`
		BotToken    string
		Users       []int64
		MessageText string `default:"Request failed for"`
		OnlyErrors  bool   `default:"false"`
	}
}
