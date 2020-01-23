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
func (n WebNotifier) Notify(updated []watcher.URL) {
	if len(updated) > 0 {
		data, _ := json.Marshal(updated)
		n.Server.Broadcast(data)
	}
}
