package notifiers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type slackSendMessageRequest struct {
	Text string `json:"text"`
}

// SlackNotifier Sends notifications to slack channel
type SlackNotifier struct {
	baseMessageNotifier
	webHookURL string
}

func (n *SlackNotifier) sendMessage(message string) {
	n.log(log.Info).Msg("sending message")
	data, err := json.Marshal(&slackSendMessageRequest{
		Text: message,
	})
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to endode data")
		return
	}
	req, err := http.NewRequest(
		"POST", n.webHookURL, bytes.NewReader(data),
	)
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to create request obj")
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to perform request")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		n.log(log.Error).Int("code", resp.StatusCode).Msg("Slack returned invalid code")
	}
}

// NewSlackNotifier Creates new slack notifier
func NewSlackNotifier(cfg SlackConfig) *SlackNotifier {
	notifier := SlackNotifier{
		webHookURL: cfg.WebHookURL,
		baseMessageNotifier: baseMessageNotifier{
			name:          "slack",
			messagePeriod: cfg.MessagePeriod}}
	notifier.sendFunc = notifier.sendMessage
	return &notifier
}
