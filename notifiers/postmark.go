package notifiers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rbhz/http_checker/watcher"
)

const (
	postMarkAPIURL = "https://api.postmarkapp.com/email"
)

type postmarkRequestData struct {
	From     string
	To       string
	CC       string
	Subject  string
	TextBody string
}

// PostMarkNotifier Sends notifications via PostMarkAPI
type PostMarkNotifier struct {
	emails    []string
	token     string
	fromEmail string
}

// Notify via email
func (n PostMarkNotifier) Notify(updated []watcher.URL) {
	var failed []watcher.URL
	for _, url := range updated {
		if !url.Good {
			failed = append(failed, url)
		}
	}
	if len(failed) > 0 {
		n.sendMessage(
			n.emails,
			n.getMessage(failed),
			"Http checker errors",
		)
	}
}

func (n PostMarkNotifier) getMessage(urls []watcher.URL) string {
	t := template.New("Message")
	t.Parse("Requests failed for:{{ range $_, $url := $ }}\n{{$url.Link}}{{end}}")
	var tpl bytes.Buffer
	t.Execute(&tpl, urls)
	return tpl.String()
}

func (n PostMarkNotifier) sendMessage(to []string, text string, subject string) {
	data, err := json.Marshal(&postmarkRequestData{
		From:     n.fromEmail,
		To:       to[0],
		CC:       strings.Join(to[1:], ","),
		Subject:  subject,
		TextBody: text,
	})
	if err != nil {
		log.Printf("Failed to endode data: %v", err)
		return
	}
	req, err := http.NewRequest(
		"POST", postMarkAPIURL, bytes.NewReader(data),
	)
	if err != nil {
		log.Printf("Failed to create request obj: %v", err)
		return
	}
	req.Header.Add("X-Postmark-Server-Token", n.token)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to perform request: %v", err)
		return
	}
	defer resp.Body.Close()
}

// NewPostMarkNotifier creates notifier
func NewPostMarkNotifier(emails []string, token string, from string) PostMarkNotifier {
	return PostMarkNotifier{
		emails:    emails,
		token:     token,
		fromEmail: from,
	}
}
