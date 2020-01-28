package notifiers

import (
	"bytes"
	"encoding/json"
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
	emails      []string
	token       string
	fromEmail   string
	subject     string
	messageText string
}

// Notify via email
func (n PostMarkNotifier) Notify(updates []watcher.URLUpdate) {
	updates = filterFails(updates)
	if len(updates) > 0 {
		n.sendMessage(
			n.emails,
			getMessage(updates, n.messageText),
		)
	}
}

func (n PostMarkNotifier) sendMessage(to []string, text string) {
	data, err := json.Marshal(&postmarkRequestData{
		From:     n.fromEmail,
		To:       to[0],
		CC:       strings.Join(to[1:], ","),
		Subject:  n.subject,
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
	if resp.StatusCode != http.StatusOK {
		log.Printf("Postmark returned invalid code: %v", resp.StatusCode)
	}
}

// NewPostMarkNotifier creates notifier
func NewPostMarkNotifier(emails []string, token, from, subject, messageText string) PostMarkNotifier {
	if len(emails) == 0 {
		log.Fatal("Specify Postmark emails or deactivate postmark")
	}
	if token == "" {
		log.Fatal("Specify Postmark token or deactivate postmark")
	}
	if from == "" {
		log.Fatal("Specify Postmark from address or deactivate postmark")
	}
	return PostMarkNotifier{
		emails:      emails,
		token:       token,
		fromEmail:   from,
		subject:     subject,
		messageText: messageText,
	}
}
