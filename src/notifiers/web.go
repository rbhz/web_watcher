package notifiers

import (
	"encoding/json"
	"watcher"
	"web"
)

// WebNotifier Send notifications for web users
type WebNotifier struct {
	Server web.Server
}

// Notify web users
func (n WebNotifier) Notify(updated []watcher.URL) {
	data, _ := json.Marshal(updated)
	n.Server.Broadcast(data)
}
