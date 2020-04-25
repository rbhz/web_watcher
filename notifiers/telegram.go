package notifiers

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rs/zerolog/log"
)

// TelegramNotifier Sends notifications via TelegramBotAPI
type TelegramNotifier struct {
	bot   *tgbotapi.BotAPI
	users []int64
	baseMessageNotifier
}

func (n *TelegramNotifier) sendMessage(message string) {
	n.log(log.Info).Msg("sending messages")
	wg := sync.WaitGroup{}
	for _, user := range n.users {
		wg.Add(1)
		go func(user int64) {
			defer wg.Done()
			msg := tgbotapi.NewMessage(user, message)
			_, err := n.bot.Send(msg)
			if err != nil {
				n.log(log.Error).Err(err).Msg("Failed to send telegram message")
			}
		}(user)
	}
	wg.Wait()
}

// NewTelegramNotifier creates notifier
func NewTelegramNotifier(cfg TelegramConfig) *TelegramNotifier {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Telegram bot")
	}
	notifier := TelegramNotifier{
		bot:   bot,
		users: cfg.Users,
		baseMessageNotifier: baseMessageNotifier{
			messagePeriod: cfg.MessagePeriod,
			name:          "telegram"}}
	notifier.sendFunc = notifier.sendMessage
	return &notifier
}
