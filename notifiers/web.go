package notifiers

import (
	"encoding/json"

	"github.com/rbhz/http_checker/watcher"
	"github.com/rbhz/http_checker/web"
)

// WebNotifier Send notifications for web users
type WebNotifier struct {
	Server web.Server
}

// Notify web users
func (n WebNotifier) Notify(updates []watcher.URLUpdate) {
	if len(updates) > 0 {
		urls := make([]*watcher.URL, len(updates))
		for idx, update := range updates {
			urls[idx] = update.New
		}
		data, _ := json.Marshal(urls)
		n.Server.Broadcast(data)
	}
}
