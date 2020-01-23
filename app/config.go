package main

// Config definition for yaml configuration
type Config struct {
	Period int    `default:"10"`
	DBPath string `default:"./.watcher.db"`
	Web    struct {
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
	}
}
