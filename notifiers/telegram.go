package notifiers

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rbhz/web_watcher/watcher"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TelegramNotifier Sends notifications via TelegramBotAPI
type TelegramNotifier struct {
	bot           *tgbotapi.BotAPI
	users         []int64
	messagePeriod time.Duration
	updates       []watcher.URLUpdate
	mux           sync.Mutex
}

// Notify users about update
func (n *TelegramNotifier) Notify(update watcher.URLUpdate) {
	if checkStatusChange(update) {
		n.mux.Lock()
		n.updates = append(n.updates, update)
		n.mux.Unlock()
	}
}

func (n *TelegramNotifier) log(level func() *zerolog.Event) *zerolog.Event {
	return level().Str("notifier", "telegram")
}

// Run sends message each n seconds
func (n *TelegramNotifier) Run() {
	n.log(log.Info).Msg("Telegram notifier started")
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
			wg := sync.WaitGroup{}
			for _, user := range n.users {
				wg.Add(1)
				go func(user int64) {
					defer wg.Done()
					n.sendMessage(user, message)
				}(user)
			}
			wg.Wait()
		}
	}
}

// SendMessage to telegram user
func (n *TelegramNotifier) sendMessage(user int64, message string) {
	msg := tgbotapi.NewMessage(user, message)
	_, err := n.bot.Send(msg)
	if err != nil {
		n.log(log.Error).Err(err).Msg("Failed to send telegram message")
	}
}

// NewTelegramNotifier creates notifier
func NewTelegramNotifier(cfg TelegramConfig) TelegramNotifier {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Telegram bot")
	}
	return TelegramNotifier{
		bot:           bot,
		users:         cfg.Users,
		messagePeriod: cfg.MessagePeriod,
	}
}
