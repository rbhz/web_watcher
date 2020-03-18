package notifiers

import (
	"bytes"
	"fmt"
	"time"

	"github.com/rbhz/web_watcher/watcher"
)

func getMessage(updates []watcher.URLUpdate) string {
	var message bytes.Buffer
	for idx, update := range updates {
		if idx != 0 {
			message.WriteRune('\n')
		}
		message.WriteString(fmt.Sprintf("%v %v: ",
			update.Created.Round(time.Second).UTC(), update.Old.Link))
		if errText := update.Error(); errText != nil {
			message.WriteString(*errText)
		} else {
			message.WriteString("OK")
		}
	}
	return message.String()
}

func checkStatusChange(update watcher.URLUpdate) bool {
	if update.Old.Good() != update.New.Good() {
		// status changed
		return true
	}
	if !update.New.Good() &&
		(update.New.Status != update.Old.Status ||
			update.New.Err != update.Old.Err) {
		// error changed
		return true
	}
	return false
}
