package notifiers

import "time"

// WebConfig describes web notifier confing
type WebConfig struct {
	Active   bool `default:"false"`
	Port     int  `default:"8080"`
	Profiler bool `default:"false"`
}

// PostMarkConfig describes postmark notifier config
type PostMarkConfig struct {
	Active        bool `default:"false"`
	APIKey        string
	Emails        []string
	FromEmail     string
	Subject       string        `default:"Http checker errors"`
	MessageText   string        `default:"Request failed for"`
	MessagePeriod time.Duration `default:"10"`
}

// TelegramConfig describes telegram notifier config
type TelegramConfig struct {
	Active        bool `default:"false"`
	BotToken      string
	Users         []int64
	MessageText   string        `default:"Request failed for"`
	MessagePeriod time.Duration `default:"10"`
}

// SlackConfig describes slack notifier configuration
type SlackConfig struct {
	Active        bool `default:"false"`
	WebHookURL    string
	MessagePeriod time.Duration `default:"10"`
}
