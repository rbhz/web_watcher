package notifiers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
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
	subject   string
	baseMessageNotifier
}

func (n *PostMarkNotifier) sendMessage(text string) {
	n.log(log.Info).Msg("sending message")
	data, err := json.Marshal(&postmarkRequestData{
		From:     n.fromEmail,
		To:       n.emails[0],
		CC:       strings.Join(n.emails[1:], ","),
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
func NewPostMarkNotifier(cfg PostMarkConfig) *PostMarkNotifier {
	if len(cfg.Emails) == 0 {
		log.Fatal().Msg("Specify Postmark emails or deactivate postmark")
	}
	if cfg.APIKey == "" {
		log.Fatal().Msg("Specify Postmark token or deactivate postmark")
	}
	if cfg.FromEmail == "" {
		log.Fatal().Msg("Specify Postmark from address or deactivate postmark")
	}
	notifier := PostMarkNotifier{
		emails:    cfg.Emails,
		token:     cfg.APIKey,
		fromEmail: cfg.FromEmail,
		subject:   cfg.Subject,
		baseMessageNotifier: baseMessageNotifier{
			messagePeriod: cfg.MessagePeriod,
			name:          "postmark"}}
	notifier.sendFunc = notifier.sendMessage
	return &notifier
}
