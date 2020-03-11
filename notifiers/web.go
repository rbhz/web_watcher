package notifiers

import (
	"encoding/json"

	"github.com/rbhz/web_watcher/watcher"
	"github.com/rbhz/web_watcher/web"
)

// WebNotifier Send notifications for web users
type WebNotifier struct {
	Server web.Server
}

// Notify web users
func (n WebNotifier) Notify(update watcher.URLUpdate) {
	data, err := json.Marshal(update.New)
	if err != nil {
		return
	}
	n.Server.Broadcast(data)
}

// NewWebNotifier initialize web notifier instance
func NewWebNotifier(cfg WebConfig, watcher watcher.Watcher) WebNotifier {
	return WebNotifier{
		Server: web.GetServer(watcher, cfg.Port),
	}
}
