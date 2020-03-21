package notifiers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"

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

func (n *PostMarkNotifier) log(level func() *zerolog.Event) *zerolog.Event {
	return level().Str("notifier", "postmark")
}

// Run sends message each n seconds
func (n *PostMarkNotifier) Run() {
	n.log(log.Info).Msg("Postmark notifier started")
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
			n.sendMessage(n.emails, message)
		}
	}
}

func (n *PostMarkNotifier) sendMessage(to []string, text string) {
	n.log(log.Info).Msg("sending message")
	data, err := json.Marshal(&postmarkRequestData{
		From:     n.fromEmail,
		To:       to[0],
		CC:       strings.Join(to[1:], ","),
		Subject:  n.subject,
		TextBody: text,
	})
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to endode data")
		return
	}
	req, err := http.NewRequest(
		"POST", postMarkAPIURL, bytes.NewReader(data),
	)
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to create request obj")
		return
	}
	req.Header.Add("X-Postmark-Server-Token", n.token)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to perform request")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		n.log(log.Error).Int("code", resp.StatusCode).Msg("Postmark returned invalid code")
	}
}

// NewPostMarkNotifier creates notifier
func NewPostMarkNotifier(cfg PostMarkConfig) PostMarkNotifier {
	if len(cfg.Emails) == 0 {
		log.Fatal().Msg("Specify Postmark emails or deactivate postmark")
	}
	if cfg.APIKey == "" {
		log.Fatal().Msg("Specify Postmark token or deactivate postmark")
	}
	if cfg.FromEmail == "" {
		log.Fatal().Msg("Specify Postmark from address or deactivate postmark")
	}
	return PostMarkNotifier{
		emails:        cfg.Emails,
		token:         cfg.APIKey,
		fromEmail:     cfg.FromEmail,
		subject:       cfg.Subject,
		messagePeriod: cfg.MessagePeriod,
	}
}
