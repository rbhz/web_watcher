package notifiers

import (
	"bytes"
	"html/template"

	"github.com/rbhz/http_checker/watcher"
)

func getMessage(updates []watcher.URLUpdate, messageText string) string {
	t := template.New("Message")
	t.Parse(
		messageText + ":{{ range $_, $update := $ }}\n{{$update.New.Link}} - {{$update.Error}} {{end}}",
	)
	var tpl bytes.Buffer
	t.Execute(&tpl, updates)
	return tpl.String()
}

func filterFails(updates []watcher.URLUpdate) (failed []watcher.URLUpdate) {
	for _, update := range updates {
		if update.Old.Good() && !update.New.Good() {
			failed = append(failed, update)
		}
	}
	return
}
