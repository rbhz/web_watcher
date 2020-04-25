package notifiers

import (
	"sync"
	"time"

	"github.com/rbhz/web_watcher/watcher"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type baseMessageNotifier struct {
	name          string
	messagePeriod time.Duration
	updates       []watcher.URLUpdate
	mux           sync.Mutex
	sendFunc      func(string)
}

func (n *baseMessageNotifier) log(level func() *zerolog.Event) *zerolog.Event {
	return level().Str("notifier", n.name)
}

func (n *baseMessageNotifier) Notify(update watcher.URLUpdate) {
	if checkStatusChange(update) {
		n.mux.Lock()
		n.updates = append(n.updates, update)
		n.mux.Unlock()
	}
}

// Run sends message each n seconds
func (n *baseMessageNotifier) Run() {
	n.log(log.Info).Msg("Notifier started")
	for range time.Tick(n.messagePeriod * time.Second) {
		n.log(log.Debug).Msg("Checking updates")
		n.mux.Lock()
		if count := len(n.updates); count == 0 {
			n.log(log.Debug).Msg("Updates not found")
			n.mux.Unlock()
		} else {
			n.log(log.Debug).Int("count", count).Msg("Sending updates")
			message := getMessage(n.updates)
			n.updates = make([]watcher.URLUpdate, 0)
			n.mux.Unlock()
			n.sendFunc(message)
		}
	}
}
