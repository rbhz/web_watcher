package watcher

import "time"

// Config describes app config section
type Config struct {
	Period      time.Duration `default:"10"`
	ErrorPeriod time.Duration `default:"1"`
	DBPath      string `default:"./.watcher.db"`
}
