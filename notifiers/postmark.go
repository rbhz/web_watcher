package notifiers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rbhz/web_watcher/watcher"
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
	emails        []string
	token         string
	fromEmail     string
	subject       string
	messagePeriod time.Duration
	updates       []watcher.URLUpdate
	mux           sync.Mutex
}

// Notify via email
func (n *PostMarkNotifier) Notify(update watcher.URLUpdate) {
	if checkStatusChange(update) {
		n.mux.Lock()
		n.updates = append(n.updates, update)
		n.mux.Unlock()
	}
}

// Run sends message each n seconds
func (n *PostMarkNotifier) Run() {
	for range time.Tick(n.messagePeriod * time.Second) {
		n.mux.Lock()
		if len(n.updates) == 0 {
			n.mux.Unlock()
		} else {
			message := getMessage(n.updates)
			n.updates = make([]watcher.URLUpdate, 0)
			n.mux.Unlock()
			n.sendMessage(n.emails, message)
		}
	}
}

func (n *PostMarkNotifier) sendMessage(to []string, text string) {
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
func NewPostMarkNotifier(cfg PostMarkConfig) PostMarkNotifier {
	if len(cfg.Emails) == 0 {
		log.Fatal("Specify Postmark emails or deactivate postmark")
	}
	if cfg.APIKey == "" {
		log.Fatal("Specify Postmark token or deactivate postmark")
	}
	if cfg.FromEmail == "" {
		log.Fatal("Specify Postmark from address or deactivate postmark")
	}
	return PostMarkNotifier{
		emails:        cfg.Emails,
		token:         cfg.APIKey,
		fromEmail:     cfg.FromEmail,
		subject:       cfg.Subject,
		messagePeriod: cfg.MessagePeriod,
	}
}
